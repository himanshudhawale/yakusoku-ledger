'use strict';

var fs = require('fs');
var http = require('http');
var path = require('path');

var host = process.env.HOST || '127.0.0.1';
var port = parseInt(process.env.PREVIEW_PORT || '4173', 10);
var publicDirectory = path.join(__dirname, 'public');
var contentTypes = {
	'.css': 'text/css; charset=utf-8',
	'.html': 'text/html; charset=utf-8',
	'.js': 'application/javascript; charset=utf-8',
	'.png': 'image/png',
	'.svg': 'image/svg+xml'
};

http.createServer(function(req, res) {
	var requestPath = req.url.split('?')[0];
	var relativePath = requestPath === '/' ? 'index.html' : requestPath.replace(/^\/+/, '');
	var filePath = path.resolve(publicDirectory, relativePath);

	if (filePath.indexOf(publicDirectory + path.sep) !== 0) {
		res.writeHead(403);
		return res.end('Forbidden');
	}

	fs.readFile(filePath, function(err, content) {
		if (err) {
			res.writeHead(err.code === 'ENOENT' ? 404 : 500);
			return res.end(err.code === 'ENOENT' ? 'Not found' : 'Unable to read file');
		}
		res.writeHead(200, {
			'Content-Type': contentTypes[path.extname(filePath)] || 'application/octet-stream',
			'Content-Security-Policy': "default-src 'self'; connect-src 'self'; img-src 'self' data:; style-src 'self'",
			'Referrer-Policy': 'no-referrer',
			'X-Content-Type-Options': 'nosniff',
			'X-Frame-Options': 'DENY'
		});
		return res.end(content);
	});
}).listen(port, host, function() {
	console.log('Yakusoku Ledger preview at http://' + host + ':' + port);
});
