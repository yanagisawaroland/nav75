#!/usr/bin/env python3
import json, os
from http.server import HTTPServer, BaseHTTPRequestHandler

DATA_FILE = '/www/devops75/nav75/custom_services.json'

def load_data():
    if os.path.exists(DATA_FILE):
        with open(DATA_FILE) as f:
            return json.load(f)
    return {"a": [], "b": [], "c": [], "d": [], "index": []}

def save_data(data):
    with open(DATA_FILE, 'w') as f:
        json.dump(data, f, ensure_ascii=False)

class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args): pass

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        self.end_headers()

    def do_GET(self):
        env = self.path.strip('/')
        data = load_data()
        result = data.get(env, [])
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(json.dumps(result, ensure_ascii=False).encode())

    def do_POST(self):
        env = self.path.strip('/')
        length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(length)
        services = json.loads(body)
        data = load_data()
        data[env] = services
        save_data(data)
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Access-Control-Allow-Origin', '*')
        self.end_headers()
        self.wfile.write(b'{"ok":true}')

HTTPServer(('0.0.0.0', 8889), Handler).serve_forever()
