package bluetooth

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"sync"
	"time"

	"github.com/avereha/pod/pkg/message"
	"github.com/davecgh/go-spew/spew"
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

	messageInput  chan *message.Message
	messageOutput chan *message.Message

	stopLoop chan bool
	device   *gatt.Device
	central  *gatt.Central

	cmdNotifier    gatt.Notifier
	cmdNotifierMtx sync.Mutex

	dataNotifier    gatt.Notifier
	dataNotifierMtx sync.Mutex
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

func New(adapterID string, podId []byte) (*Ble, error) {
	d, err := gatt.NewDevice(DefaultServerOptions...)
	if err != nil {
		log.Fatalf("pkg bluetooth; failed to open device, err: %s", err)
	}

	b := &Ble{
		dataInput:     make(chan Packet, 5),
		cmdInput:      make(chan Packet, 5),
		dataOutput:    make(chan Packet, 5),
		cmdOutput:     make(chan Packet, 5),
		messageInput:  make(chan *message.Message, 5),
		messageOutput: make(chan *message.Message, 2),
		device:        &d,
	}

	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) {
			fmt.Println("pkg bluetooth; ** New connection from: ", c.ID())
			// b.StopMessageLoop()
			b.central = &c
		}),
		gatt.CentralDisconnected(func(c gatt.Central) {
			log.Tracef("pkg bluetooth; ** disconnect: %s", c.ID())
		}),
	)

	// Start cmd writing goroutine
	go func() {
		for {
			packet := <-b.cmdOutput
			b.cmdNotifierMtx.Lock()
			if b.cmdNotifier.Done() {
				log.Fatalf("pkg bluetooth; CMD closed")
			}
			ret, err := b.cmdNotifier.Write(packet)
			b.cmdNotifierMtx.Unlock()
			log.Tracef("pkg bluetooth; CMD notification return: %d/%s", ret, hex.EncodeToString(packet))
			if err != nil {
				log.Fatalf("pkg bluetooth; error writing CMD: %s", err)
			}
		}
	}()

	// Start data writing goroutine
	go func() {
		for {
			packet := <-b.dataOutput
			b.dataNotifierMtx.Lock()
			if b.dataNotifier.Done() {
				log.Fatalf("pkg bluetooth; DATA closed")
			}
			ret, err := b.dataNotifier.Write(packet)
			b.dataNotifierMtx.Unlock()
			log.Tracef("pkg bluetooth; DATA notification return: %d/%s", ret, hex.EncodeToString(packet))
			if err != nil {
				log.Fatalf("pkg bluetooth; error writing DATA: %s ", err)
			}
		}
	}()

	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("state: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			var serviceUUID = gatt.MustParseUUID("1a7e-4024-e3ed-4464-8b7e-751e03d0dc5f")
			var cmdCharUUID = gatt.MustParseUUID("1a7e-2441-e3ed-4464-8b7e-751e03d0dc5f")  // 0x2E // h 0x000d // vh 0x000e
			var dataCharUUID = gatt.MustParseUUID("1a7e-2443-e3ed-4464-8b7e-751e03d0dc5f") // 0x16 // h 0x0010 // vh 0x0011

			var service2UUID = gatt.MustParseUUID("ECF301E2-674B-4474-94D0-364F3AA653E6")
			var heartbeatCharUUID = gatt.MustParseUUID("7DED7A6C-CA72-46A7-A3A2-6061F6FDCAEB") // 0x22

			s := gatt.NewService(serviceUUID)

			cmdCharacteristic := s.AddCharacteristic(cmdCharUUID)
			cmdCharacteristic.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Tracef("pkg bluetooth; received CMD, %x -- %d", data, len(data))
					ret := make([]byte, len(data))
					copy(ret, data)
					b.cmdInput <- Packet(ret)
					return 0
				})

			cmdCharacteristic.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					b.cmdNotifierMtx.Lock()
					b.cmdNotifier = n
					b.cmdNotifierMtx.Unlock()
					log.Infof("pkg bluetooth; handling CMD notifications on new connection from:  %s", r.Central.ID())
				})

			dataCharacteristic := s.AddCharacteristic(dataCharUUID)
			dataCharacteristic.HandleNotifyFunc(
				func(r gatt.Request, n gatt.Notifier) {
					b.dataNotifierMtx.Lock()
					b.dataNotifier = n
					b.dataNotifierMtx.Unlock()
					log.Infof("pkg bluetooth; handling DATA notifications on new connection from: %s", r.Central.ID())
					log.Infof("     *** OK to send commands from the phone app ***")
				})

			dataCharacteristic.HandleWriteFunc(
				func(r gatt.Request, data []byte) (status byte) {
					log.Tracef("pkg bluetooth; received DATA, %x -- %d", data, len(data))
					ret := make([]byte, len(data))
					copy(ret, data)
					b.dataInput <- Packet(ret)
					return 0
				})

			h := gatt.NewService(service2UUID)
			hbCharacteristic := s.AddCharacteristic(heartbeatCharUUID)
			hbCharacteristic.HandleReadFunc(
				func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
					for range time.Tick(time.Second * 10) {
						rsp.Write([]byte{0x00})
						log.Tracef("pkg bluetooth; heartbeat sent")
					}
				},
			)

			err = d.SetServices([]*gatt.Service{s, h})
			if err != nil {
				log.Fatalf("pkg bluetooth; could not add service: %s", err)
			}

			podIdArray, err := hex.DecodeString("fffffffe")
			if err != nil {
				log.Fatalf("pkg bluetooth; could not parse default address: %s", err)
			}

			if podId != nil {
				podIdArray = podId
			}

			// CE1F923D-C539-48EA-7300-0AFFFFFFFE00
			// Advertise device name and service's UUIDs.
			mfgData, _ := hex.DecodeString("60030001000000")
			err = d.AdvertiseNameServicesMfgData(
				"AP "+strings.ToUpper(hex.EncodeToString(podIdArray))+" 0A95B6110002761B",
				[]gatt.UUID{
					gatt.MustParseUUID("CE1F923D-C539-48EA-7300-0A" + hex.EncodeToString(podIdArray) + "00"),
				},
				mfgData,
			)
			if err != nil {
				log.Fatalf("pkg bluetooth; could not advertise: %s", err)
			}
		default:
		}
	}
	err = d.Init(onStateChanged)
	if err != nil {
		log.Fatalf("pkg bluetooth; could not init bluetooth: %s", err)
	}
	return b, nil
}

