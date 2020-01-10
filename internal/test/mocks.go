package mocks

import (
	"github.com/devigned/apmz-sdk/apmz"
	"github.com/stretchr/testify/mock"

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
