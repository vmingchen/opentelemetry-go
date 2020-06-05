module go.opentelemetry.io/otel/example/dynamicconfig

go 1.13

replace go.opentelemetry.io/otel => ../..

replace go.opentelemetry.io/otel/exporters/dynamicconfig => ../../exporters/dynamicconfig

replace go.opentelemetry.io/otel/exporters/otlp => ../../exporters/otlp

require (
	github.com/open-telemetry/opentelemetry-collector v0.3.0 // indirect
	go.opentelemetry.io/otel v0.6.0
	go.opentelemetry.io/otel/exporters/dynamicconfig v0.6.0
	go.opentelemetry.io/otel/exporters/otlp v0.6.0
)
