# Processing Multipart Mixed Request

1. Run server

```
$ go run ./ -addr=127.0.0.1:8080
```

2. Send example request

```
cat example.http | nc 127.0.0.1 8080
```