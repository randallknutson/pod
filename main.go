package main

import (
	"fmt"
	"time"

	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/agent"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	"github.com/muka/go-bluetooth/hw"
	log "github.com/sirupsen/logrus"
	//	"github.com/muka/go-bluetooth/api/service"
	//	"github.com/muka/go-bluetooth/bluez/profile/agent"
	//	"github.com/muka/go-bluetooth/bluez/profile/gatt"
)

func btInit(adapterID string) {

	btmgmt := hw.NewBtMgmt(adapterID)

	// set LE mode
	btmgmt.SetPowered(false)
	btmgmt.SetLe(true)
	btmgmt.SetBredr(false)
	btmgmt.SetPowered(true)
}

func serve(adapterID string) error {

	options := service.AppOptions{
		AdapterID:  adapterID,
		AgentCaps:  agent.CapNoInputNoOutput,
		UUIDSuffix: "-e3ed-4464-8b7e-751e03d0dc5f",
		UUID:       "1a7e",
	}

	a, err := service.NewApp(options)
	if err != nil {
		return err
	}
	defer a.Close()

	a.SetName("go_bluetooth")

	log.Infof("HW address %s", a.Adapter().Properties.Address)

	if !a.Adapter().Properties.Powered {
		err = a.Adapter().SetPowered(true)
		if err != nil {
			log.Fatalf("Failed to power the adapter: %s", err)
		}
	}

	service1, err := a.NewService("4024")
	if err != nil {
		return err
	}

	err = a.AddService(service1)
	if err != nil {
		return err
	}

	cmdCharacteristic, err := service1.NewChar("2441")
	if err != nil {
		return err
	}

	cmdCharacteristic.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWrite,
		gatt.FlagCharacteristicIndicate,
		//	gatt.FlagCharacteristicNotify,
	}

	cmdDescriptor, err := cmdCharacteristic.NewDescr("2902")
	if err != nil {
		return err
	}

	cmdCharacteristic.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Infof("cmd READ request 1")
		return []byte{42}, nil
	}))

	cmdCharacteristic.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		log.Infof("cmd WRITE %x %x", value, value[0])
		if value[0] == 0x00 { // flow control, the PDM wants to write
			cmdCharacteristic.Confirm()

			w := cmdDescriptor.WriteValue([]byte{1}, nil) // send a RTS
			if w != nil {
				fmt.Println("Error ", w)
				log.Warnf("write err %v", w)
			}

			return []byte{1, 0}, nil
		}

		return value, nil
	}))

	cmdDescriptor.Properties.Flags = []string{
		gatt.FlagDescriptorRead,
		gatt.FlagDescriptorWrite,
	}

	cmdDescriptor.OnRead(service.DescrReadCallback(func(c *service.Descr, options map[string]interface{}) ([]byte, error) {
		log.Infof("cmd DESC READ %+v", options)
		return []byte{42}, nil
	}))

	cmdDescriptor.OnWrite(service.DescrWriteCallback(func(d *service.Descr, value []byte) ([]byte, error) {
		log.Infof("cmd DESC WRITE %x", value)

		return value, nil
	}))
	err = cmdCharacteristic.AddDescr(cmdDescriptor)
	if err != nil {
		return err
	}

	err = service1.AddChar(cmdCharacteristic)
	if err != nil {
		return err
	}

	dataCharacteristic, err := service1.NewChar("2442")
	if err != nil {
		return err
	}

	dataCharacteristic.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWrite,
		gatt.FlagCharacteristicIndicate,
	}

	dataCharacteristic.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Warnf("data READ REQUEST ")
		return []byte{42}, nil
	}))

	dataCharacteristic.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		log.Warnf("data WRITE REQUEST %x", value)
		dataCharacteristic.Confirm()
		return value, nil
	}))

	dataDescriptor, err := dataCharacteristic.NewDescr("2902")
	if err != nil {
		return err
	}

	dataDescriptor.Properties.Flags = []string{
		gatt.FlagDescriptorRead,
		gatt.FlagDescriptorWrite,
	}

	dataDescriptor.OnRead(service.DescrReadCallback(func(c *service.Descr, options map[string]interface{}) ([]byte, error) {
		log.Infof("data desc READ %+v", options)
		return []byte{42}, nil
	}))

	dataDescriptor.OnWrite(service.DescrWriteCallback(func(d *service.Descr, value []byte) ([]byte, error) {
		log.Infof("data desc WRITE %x", value)
		return value, nil
	}))

	err = dataCharacteristic.AddDescr(dataDescriptor)
	if err != nil {
		return err
	}

	err = service1.AddChar(dataCharacteristic)
	if err != nil {
		return err
	}

	err = a.Run()
	if err != nil {
		return err
	}

	log.Infof("Exposed service %s", service1.Properties.UUID)

	timeout := uint32(6 * 3600) // 6h
	adv := a.GetAdvertisement()
	log.Infof("Adv data: %+v", adv)
	adv.Discoverable = true
	adv.ServiceUUIDs = []string{
		"00004024-0000-1000-8000-00805f9b34fb",
		"00002470-0000-1000-8000-00805f9b34fb",
		"0000000a-0000-1000-8000-00805f9b34fb",
		// id
		"0000ffff-0000-1000-8000-00805f9b34fb",
		"0000fffe-0000-1000-8000-00805f9b34fb",
		// lot number, serial
		"0000aaaa-0000-1000-8000-00805f9b34fb",
		"0000aaaa-0000-1000-8000-00805f9b34fb",
		"0000aaaa-0000-1000-8000-00805f9b34fb",
		"0000aaaa-0000-1000-8000-00805f9b34fb",
	}
	// Service UUIDs advertisement format
	// [40 24, 24 70, 00 0a, <<4, ID>>, <<5, Lot number>>, <<3, Serial>> ]

	log.Infof("Advertising for %ds...", timeout)
	cancel, err := a.Advertise(timeout)
	if err != nil {
		return err
	}
	adv = a.GetAdvertisement()
	log.Infof("Adv data: %+v", adv)
	defer cancel()

	time.Sleep((time.Duration(timeout) * time.Second))
	wait := make(chan bool)
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		wait <- true
	}()

	<-wait

	return nil
}

func main() {
	log.SetLevel(log.TraceLevel)
	adapterID := "hci0"
	btInit(adapterID)
	serve(adapterID)
}
