qm - Quick Mock
===============

`qm` is a small utility to quickly mock an HTTP resource. It will output a local
URL and continue to serve the given file for 15 minutes, refreshing this timer
on each request.


Example
-------

A code snippets says more then a thousend words:
```bash
$ echo '{"name": "John", "realdata": true}' > ./user-resource.json
$ qm ./user-resource.json
http://127.0.0.1:60466
$ curl -i "http://127.0.0.1:60466"
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Length: 31
Content-Type: application/json
Last-Modified: Tue, 15 Mar 2016 08:32:33 GMT
Date: Tue, 15 Mar 2016 08:32:48 GMT

{"name": "John", "realdata": true}
```


Installation
------------

Using [Golang](https://golang.org):
```bash
$ go get -u github.com/koenbollen/qm
```
