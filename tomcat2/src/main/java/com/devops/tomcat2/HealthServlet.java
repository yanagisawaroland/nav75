package com.devops.tomcat2;

import javax.servlet.annotation.WebServlet;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;

@WebServlet("/health")
public class HealthServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws IOException {
        resp.setContentType("application/json;charset=UTF-8");
        resp.setStatus(200);
        PrintWriter out = resp.getWriter();
        String version = System.getenv("APP_VERSION") != null ? System.getenv("APP_VERSION") : "1.0.0";
        out.print("{\"status\":\"ok\",\"app\":\"tomcat2\",\"version\":\"" + version + "\"}");
    }
}
