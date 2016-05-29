// Package pcsensor provides a library for reading data from
// PCSensor/TEMPer2 temperature sensors.
package pcsensor

import (
	"fmt"

	"github.com/kylelemons/gousb/usb"
)

const (
	idVendor  = 0x0c45
	idProduct = 0x7401

	reqIntLen = 8
)

type Sensor interface {
	Temperatures() (Temperatures, error)
	Close() error
}

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

	uTemperature := []byte{0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00}
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

	err = control(s.dev, uTemperature)
	if err != nil {
		return fmt.Errorf("error communicating with sensor: %s", err)
	}

	return nil
}

func control(dev *usb.Device, msg []byte) error {
	_, err := dev.Control(0x21, 0x09, 0x200, 0x01, msg)
	return err
}

func New(ctx *usb.Context) (Sensor, error) {
	// TODO find all sensors, not just the first
	dev, err := ctx.OpenDeviceWithVidPid(idVendor, idProduct)
	if err != nil {
		return nil, err
	}
	sensor := &temperv2{dev: dev}
	if err = sensor.init(); err != nil {
		sensor.Close()
		return nil, err
	}
	return sensor, nil
}

func (s *temperv2) Close() error {
	return s.dev.Close()
}

func (s *temperv2) Temperatures() (temps Temperatures, err error) {
	b := make([]byte, reqIntLen)
	_, err = s.ep.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}
	// TODO look up a spec sheet of the sensor to make sense of these
	// values
	temp := int16(b[3]&0xFF) + (int16(b[2]) << 8)
	inner := float64(temp) * (125.0 / 32000.0)

	temp = int16(b[5]&0xFF) + (int16(b[4]) << 8)
	outer := float64(temp) * (125.0 / 32000.0)

	return Temperatures{
		"inner": inner,
		"outer": outer,
	}, nil
}
