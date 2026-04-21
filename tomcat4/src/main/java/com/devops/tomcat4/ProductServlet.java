package com.devops.tomcat4;

import javax.servlet.annotation.WebServlet;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;

@WebServlet("/api/products")
public class ProductServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws IOException {
        resp.setContentType("application/json;charset=UTF-8");
        resp.setStatus(200);
        PrintWriter out = resp.getWriter();
        out.print("{\"products\":[" +
            "{\"id\":1,\"name\":\"iPhone 16\",\"price\":9999.0}," +
            "{\"id\":2,\"name\":\"MacBook Pro\",\"price\":19999.0}," +
            "{\"id\":3,\"name\":\"iPad Pro\",\"price\":8999.0}" +
            "],\"count\":3}");
    }
}
