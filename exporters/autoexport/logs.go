// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package autoexport // import "go.opentelemetry.io/contrib/exporters/autoexport"

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
)

// LogOption applies an autoexport configuration option.
type LogOption = option[log.Exporter]

var logsSignal = newSignal[log.Exporter]("OTEL_LOGS_EXPORTER")

// NewLogExporter returns a configured [go.opentelemetry.io/otel/sdk/log.Exporter]
// defined using the environment variables described below.
//
// OTEL_LOGS_EXPORTER defines the logs exporter; supported values:
//   - "none" - "no operation" exporter
//   - "otlp" (default) - OTLP exporter; see [go.opentelemetry.io/otel/exporters/otlp/otlplog]
//   - "console" - Standard output exporter; see [go.opentelemetry.io/otel/exporters/stdout/stdoutlog]
//
// OTEL_EXPORTER_OTLP_PROTOCOL defines OTLP exporter's transport protocol;
// supported values:
//   - "http/protobuf" (default) -  protobuf-encoded data over HTTP connection;
//     see: [go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp]
//
// An error is returned if an environment value is set to an unhandled value.
//
// Use [RegisterLogExporter] to handle more values of OTEL_LOGS_EXPORTER.
//
// Use [WithFallbackLogExporter] option to change the returned exporter
// when OTEL_LOGS_EXPORTER is unset or empty.
//
// Use [IsNoneLogExporter] to check if the returned exporter is a "no operation" exporter.
func NewLogExporter(ctx context.Context, opts ...LogOption) (log.Exporter, error) {
	return logsSignal.create(ctx, opts...)
}

// RegisterLogExporter sets the log.Exporter factory to be used when the
// OTEL_LOGS_EXPORTER environment variable contains the exporter name.
// This will panic if name has already been registered.
func RegisterLogExporter(name string, factory func(context.Context) (log.Exporter, error)) {
	must(logsSignal.registry.store(name, factory))
}

func init() {
	RegisterLogExporter("otlp", func(ctx context.Context) (log.Exporter, error) {
		proto := os.Getenv(otelExporterOTLPProtoEnvKey)
		if proto == "" {
			proto = "http/protobuf"
		}

		switch proto {
		// grpc is not supported yet, should comment out when it is supported
		// case "grpc":
		// 	return otlploggrpc.New(ctx)
		case "http/protobuf":
			return otlploghttp.New(ctx)
		default:
			return nil, errInvalidOTLPProtocol
		}
	})
	RegisterLogExporter("console", func(ctx context.Context) (log.Exporter, error) {
		return stdoutlog.New()
	})
	RegisterLogExporter("none", func(ctx context.Context) (log.Exporter, error) {
		return noopLogExporter{}, nil
	})
}
