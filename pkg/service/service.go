package service

import (
	"github.com/devigned/apmz-sdk/apmz"

	"github.com/devigned/apmz/pkg/format"
)

type (
	// Registry holds the factories and services needed for command execution
	Registry struct {
		APMerFactory   func() (APMer, error)
		PrinterFactory func() format.Printer
	}

	// CommandServicer provides all functionality needed for command execution
	CommandServicer interface {
		GetAPMer() (APMer, error)
		GetPrinter() format.Printer
	}

	//// Closer provides the ability to close the client
	//Closer interface {
	//	Close(retryTimeout ...time.Duration) <-chan struct{}
	//}

	// APMer provides the behaviors needed to send events to Azure Application Insights
	APMer interface {
		Track(telemetry apmz.Telemetry)
		Channel() apmz.TelemetryChannel
	}
)

// GetAPMer returns an instance of an Azure Application Insights client
func (r *Registry) GetAPMer() (APMer, error) {
	return r.APMerFactory()
}

// GetPrinter will return a printer for printing command output
func (r *Registry) GetPrinter() format.Printer {
	return r.PrinterFactory()
}
