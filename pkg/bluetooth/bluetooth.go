package bluetooth

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/linux/cmd"
	log "github.com/sirupsen/logrus"
)

type Packet []byte

var (
	CmdRTS     = Packet([]byte{0})
	CmdCTS     = Packet([]byte{1})
	CmdNACK    = Packet([]byte{2, 0})
	CmdAbort   = Packet([]byte{3})
	CmdSuccess = Packet([]byte{4})
	CmdFail    = Packet([]byte{5})
)

type Ble struct {
	dataInput  chan Packet
	cmdInput   chan Packet
	dataOutput chan Packet
	cmdOutput  chan Packet

	messageInput  chan *Message
	messageOutput chan *Message
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
		dataInput:     make(chan Packet, 5),
		cmdInput:      make(chan Packet, 5),
		dataOutput:    make(chan Packet, 5),
		cmdOutput:     make(chan Packet, 5),
		messageInput:  make(chan *Message, 5),
		messageOutput: make(chan *Message, 2),
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
					ret := make([]byte, len(data))
					copy(ret, data)
					b.cmdInput <- Packet(ret)
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
							ret, err := n.Write(packet)
							log.Debugf("CMD notification return: %d/%s", ret, hex.EncodeToString(packet))
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
							ret, err := n.Write(packet)
							log.Debugf("DATA notification return: %d/%s", ret, hex.EncodeToString(packet))
							if err != nil {
								log.Fatalf("Error writing DATA: %s ", err)
							}
						}
					}()
				})

			dataCharacteristic.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Debugf("Received DATA %x", data)
					ret := make([]byte, len(data))
					copy(ret, data)
					b.dataInput <- Packet(ret)
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
	go b.loop()
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

func (b *Ble) ReadCmd() (Packet, error) {
	packet := <-b.cmdInput
	return packet, nil
}

func (b *Ble) ReadData() (Packet, error) {
	packet := <-b.dataInput
	return packet, nil
}

func (p Packet) String() string {
	return hex.EncodeToString(p)
}

func (m Message) String() string {
	return hex.EncodeToString(m.Data)
}

func (b *Ble) ReadMessage() (*Message, error) {
	message := <-b.messageInput
	return message, nil
}

func (b *Ble) WriteMessage(message *Message) {
	b.messageOutput <- message
}

func (b *Ble) loop() {
	for {
		select {
		case msg := <-b.messageOutput:
			b.writeMessage(msg)
		case cmd := <-b.cmdInput:
			msg, err := b.readMessage(cmd)
			if err != nil {
				log.Fatalf("Error reading message: %s", err)
			}
			b.messageInput <- msg
		}
	}
}

func (b *Ble) expectCommand(expected Packet) {
	cmd, _ := b.ReadCmd()
	if bytes.Compare(expected, cmd) != 0 {
		log.Fatalf("Expected: %s. Received: %s", expected, cmd)
	}
}

func (b *Ble) writeMessage(msg *Message) {

	b.WriteCmd(CmdRTS)
	b.expectCommand(CmdCTS)

	// serialize the thing
	// and split it in packets
	b.expectCommand(CmdSuccess)
}

func (b *Ble) readMessage(cmd Packet) (*Message, error) {
	var buf bytes.Buffer
	var checksum []byte

	if bytes.Compare(cmd, CmdRTS) != 0 {
		log.Fatalf("Expected: %s. Received: %s", CmdRTS, cmd)
	}
	b.WriteCmd(CmdCTS)

	first, _ := b.ReadData()
	fragments := int(first[1])
	expectedIndex := 1
	oneExtra := false
	if fragments == 0 {
		checksum = first[2:6]
		len := first[6]
		end := len + 7
		if len > 13 {
			oneExtra = true
			end = 20
		}
		buf.Write(first[7:end])
	} else {
		buf.Write(first[2:20])
	}
	for i := 1; i < fragments; i++ {
		data, _ := b.ReadData()
		if i == expectedIndex {
			buf.Write(data[1:20])
		} else {
			log.Warnf("Sending NACK, packet index is wrong")
			buf.Write(data[:])
			CmdNACK[1] = byte(expectedIndex)
			b.WriteCmd(CmdNACK)
		}
		expectedIndex++
	}
	if fragments != 0 {
		data, _ := b.ReadData()
		len := data[1]
		if len > 14 {
			oneExtra = true
			len = 14
		}
		checksum = data[2:6]
		buf.Write(data[6 : len+6])
	}
	log.Debugf("One extra: %b", oneExtra)
	if oneExtra {
		data, _ := b.ReadData()
		buf.Write(data[2 : data[1]+2])
	}
	bytes := buf.Bytes()
	sum := crc32.ChecksumIEEE(bytes)
	if binary.BigEndian.Uint32(checksum) != sum {
		log.Warnf("Checksum missmatch. checksum is: %x. want: %d", sum, checksum)
		log.Warnf("Data: %s", hex.EncodeToString(bytes))

		b.WriteCmd(CmdFail)
		return nil, errors.New("Checksum missmatch")
	}

	b.WriteCmd(CmdSuccess)

	return fromByteArray(bytes)
}
