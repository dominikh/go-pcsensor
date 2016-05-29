// Package pcsensor provides a library for reading data from some
// PCsensor temperature sensors.
//
// Currently, only the TEMPer2 device is supported.
package pcsensor

import (
	"fmt"
	"math"

	"github.com/kylelemons/gousb/usb"
)

const (
	idVendor  = 0x0c45
	idProduct = 0x7401

	reqIntLen = 8
)

var uTemperature = []byte{0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00}

// Sensor is a single PCsensor device.
type Sensor interface {
	// Temperatures returns all available temperatures from the
	// sensor.
	Temperatures() (Temperatures, error)
	// Close closes the underlying USB handle. This must be called,
	// even if the program is exiting.
	Close() error
}

// Temperatures map sensor names (e.g. "internal" or "external") to
// measured temperatures, in degrees Celsius.
type Temperatures map[string]float64

type temperv2 struct {
	dev *usb.Device
	ep  usb.Endpoint
}

func (s *temperv2) init() error {
	const interface1 = 0x00
	const interface2 = 0x01
	_, err := s.dev.OpenEndpoint(0x01, interface1, 0, 0x81)
	if err != nil {
		return fmt.Errorf("error opening endpoint 1: %s", err)
	}
	endpoint2, err := s.dev.OpenEndpoint(0x01, interface2, 0, 0x82)
	if err != nil {
		return fmt.Errorf("error opening endpoint 2: %s", err)
	}
	s.ep = endpoint2

	_, err = s.dev.Control(0x21, 0x09, 0x0201, interface1, []byte{0x01, 0x01})
	if err != nil {
		return fmt.Errorf("error communicating with sensor: %s", err)
	}

	err = control(s.dev, uTemperature)
	if err != nil {
		return fmt.Errorf("error communicating with sensor: %s", err)
	}
	b := make([]byte, reqIntLen)
	_, err = endpoint2.Read(b)
	if err != nil {
		return fmt.Errorf("error reading from endpoint: %s", err)
	}

	uIni1 := []byte{0x01, 0x82, 0x77, 0x01, 0x00, 0x00, 0x00, 0x00}
	err = control(s.dev, uIni1)
	if err != nil {
		return fmt.Errorf("error communicating with sensor: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return fmt.Errorf("error reading from endpoint: %s", err)
	}

	uIni2 := []byte{0x01, 0x86, 0xff, 0x01, 0x00, 0x00, 0x00, 0x00}
	err = control(s.dev, uIni2)
	if err != nil {
		return fmt.Errorf("error communicating with sensor: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return fmt.Errorf("error reading from endpoint: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return fmt.Errorf("error reading from endpoint: %s", err)
	}
	return nil
}

func (s *temperv2) Close() error {
	return s.dev.Close()
}

func fm75(b1, b2 byte) float64 {
	if b1 == 255 && b2 == 255 {
		return math.NaN()
	}
	return float64(int16(b1)<<8|int16(b2)) / 256
}

func (s *temperv2) Temperatures() (temps Temperatures, err error) {
	err = control(s.dev, uTemperature)
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}
	b := make([]byte, reqIntLen)
	_, err = s.ep.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}
	inner := fm75(b[2], b[3])
	outer := fm75(b[4], b[5])

	return Temperatures{
		"inner": inner,
		"outer": outer,
	}, nil
}

func control(dev *usb.Device, msg []byte) error {
	_, err := dev.Control(0x21, 0x09, 0x200, 0x01, msg)
	return err
}

// New returns all connected sensors.
//
// Each sensor must be closed when done using it. Sensors need to be
// closed even when the program exits, to release USB handles.
//
// BUG(dh): Currently, only one sensor will be returned.
func New(ctx *usb.Context) ([]Sensor, error) {
	// TODO find all sensors, not just the first
	dev, err := ctx.OpenDeviceWithVidPid(idVendor, idProduct)
	if err != nil {
		return nil, err
	}
	sensor := &temperv2{dev: dev}
	if err = sensor.init(); err != nil {
		_ = sensor.Close()
		return nil, err
	}
	return []Sensor{sensor}, nil
}
