# registry redis for Kitex 

Redis as service discovery for Kitex.

## How to use?

### Server

```go
import (
    // ...
    registry "github.com/ydmxcz/kitex-registry-redis"
    "github.com/cloudwego/kitex/pkg/rpcinfo"
	
    // ...
)

func main() {
    // ... 
    r, err := registry.NewRedisRegistry("127.0.0.1:6379")
    if err != nil {
        panic(err)
    }
    svr := echo.NewServer(
        new(EchoImpl), 
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "echo"}),
        server.WithRegistry(r), 
	)
    if err := svr.Run(); err != nil {
        log.Println("server stopped with error:", err)
    } else {
        log.Println("server stopped")
    }
    // ...
}
```

### Client

```go
import (
    // ...
    resolver "github.com/ydmxcz/kitex-registry-redis"
    // ...
)

func main() {
    // ... 
    r, err := resolver.NewRedisResolver()
	if err != nil {
	    panic(err)	
    }
    client, err := echo.NewClient("echo", client.WithResolver(r))
    if err != nil {
        log.Fatal(err)
    }
    // ...
}
```