# loginsrv-grpc
A grpc wrapper lib for [loginsrv](https://github.com/tarent/loginsrv) authentication service.

## Usage
Authentication is does on both of the server and client an interceptor.
See [example](https://github.com/motia/loginsrv-grpc/blob/master/example) for more details
### server

```go
import (
  loginsrv_grpc "github.com/motia/loginsrv-grpc"
  grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

loginSrv := loginsrv_grpc.NewLoginSrvServer("http://localhost:8080")

s := grpc.NewServer(
  grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
    grpc_auth.StreamServerInterceptor(loginSrv.Authenticate),
  )),
  grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
    grpc_auth.UnaryServerInterceptor(loginSrv.Authenticate),
  )),
)
loginsrv_grpc.RegisterAuthServer(s, loginSrv)
```

> If you want to define a custom/no authentication for a grpc service in your server, define a `AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error)` for it.

### client
In principle, clients should add a metadata entry to their RPC with `authorization` as key and `bearer $JWT_TOKEN$` as a value. An interceptor is a good place to implement that.


Gophers can use the helper  `loginsrv_grpc.NewClientTokenInterceptor` to create the interceptor
```go
# for gopher clients
import (
  loginsrv_grpc "github.com/motia/loginsrv-grpc"
  grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

token := "JWT_ACCESS_TOKEN"
tokenAdderInterceptor := grpc.UnaryClientInterceptor(
  loginsrv_grpc.NewClientTokenInterceptor(func () {
    return &token
  }))

conn, err := grpc.Dial(
  address,
  grpc.WithChainUnaryInterceptor(tokenAdderInterceptor),
)
```

## Development
- Tests are executed against a docker container of `loginsrv`
```bash
# run container
docker run -p 8080:8080 tarent/loginsrv -cookie-secure=false \
    -jwt-secret my_secret -simple bob=secret -jwt-refreshes 20
# run tests
go test
```
- Use `gen.sh` to sync the generated protocol buffers.
