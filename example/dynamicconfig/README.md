# Dynamic config service example

Right now the generated protobuf .go file we're using sits locally in opentelemetry-go/exporters/dynamicconfig/v1. Later we'll, want to import it from opentelemetry-proto.

### Expected behaviour for this example
After about 10 seconds, you'll see metrics get outputted to stdout once every second.

### Run server

```sh
go run server.go
```

### Run sdk

```sh
go run main.go
```
