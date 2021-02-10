package bluetooth

import (
	"encoding/hex"
	"fmt"

	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/agent"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	"github.com/muka/go-bluetooth/hw"
	log "github.com/sirupsen/logrus"
)

type PacketType int

const (
	PacketTypeCmd PacketType = iota
	PacketTypeData
)

type Packet struct {
	Type PacketType
	Data []byte
}

type Ble struct {
	input              chan Packet
	output             chan Packet
	app                *service.App
	cancelAdv          func()
	cmdCharacteristic  *service.Char
	dataCharacteristic *service.Char
}

func btInit(adapterID string) {

	btmgmt := hw.NewBtMgmt(adapterID)

	// set LE mode
	btmgmt.SetPowered(false)
	btmgmt.SetLe(true)
	btmgmt.SetBredr(false)
	btmgmt.SetPowered(true)
}

func New(adapterID string) (*Ble, error) {

	btInit(adapterID)

	b := &Ble{
		input:  make(chan Packet, 5),
		output: make(chan Packet, 5),
	}
	go b.sendLoop()
	options := service.AppOptions{
		AdapterID:  adapterID,
		AgentCaps:  agent.CapNoInputNoOutput,
		UUIDSuffix: "-e3ed-4464-8b7e-751e03d0dc5f",
		UUID:       "1a7e",
	}

	a, err := service.NewApp(options)
	b.app = a
	if err != nil {
		return b, err
	}

	a.SetName(":: Fake POD ::")

	log.Infof("HW address %s", a.Adapter().Properties.Address)

	if !a.Adapter().Properties.Powered {
		err = a.Adapter().SetPowered(true)
		if err != nil {
			log.Fatalf("Failed to power the adapter: %s", err)
		}
	}

	service1, err := a.NewService("4024")
	if err != nil {
		return b, err
	}
	if err = a.AddService(service1); err != nil {
		return b, err
	}

	cmdCharacteristic, err := service1.NewChar("2441")
	b.cmdCharacteristic = cmdCharacteristic
	if err != nil {
		return b, err
	}

	cmdCharacteristic.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWriteWithoutResponse,
		gatt.FlagCharacteristicIndicate,
	}
	cmdCharacteristic.StopNotify()

	cmdCharacteristic.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Fatalf("CMD read request. We did not expect that")
		return []byte{00}, nil
	}))

	cmdCharacteristic.OnWrite(service.CharWriteCallback(func(char *service.Char, value []byte) ([]byte, error) {
		log.Infof("received command %x", value)
		packet := Packet{
			Type: PacketTypeCmd,
			Data: value,
		}
		cmdCharacteristic.StopNotify()
		b.input <- packet
		return value, nil
	}))

	cmdDescriptor, err := cmdCharacteristic.NewDescr("2902")
	if err != nil {
		return b, err
	}
	cmdDescriptor.Properties.Flags = []string{
		gatt.FlagDescriptorRead,
		gatt.FlagDescriptorWrite,
	}

	cmdDescriptor.OnRead(service.DescrReadCallback(func(c *service.Descr, options map[string]interface{}) ([]byte, error) {
		log.Fatalf("cmd DESC read, did not expect that ")
		return []byte{00}, nil
	}))

	cmdDescriptor.OnWrite(service.DescrWriteCallback(func(d *service.Descr, value []byte) ([]byte, error) {
		log.Infof("CMD Descriptor: %x", value)
		return value, nil
	}))

	err = cmdCharacteristic.AddDescr(cmdDescriptor)
	if err != nil {
		return b, err
	}

	err = service1.AddChar(cmdCharacteristic)
	if err != nil {
		return b, err
	}

	dataCharacteristic, err := service1.NewChar("2442")
	b.dataCharacteristic = dataCharacteristic
	if err != nil {
		return b, err
	}

	dataCharacteristic.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWrite,
		gatt.FlagCharacteristicIndicate,
	}

	dataCharacteristic.OnRead(service.CharReadCallback(func(c *service.Char, options map[string]interface{}) ([]byte, error) {
		log.Fatalf("DATA read request. Did not expect that")
		return []byte{0}, nil
	}))

	dataCharacteristic.OnWrite(service.CharWriteCallback(func(c *service.Char, value []byte) ([]byte, error) {
		log.Warnf("received DATA %x", value)
		packet := Packet{
			Type: PacketTypeData,
			Data: value,
		}
		b.input <- packet
		return value, nil
	}))

	dataDescriptor, err := dataCharacteristic.NewDescr("2902")
	if err != nil {
		return b, err
	}

	dataDescriptor.Properties.Flags = []string{
		gatt.FlagDescriptorRead,
		gatt.FlagDescriptorWrite,
	}

	dataDescriptor.OnRead(service.DescrReadCallback(func(c *service.Descr, options map[string]interface{}) ([]byte, error) {
		log.Fatal("data desc READ. Did not expect that")
		return []byte{0}, nil
	}))

	dataDescriptor.OnWrite(service.DescrWriteCallback(func(d *service.Descr, value []byte) ([]byte, error) {
		log.Debugf("data desc WRITE %x", value)
		return value, nil
	}))

	err = dataCharacteristic.AddDescr(dataDescriptor)
	if err != nil {
		return b, err
	}

	err = service1.AddChar(dataCharacteristic)
	if err != nil {
		return b, err
	}

	err = a.Run()
	if err != nil {
		return b, err
	}

	log.Infof("Exposed service %s", service1.Properties.UUID)

	adv := a.GetAdvertisement()
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

	timeout := uint32(12 * 3600) // 12h
	log.Infof("Advertising for %ds...", timeout)
	cancel, err := a.Advertise(timeout)
	b.cancelAdv = cancel
	if err != nil {
		return b, err
	}
	adv = a.GetAdvertisement()
	log.Infof("Adv data: %+v", adv)

	return b, nil
}

func (b *Ble) Close() {
	if b.app != nil {
		b.app.Close()
	}
	if b.cancelAdv != nil {
		b.cancelAdv()
	}
}

func (b *Ble) Write(packet Packet) error {
	b.output <- packet
	return nil
}

func (b *Ble) Read() (Packet, error) {
	packet := <-b.input
	return packet, nil
}

func (b *Ble) sendLoop() {
	for out := range b.output {
		char := b.dataCharacteristic
		if out.Type == PacketTypeCmd {
			char = b.cmdCharacteristic
		}
		char.StartNotify()
		err := char.WriteValue(out.Data, nil)
		if err != nil {
			log.Fatalf("Error writing: %v", err)
		}
		char.StopNotify()
	}
}

func (p Packet) String() string {
	t := "CMD"
	if p.Type == PacketTypeData {
		t = "DATA"
	}
	return fmt.Sprintf("%s : %s", t, hex.EncodeToString(p.Data))
}
