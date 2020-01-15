package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/devigned/apmz-sdk/apmz"

	"github.com/devigned/apmz/pkg/azmeta"
	"github.com/devigned/apmz/pkg/format"
)

type (
	// Registry holds the factories and services needed for command execution
	Registry struct {
		APMerFactory    func() (APMer, error)
		PrinterFactory  func() format.Printer
		APIKeyFactory   func() string
		MetadataFactory func() (Metadater, error)
	}

	// CommandServicer provides all functionality needed for command execution
	CommandServicer interface {
		GetMetadater() (Metadater, error)
		GetAPMer() (APMer, error)
		GetPrinter() format.Printer
		GetKey() string
	}

	// Metadater abstracts the underlying implementation of the instance metadata service
	Metadater interface {
		GetInstance(ctx context.Context, middleware ...azmeta.MiddlewareFunc) (*azmeta.Instance, error)
		GetAttestation(ctx context.Context, nonce string, middleware ...azmeta.MiddlewareFunc) (*azmeta.Attestation, error)
		GetScheduledEvents(ctx context.Context, middleware ...azmeta.MiddlewareFunc) (*azmeta.ScheduledEvents, error)
		AckScheduledEvents(ctx context.Context, acks azmeta.AckEvents, middleware ...azmeta.MiddlewareFunc) error
	}

	// APMer provides the behaviors needed to send events to Azure Application Insights
	APMer interface {
		Track(telemetry apmz.Telemetry)
		Channel() apmz.TelemetryChannel
	}

	// APMZProxy will proxy calls to the APMZ client or print if running locally
	APMZProxy struct {
		Printer format.Printer
		apmz.TelemetryClient
	}

	// EventType represents the enumeration of all the event types the Batch command understands
	EventType string

	// Event is a typed batch event
	Event struct {
		Type string         `json:"type,omitempty"`
		Item apmz.Telemetry `json:"item,omitempty"`
	}
)

const (
	// Trace is a "traces" event type in Application Insights
	Trace EventType = "trace"
	// Metric is a "customMetrics" event type in Application Insights
	Metric EventType = "metric"
)

// GetAPMer returns an instance of an Azure Application Insights client
func (r *Registry) GetAPMer() (APMer, error) {
	return r.APMerFactory()
}

// GetMetadater returns an instance of an instance metadata service
func (r *Registry) GetMetadater() (Metadater, error) {
	return r.MetadataFactory()
}

// GetPrinter will return a printer for printing command output
func (r *Registry) GetPrinter() format.Printer {
	return r.PrinterFactory()
}

// GetKey will return the api-key for application insights
func (r *Registry) GetKey() string {
	return r.APIKeyFactory()
}

// Track will either send to the client or print depending if the proxy printer is set
func (apmzp APMZProxy) Track(item apmz.Telemetry) {
	if apmzp.Printer != nil {
		t := reflect.TypeOf(item)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		evt := Event{
			Type: t.Name(),
			Item: item,
		}
		_ = apmzp.Printer.Print(evt)
		return
	}

	apmzp.TelemetryClient.Track(item)
}

// UnmarshalJSON takes json bytes and turns them into an event
func (evt *Event) UnmarshalJSON(b []byte) error {
	tmp := &struct {
		Type string `json:"type,omitempty"`
		Item *json.RawMessage
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	evt.Type = tmp.Type
	var telemetry apmz.Telemetry
	switch evt.Type {
	case "TraceTelemetry":
		tt := &apmz.TraceTelemetry{}
		if err := json.Unmarshal(*tmp.Item, tt); err != nil {
			return err
		}
		telemetry = tt
	case "MetricTelemetry":
		mt := &apmz.MetricTelemetry{}
		if err := json.Unmarshal(*tmp.Item, mt); err != nil {
			return err
		}
		telemetry = mt
	default:
		return fmt.Errorf("don't know how to unmarshal type: %v", evt.Type)
	}

	evt.Item = telemetry
	return nil
}
