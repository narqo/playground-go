# Processing Multipart Mixed Request

1. Run server

```
$ go run ./ -addr=127.0.0.1:8080
```

2. Send example request

```
$ nc 127.0.0.1 8080 <example.http

HTTP/1.1 200 OK
Content-Type: multipart/mixed; boundary=5ee9621e57d3d82ffc5fcf3295e4d32aa04f9039f6c18f9b99ac20b57d1c
Date: Sat, 23 May 2020 18:18:46 GMT
Content-Length: 298
Connection: close

--5ee9621e57d3d82ffc5fcf3295e4d32aa04f9039f6c18f9b99ac20b57d1c
Content-Type: application/json

{"status": "ok"}
--5ee9621e57d3d82ffc5fcf3295e4d32aa04f9039f6c18f9b99ac20b57d1c
Content-Type: application/json

{"status": "ok"}
--5ee9621e57d3d82ffc5fcf3295e4d32aa04f9039f6c18f9b99ac20b57d1c--
```