func (b *Ble) RefreshAdvertisingWithSpecifiedId(id []byte) error { // 4 bytes, first 2 usually empty
	log.Debugf("RefreshAdvertisingWithSpecifiedId %x", id)
	// Looking at the paypal/gatt source code, we don't need to call StopAdvertising,
	// but just call AdvertiseNameAndServices and it should update

	mfgData, _ := hex.DecodeString("60030001000000")
	err := (*b.device).AdvertiseNameServicesMfgData(
		"AP "+strings.ToUpper(hex.EncodeToString(id))+" 0A95B6110002761B",
		[]gatt.UUID{
			gatt.MustParseUUID("CE1F923D-C539-48EA-7300-0A" + hex.EncodeToString(id) + "00"),
		},
		mfgData,
	)
	if err != nil {
		log.Infof("pkg bluetooth; could not re-advertise: %s", err)
	}
	return err
}

func (b *Ble) WriteCmd(packet Packet) error {

	b.cmdOutput <- packet
	return nil
}

func (b *Ble) WriteData(packet Packet) error {
	b.dataOutput <- packet
	return nil
}

func (b *Ble) writeDataBuffer(buf *bytes.Buffer) error {
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	buf.Reset()
	return b.WriteData(data)
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

func (b *Ble) ReadMessage() (*message.Message, error) {
	message := <-b.messageInput
	return message, nil
}

func (b *Ble) ReadMessageWithTimeout(d time.Duration) (*message.Message, bool) {
	select {
	case message := <-b.messageInput:
		return message, false
	case <-time.After(d):
		log.Debugf("ReadMessage timeout")
		return nil, true
	}
}

func (b *Ble) ShutdownConnection() {
	(*b.central).Close()
}

func (b *Ble) WriteMessage(message *message.Message) {
	b.messageOutput <- message
}

func (b *Ble) loop(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case msg := <-b.messageOutput:
			b.writeMessageData(msg)
		case data := <-b.dataInput:
			msg, err := b.readMessageData(data)
			if err != nil {
				log.Fatalf("pkg bluetooth; error reading message: %s", err)
			}
			b.messageInput <- msg
		case cmd := <-b.cmdInput:
			msg, err := b.readMessage(cmd)
			if err != nil {
				log.Fatalf("pkg bluetooth; error reading message: %s", err)
			}
			if msg != nil {
				b.messageInput <- msg
			}
		}
	}
}

