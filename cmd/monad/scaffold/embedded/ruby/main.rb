require 'webrick'
require 'json'

server = WEBrick::HTTPServer.new(
  Port: ENV["PORT"] || 8090,
  BindAddress: '0.0.0.0'
)

server.mount_proc '/events' do |req, res|
  res.content_type = 'application/json'
  res.body = { statusCode: 200, body: 'event received' }.to_json
end

server.mount_proc '/health' do |req, res|
  res.content_type = 'application/json'
  res.body = { statusCode: 200, body: 'health check passed' }.to_json
end

server.mount_proc '/' do |req, res|
  res.content_type = 'application/json'
  res.body = req.header.to_json
end

trap('INT') { server.shutdown }
server.start