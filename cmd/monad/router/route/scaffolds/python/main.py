from http.server import HTTPServer, BaseHTTPRequestHandler
import json
from os import getenv

class RequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == "/events":
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            response = {"statusCode": 200, "body": "event received"}
            self.wfile.write(json.dumps(response).encode())

    def do_GET(self):
        path = self.path.rstrip('?').rstrip('/')
        
        if path == "/health":
            response = {"statusCode": 200, "body": "health check passed"}
        elif path == "" or path == "/":
            response = dict(self.headers)
        else:
            response = {"statusCode": 404, "body": "not found", "path": path}
            
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(response).encode())

if __name__ == "__main__":
    port = int(getenv("PORT", "8090"))
    server = HTTPServer(('0.0.0.0', port), RequestHandler)
    print(f"Server running on port {port}")
    server.serve_forever()