func (b *Ble) StartMessageLoop() {
	if b.stopLoop != nil {
		log.Fatalf("pkg bluetooth; Messaging loop is already running")
	}
	b.stopLoop = make(chan bool)
	go b.loop(b.stopLoop)
}

func (b *Ble) StopMessageLoop() {
	// race condition, but this is called only on device disconnect
	if b.stopLoop != nil {
		close(b.stopLoop)
		b.stopLoop = nil
	}
}

func (b *Ble) expectCommand(expected Packet) {
	cmd, _ := b.ReadCmd()
	if !bytes.Equal(expected[:1], cmd[:1]) {
		log.Fatalf("pkg bluetooth; expected command: %s. received command: %s", expected, cmd)
	}
}

func (b *Ble) writeMessageData(msg *message.Message) {
	var buf bytes.Buffer
	var index byte = 0

	bytes, err := msg.Marshal()
	if err != nil {
		log.Fatalf("pkg bluetooth; could not marshal the message %s", err)
	}
	log.Tracef("pkg bluetooth; Sending message: %x", bytes)

	sum := crc32.ChecksumIEEE(bytes)

	buf.WriteByte(index) // index
	buf.WriteByte(0)     // fragments

	buf.WriteByte(byte(sum >> 24))
	buf.WriteByte(byte(sum >> 16))
	buf.WriteByte(byte(sum >> 8))
	buf.WriteByte(byte(sum))
	buf.WriteByte((byte(len(bytes))))
	buf.Write(bytes[:])
	b.writeDataBuffer(&buf)
	return
}

