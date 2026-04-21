package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	db *sql.DB

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapp_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "goapp_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	dbQueryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "goapp_db_queries_total",
		Help: "Total number of database queries",
	}, []string{"operation", "status"})

	appInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "goapp_info",
		Help: "Application info",
	}, []string{"version"})
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func initDB() {
	host := getEnv("DB_HOST", "pxc-db-haproxy.mysql.svc.cluster.local")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASS", "Root@Percona2024!")
	name := getEnv("DB_NAME", "goapp")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)

	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				log.Println("Database connected successfully")
				break
			}
		}
		log.Printf("DB connection attempt %d failed: %v, retrying in 3s...", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	log.Println("Database table initialized")
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

func main() {
	version := getEnv("APP_VERSION", "unknown")
	appInfo.WithLabelValues(version).Set(1)

	initDB()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(prometheusMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": version,
			"service": "goapp",
		})
	})

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_error", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/users", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, email, created_at FROM users LIMIT 100")
		if err != nil {
			dbQueryTotal.WithLabelValues("select", "error").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
		dbQueryTotal.WithLabelValues("select", "ok").Inc()
		if users == nil {
			users = []User{}
		}
		c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
	})

	r.POST("/users", func(c *gin.Context) {
		var u User
		if err := c.ShouldBindJSON(&u); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if u.Name == "" || u.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name and email are required"})
			return
		}
		result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
		if err != nil {
			dbQueryTotal.WithLabelValues("insert", "error").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		dbQueryTotal.WithLabelValues("insert", "ok").Inc()
		c.JSON(http.StatusCreated, gin.H{"id": id, "name": u.Name, "email": u.Email})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	port := getEnv("PORT", "8080")
	log.Printf("Starting goapp v%s on port %s", version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
