// Package pcsensor provides a library for reading data from
// PCSensor/TEMPer2 temperature sensors.
package main

//package pcsensor

import (
	"log"

	"github.com/kylelemons/gousb/usb"
)

// const static char uTemperatura[] = { 0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00 };

const (
	idVendor  = 0x0c45
	idProduct = 0x7401
)

// interrupt_read(lvr_winusb);

// control_transfer(lvr_winusb, uIni1 );
// interrupt_read(lvr_winusb);

// control_transfer(lvr_winusb, uIni2 );
// interrupt_read(lvr_winusb);
// interrupt_read(lvr_winusb);

func main() {
	ctx := usb.NewContext()
	defer ctx.Close()

	ctx.Debug(3)
	// TODO there may be more than one device
	dev, err := ctx.OpenDeviceWithVidPid(idVendor, idProduct)
	if err != nil {
		log.Fatal(err) // XXX
	}
	defer dev.Close()

	const interface1 = 0x00
	const interface2 = 0x01
	endpoint1, err := dev.OpenEndpoint(0x01, interface1, 0, 0x81)
	if err != nil {
		log.Fatalln("Error opening endpoint 1:", err) // XXX
	}
	endpoint2, err := dev.OpenEndpoint(0x01, interface2, 0, 0x82)
	if err != nil {
		log.Fatalln("Error opening endpoint 2:", err) // XXX
	}

	_ = endpoint1 // XXX
	_ = endpoint2 // XXX

	// ini_control_transfer(lvr_winusb);
	_, err = dev.Control(0x21, 0x09, 0x0201, 0x00, []byte{0x01, 0x01})
	if err != nil {
		log.Fatalln("Error communicating with sensor (1):", err) // XXX
	}

	// control_transfer(lvr_winusb, uTemperatura );
	uTemperatura := []byte{0x01, 0x80, 0x33, 0x01, 0x00, 0x00, 0x00, 0x00}
	_, err = dev.Control(0x21, 0x09, 0x0200, 0x01, uTemperatura)
	if err != nil {
		log.Fatalln("Error communicating with sensor (2):", err) // XXX
	}

	// interrupt_read(lvr_winusb);
	const reqIntLen = 8
	b := make([]byte, reqIntLen)
	_, err = endpoint2.Read(b)
	if err != nil {
		log.Fatalln("Error reading from endpoint:", err) // XXX
	}

	// control_transfer(lvr_winusb, uIni1 );
	uIni1 := []byte{0x01, 0x82, 0x77, 0x01, 0x00, 0x00, 0x00, 0x00}
	_, err = dev.Control(0x21, 0x09, 0x0200, 0x01, uIni1)
	if err != nil {
		log.Fatalln("Error communicating with sensor (3):", err) // XXX
	}
	// interrupt_read(lvr_winusb);
	_, err = endpoint2.Read(b)
	if err != nil {
		log.Fatalln("Error reading from endpoint:", err) // XXX
	}

	// control_transfer(lvr_winusb, uIni2 );
	uIni2 := []byte{0x01, 0x86, 0xff, 0x01, 0x00, 0x00, 0x00, 0x00}
	_, err = dev.Control(0x21, 0x09, 0x0200, 0x01, uIni2)
	if err != nil {
		log.Fatalln("Error communicating with sensor (4):", err) // XXX
	}
	// interrupt_read(lvr_winusb);
	_, err = endpoint2.Read(b)
	if err != nil {
		log.Fatalln("Error reading from endpoint:", err) // XXX
	}
	// interrupt_read(lvr_winusb);
	_, err = endpoint2.Read(b)
	if err != nil {
		log.Fatalln("Error reading from endpoint:", err) // XXX
	}

	// control_transfer(lvr_winusb, uTemperatura );
	_, err = dev.Control(0x21, 0x09, 0x0200, 0x01, uTemperatura)
	if err != nil {
		log.Fatalln("Error communicating with sensor (5):", err) // XXX
	}

	// interrupt_read_temperatura(lvr_winusb, &tempInC, &tempOutC);
	_, err = endpoint2.Read(b)
	if err != nil {
		log.Fatalln("Error reading from endpoint:", err) // XXX
	}
	temp := int16(b[3]&0xFF) + (int16(b[2]) << 8)
	log.Println(float64(temp) * (125.0 / 32000.0))

	temp = int16(b[5]&0xFF) + (int16(b[4]) << 8)
	log.Println(float64(temp) * (125.0 / 32000.0))
}