func (b *Ble) writeMessage(msg *message.Message) {
	var buf bytes.Buffer
	var index byte = 0

	b.WriteCmd(CmdRTS)
	b.expectCommand(CmdCTS) // TODO figure out what to do if !CTS
	bytes, err := msg.Marshal()
	if err != nil {
		log.Fatalf("pkg bluetooth; could not marshal the message %s", err)
	}
	log.Tracef("pkg bluetooth; Sending message: %x", bytes)
	sum := crc32.ChecksumIEEE(bytes)
	if len(bytes) <= 18 {
		buf.WriteByte(index) // index
		buf.WriteByte(0)     // fragments

		buf.WriteByte(byte(sum >> 24))
		buf.WriteByte(byte(sum >> 16))
		buf.WriteByte(byte(sum >> 8))
		buf.WriteByte(byte(sum))
		buf.WriteByte((byte(len(bytes))))
		end := len(bytes)
		if len(bytes) > 14 {
			end = 14
		}
		buf.Write(bytes[:end])
		b.writeDataBuffer(&buf)

		if len(bytes) > 14 {
			buf.WriteByte(index)
			buf.WriteByte(byte(len(bytes) - 14))
			buf.Write(bytes[14:])
			b.writeDataBuffer(&buf)
		}
		return
	}

	size := len(bytes)
	fullFragments := (byte)((size - 18) / 19)
	rest := (byte)((size - (int(fullFragments) * 19)) - 18)
	buf.WriteByte(index)
	buf.WriteByte(fullFragments + 1)
	buf.Write(bytes[:18])

	b.writeDataBuffer(&buf)

	for index = 1; index <= fullFragments; index++ {
		buf.WriteByte(index)
		if index == 1 {
			buf.Write(bytes[18:37])
		} else {
			buf.Write(bytes[(index-1)*19+18 : (index-1)*19+18+19])
		}
		b.writeDataBuffer(&buf)
	}

	buf.WriteByte(index)
	buf.WriteByte(rest)
	buf.WriteByte(byte(sum >> 24))
	buf.WriteByte(byte(sum >> 16))
	buf.WriteByte(byte(sum >> 8))
	buf.WriteByte(byte(sum))
	end := rest
	if rest > 14 {
		end = 14
	}
	buf.Write(bytes[(fullFragments*19)+18 : (fullFragments*19)+18+end])
	b.writeDataBuffer(&buf)
	if rest > 14 {
		index++
		buf.WriteByte(index)
		buf.WriteByte(rest - 14)
		buf.Write(bytes[fullFragments*19+18+14:])
		for buf.Len() < 20 {
			buf.WriteByte(0)
		}
		b.writeDataBuffer(&buf)
	}
	b.expectCommand(CmdSuccess)
}

func (b *Ble) readMessageData(data Packet) (*message.Message, error) {
	var buf bytes.Buffer
	var checksum []byte

	// fragments := int(data[1])
	checksum = data[2:6]
	len := data[6]
	end := len + 7
	buf.Write(data[7:end])

	bytes := buf.Bytes()
	sum := crc32.ChecksumIEEE(bytes)
	if binary.BigEndian.Uint32(checksum) != sum {
		log.Warnf("pkg bluetooth; checksum missmatch. checksum is: %x. want: %x", sum, checksum)
		log.Warnf("pkg bluetooth; data: %s", hex.EncodeToString(bytes))

		b.WriteCmd(CmdFail)
		return nil, errors.New("checksum missmatch")
	}

	msg, _err := message.Unmarshal(bytes)

	if _err == nil {
		b.WriteCmd(CmdSuccess)
	} else {
		b.WriteCmd(CmdFail)
	}

	log.Tracef("pkg bluetooth; Received message:", spew.Sdump(msg))

	return msg, _err
}

func (b *Ble) readMessage(cmd Packet) (*message.Message, error) {
	var buf bytes.Buffer
	var checksum []byte

	if bytes.Equal(CmdSuccess[:1], cmd[:1]) {
		return nil, nil
	}

	log.Trace("pkg bluetooth; Reading RTS")
	if !bytes.Equal(CmdRTS[:1], cmd[:1]) {
		log.Fatalf("pkg bluetooth; expected command: %x. received command: %x", CmdRTS, cmd)
	}
	log.Trace("pkg bluetooth; Sending CTS")

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
			log.Warnf("pkg bluetooth; sending NACK, packet index is wrong")
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
	log.Tracef("pkg bluetooth; One extra: %t", oneExtra)
	if oneExtra {
		data, _ := b.ReadData()
		buf.Write(data[2 : data[1]+2])
	}
	bytes := buf.Bytes()
	sum := crc32.ChecksumIEEE(bytes)
	if binary.BigEndian.Uint32(checksum) != sum {
		log.Warnf("pkg bluetooth; checksum missmatch. checksum is: %x. want: %x", sum, checksum)
		log.Warnf("pkg bluetooth; data: %s", hex.EncodeToString(bytes))

		b.WriteCmd(CmdFail)
		return nil, errors.New("checksum missmatch")
	}

	b.WriteCmd(CmdSuccess)

	msg, _err := message.Unmarshal(bytes)
	log.Tracef("pkg bluetooth; Received message:", spew.Sdump(msg))

	return msg, _err
}
