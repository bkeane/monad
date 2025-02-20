const http = require('http');

const PORT = process.env.PORT || 8090;

const server = http.createServer((req, res) => {
  res.setHeader('Content-Type', 'application/json');
  
  const path = req.url.replace(/\?$/, '').replace(/\/+$/, '');
  const routeKey = `${req.method} ${path || '/'}`;
  console.log({ routeKey });

  switch (routeKey) {
    case 'GET /health':
      res.writeHead(200);
      res.end(JSON.stringify({
        statusCode: 200,
        body: "health check passed"
      }));
      break;
    case 'GET /':
      res.writeHead(200);
      res.end(JSON.stringify(req.headers));
      break;
    case 'POST /events':
      res.writeHead(200);
      res.end(JSON.stringify({
        statusCode: 200,
        body: "event received"
      }));
      break;
    default:
      res.writeHead(404);
      res.end(JSON.stringify({
        statusCode: 404,
        body: "not found"
      }));
  }
});

server.on('error', (error) => {
  console.error('Server error:', error);
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running on port ${PORT}`);
});

