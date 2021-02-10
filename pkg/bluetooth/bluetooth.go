package bluetooth

import (
	"encoding/hex"
	"fmt"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/linux/cmd"
	log "github.com/sirupsen/logrus"
)

type PacketType int

type Packet struct {
	Type PacketType
	Data []byte
}

const (
	PacketTypeCmd PacketType = iota
	PacketTypeData
)

var (
	CmdRTS Packet = Packet{
		Type: PacketTypeCmd,
		Data: []byte{0},
	}
	CmdCTS Packet = Packet{
		Type: PacketTypeCmd,
		Data: []byte{1},
	}
	CmdNACK Packet = Packet{
		Type: PacketTypeCmd,
		Data: []byte{2},
	}
	CmdSuccess Packet = Packet{
		Type: PacketTypeCmd,
		Data: []byte{4},
	}
)

type Ble struct {
	input      chan Packet
	dataOutput chan Packet
	cmdOutput  chan Packet
}

var DefaultServerOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
	gatt.LnxSetAdvertisingParameters(&cmd.LESetAdvertisingParameters{
		AdvertisingIntervalMin: 0x00f4,
		AdvertisingIntervalMax: 0x00f4,
		AdvertisingChannelMap:  0x7,
	}),
}

func New(adapterID string) (*Ble, error) {
	b := &Ble{
		input:      make(chan Packet, 5),
		dataOutput: make(chan Packet, 5),
		cmdOutput:  make(chan Packet, 5),
	}

	d, err := gatt.NewDevice(DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { fmt.Println("Disconnect: ", c.ID()) }),
	)

	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			var serviceUUID = gatt.MustParseUUID("1a7e-4024-e3ed-4464-8b7e-751e03d0dc5f")
			var cmdCharUUID = gatt.MustParseUUID("1a7e-2441-e3ed-4464-8b7e-751e03d0dc5f")
			var dataCharUUID = gatt.MustParseUUID("1a7e-2442-e3ed-4464-8b7e-751e03d0dc5f")

			s := gatt.NewService(serviceUUID)

			cmdCharacteristic := s.AddCharacteristic(cmdCharUUID)
			cmdCharacteristic.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Debugf("Received CMD:  %x", data)
					b.input <- Packet{
						Type: PacketTypeCmd,
						Data: data,
					}
					return 0
				})

			cmdCharacteristic.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					log.Infof("Enabled notifications for CMD: %s", r.Central.ID())
					go func() {
						for {
							if n.Done() {
								log.Fatalf("CMD closed")
							}
							packet := <-b.cmdOutput
							ret, err := n.Write(packet.Data)
							log.Debugf("CMD notification return: %d", ret)
							if err != nil {
								log.Fatalf("Error writing CMD: %s", err)
							}
						}
					}()
				})

			dataCharacteristic := s.AddCharacteristic(dataCharUUID)
			dataCharacteristic.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					log.Infof("Enabled notifications for DATA: %s", r.Central.ID())
					go func() {
						for {
							if n.Done() {
								log.Fatalf("DATA closed")
							}
							packet := <-b.dataOutput
							ret, err := n.Write(packet.Data)
							log.Info("DATA notification return: %d", ret)
							if err != nil {
								log.Fatalf("Error writing DATA: %s", err)
							}
						}
					}()
				})

			dataCharacteristic.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Debugf("Received DATA %x", data)
					b.input <- Packet{
						Type: PacketTypeData,
						Data: data,
					}
					return 0
				})

			d.AddService(s)

			// Advertise device name and service's UUIDs.
			d.AdvertiseNameAndServices(" :: Fake POD ::", []gatt.UUID{
				gatt.UUID16(0x4024),
				gatt.UUID16(0x2470),
				gatt.UUID16(0x000a),
				gatt.UUID16(0xffff),
				gatt.UUID16(0xfffe),
				gatt.UUID16(0xaaaa),
				gatt.UUID16(0xaaaa),
				gatt.UUID16(0xaaaa),
				gatt.UUID16(0xaaaa),
			})
		default:
		}
	}

	d.Init(onStateChanged)
	return b, nil
}

func (b *Ble) Close() {

}

func (b *Ble) WriteCmd(packet Packet) error {

	b.cmdOutput <- packet
	return nil
}

func (b *Ble) WriteData(packet Packet) error {
	b.dataOutput <- packet
	return nil
}

func (b *Ble) Read() (Packet, error) {
	packet := <-b.input
	return packet, nil
}

func (p Packet) String() string {
	t := "CMD"
	if p.Type == PacketTypeData {
		t = "DATA"
	}
	return fmt.Sprintf("%s : %s", t, hex.EncodeToString(p.Data))
}
