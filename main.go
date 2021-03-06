package main

import (
	"fmt"
	"os"

	_ "github.com/devigned/tab/opencensus"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"

	"github.com/devigned/apmz/cmd"
)

func main() {
	if os.Getenv("TRACING") == "true" {
		closer, err := initOpenCensus()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer closer()
	}

	cmd.Execute()
}

func initOpenCensus() (func(), error) {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     "localhost:6831",
		CollectorEndpoint: "http://localhost:14268/api/traces",
		Process: jaeger.Process{
			ServiceName: "pub",
		},
	})

	if err != nil {
		return nil, err
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(exporter)
	return exporter.Flush, nil
}
