package mocks

import (
	"context"

	"github.com/devigned/apmz-sdk/apmz"
	"github.com/stretchr/testify/mock"

	"github.com/devigned/apmz/pkg/azmeta"
	"github.com/devigned/apmz/pkg/format"
	"github.com/devigned/apmz/pkg/service"
)

type (
	ServiceMock struct {
		mock.Mock
		Printer *PrinterMock
		APM     *APMMock
	}

	PrinterMock struct {
		mock.Mock
	}

	APMMock struct {
		mock.Mock
	}

	MetadataMock struct {
		mock.Mock
	}
)

func (sm *ServiceMock) GetAPMer() (service.APMer, error) {
	args := sm.Called()
	return args.Get(0).(service.APMer), args.Error(1)
}

func (sm *ServiceMock) GetPrinter() format.Printer {
	args := sm.Called()
	return args.Get(0).(format.Printer)
}

func (sm *ServiceMock) GetKey() string {
	args := sm.Called()
	return args.String(0)
}

func (sm *ServiceMock) GetMetadater() (service.Metadater, error) {
	args := sm.Called()
	return args.Get(0).(service.Metadater), args.Error(1)
}

func (pm *PrinterMock) Print(obj interface{}) error {
	args := pm.Called(obj)
	return args.Error(0)
}

func (pm *PrinterMock) Printf(format string, args ...interface{}) {
	pm.Called(format, args)
}

func (pm *PrinterMock) ErrPrintf(format string, args ...interface{}) {
	pm.Called(format, args)
}

func (am *APMMock) Track(telemetry apmz.Telemetry) {
	am.Called(telemetry)
}

func (am *APMMock) Channel() apmz.TelemetryChannel {
	args := am.Called()
	return args.Get(0).(apmz.TelemetryChannel)
}

func (mm *MetadataMock) GetInstance(ctx context.Context, middleware ...azmeta.MiddlewareFunc) (*azmeta.Instance, error) {
	args := mm.Called(ctx, middleware)
	return args.Get(0).(*azmeta.Instance), args.Error(1)
}

func (mm *MetadataMock) GetAttestation(ctx context.Context, nonce string, middleware ...azmeta.MiddlewareFunc) (*azmeta.Attestation, error) {
	args := mm.Called(ctx, nonce, middleware)
	return args.Get(0).(*azmeta.Attestation), args.Error(1)
}

func (mm *MetadataMock) GetScheduledEvents(ctx context.Context, middleware ...azmeta.MiddlewareFunc) (*azmeta.ScheduledEvents, error) {
	args := mm.Called(ctx, middleware)
	return args.Get(0).(*azmeta.ScheduledEvents), args.Error(1)
}

func (mm *MetadataMock) AckScheduledEvents(ctx context.Context, acks azmeta.AckEvents, middleware ...azmeta.MiddlewareFunc) error {
	args := mm.Called(ctx, acks, middleware)
	return args.Error(0)
}

func (mm *MetadataMock) GetIdentityToken(ctx context.Context, tokenReq azmeta.ResourceAndIdentity, middleware ...azmeta.MiddlewareFunc) (*azmeta.IdentityToken, error) {
	args := mm.Called(ctx, tokenReq, middleware)
	return args.Get(0).(*azmeta.IdentityToken), args.Error(1)
}
