# gRPC authentication example
The example contains two main files. One for the gRPC server process and the other is for the client. The client will login, refresh the token then load the user profile informtions.
The example requires a separte process for `loginsrv`.

## How to run it?
On separate terminals, run the docker container, server and login processes

- run container
```bash
docker run -p 8080:8080 tarent/loginsrv -cookie-secure=false \
    -jwt-secret my_secret -simple bob=secret -jwt-refreshes 20
```
- Start server
```bash
go run server/main.go
```

- Start client
```bash
go run client/main.go
```
