module go.opentelemetry.io/otel/exporters/dynamicconfig

go 1.13

replace go.opentelemetry.io/otel => ../..
replace go.opentelemetry.io/otel/exporters/otlp => ../otlp

require (
	github.com/golang/protobuf v1.4.2
	github.com/open-telemetry/opentelemetry-proto v0.3.0
	go.opentelemetry.io/otel v0.6.0
	go.opentelemetry.io/otel/exporters/otlp v0.6.0
	google.golang.org/grpc v1.29.1
)
