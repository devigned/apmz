package service

import (
	"github.com/devigned/apmz/pkg/format"
)

type (
	// Registry holds the factories and services needed for command execution
	Registry struct {
		PrinterFactory func() format.Printer
	}

	// CommandServicer provides all functionality needed for command execution
	CommandServicer interface {
		GetPrinter() format.Printer
	}

	// CloudPartnerServicer provides Azure Cloud Partner functionality
	CloudPartnerServicer interface {
	}
)

// GetPrinter will return a printer for printing command output
func (r *Registry) GetPrinter() format.Printer {
	return r.PrinterFactory()
}
