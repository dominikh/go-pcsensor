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

type Sensor struct {
	dev *usb.Device
	ep  usb.Endpoint
}

func control(dev *usb.Device, msg []byte) error {
	_, err := dev.Control(0x21, 0x09, 0x200, 0x01, msg)
	return err
}

func New(ctx *usb.Context) (sensor *Sensor, err error) {
	dev, err := ctx.OpenDeviceWithVidPid(idVendor, idProduct)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = dev.Close()
		}
	}()

	const interface1 = 0x00
	const interface2 = 0x01
	_, err = dev.OpenEndpoint(0x01, interface1, 0, 0x81)
	if err != nil {
		return nil, fmt.Errorf("error opening endpoint 1: %s", err)
	}
	endpoint2, err := dev.OpenEndpoint(0x01, interface2, 0, 0x82)
	if err != nil {
		return nil, fmt.Errorf("error opening endpoint 2: %s", err)
	}

	_, err = dev.Control(0x21, 0x09, 0x0201, interface1, []byte{0x01, 0x01})
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}

	uTemperature := []byte{0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00}
	err = control(dev, uTemperature)
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}
	b := make([]byte, reqIntLen)
	_, err = endpoint2.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}

	uIni1 := []byte{0x01, 0x82, 0x77, 0x01, 0x00, 0x00, 0x00, 0x00}
	err = control(dev, uIni1)
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}

	uIni2 := []byte{0x01, 0x86, 0xff, 0x01, 0x00, 0x00, 0x00, 0x00}
	err = control(dev, uIni2)
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}
	_, err = endpoint2.Read(b)
	if err != nil {
		return nil, fmt.Errorf("error reading from endpoint: %s", err)
	}

	err = control(dev, uTemperature)
	if err != nil {
		return nil, fmt.Errorf("error communicating with sensor: %s", err)
	}

	return &Sensor{dev, endpoint2}, nil
}

func (s *Sensor) Close() error {
	return s.dev.Close()
}

func (s *Sensor) Temperatures() (inner, outer float64, err error) {
	b := make([]byte, reqIntLen)
	_, err = s.ep.Read(b)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading from endpoint: %s", err)
	}
	temp := int16(b[3]&0xFF) + (int16(b[2]) << 8)
	inner = float64(temp) * (125.0 / 32000.0)

	temp = int16(b[5]&0xFF) + (int16(b[4]) << 8)
	outer = float64(temp) * (125.0 / 32000.0)

	return inner, outer, nil
}
