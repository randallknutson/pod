package pod

import (
	"bytes"

	"github.com/avereha/pod/pkg/bluetooth"
	log "github.com/sirupsen/logrus"
)

type podState int

const (
	podStateNotInitialized podState = iota
	podStateWithID
	podStateWithLTK
	podStateWithCK
)

type Pod struct {
	ble *bluetooth.Ble
}

func New(ble *bluetooth.Ble) *Pod {
	ret := &Pod{
		ble: ble,
	}
	return ret
}

func (p *Pod) expectCommand(expected bluetooth.Packet) {
	cmd, _ := p.ble.Read()
	if cmd.Type != expected.Type || bytes.Compare(cmd.Data, expected.Data) != 0 {
		log.Fatalf("Expected: %s. Received: %s", expected, cmd)
	}
}

func (p *Pod) StartActivation() {

	activationCmd, _ := p.ble.Read()
	log.Infof("Got activation command: %s", activationCmd)

	p.expectCommand(bluetooth.CmdRTS)
	p.ble.WriteCmd(bluetooth.CmdCTS)

	data1, _ := p.ble.Read()
	log.Infof("Got DATA1: %s", data1)
	data2, _ := p.ble.Read()
	log.Infof("Got DATA2: %s", data2)
	data3, _ := p.ble.Read()
	log.Infof("Got DATA3: %s", data3)

	log.Info("Sending SUCCESS")
	p.ble.WriteCmd(bluetooth.CmdSuccess)

	p.expectCommand(bluetooth.CmdRTS)
	p.ble.WriteCmd(bluetooth.CmdCTS)

	sp1_1, _ := p.ble.Read()
	log.Infof("Got SP1 1: %s", sp1_1)
	sp1_2, _ := p.ble.Read()
	log.Infof("Got SP1 2: %s", sp1_2)
	sp1_3, _ := p.ble.Read()
	log.Infof("Got SP1 3: %s", sp1_3)
	sp1_4, _ := p.ble.Read()
	log.Infof("Got SP1 4: %s", sp1_4)
	sp1_5, _ := p.ble.Read()
	log.Infof("Got SP1 5: %s", sp1_5)

	p.ble.WriteCmd(bluetooth.CmdSuccess)

	p.ble.WriteCmd(bluetooth.CmdRTS)
	p.expectCommand(bluetooth.CmdCTS)

	sp2_1 := bluetooth.Packet{
		Type: bluetooth.PacketTypeData,
		Data: sp1_1.Data, // THIS ID WRONG
	}
	log.Infof("Sending SP2 1: %s", sp2_1)
	p.ble.WriteData(sp2_1)

	sp2_2 := bluetooth.Packet{
		Type: bluetooth.PacketTypeData,
		Data: sp1_2.Data, // THIS ID WRONG
	}
	log.Infof("Sending SP2 2: %s ", sp2_2)
	p.ble.WriteData(sp2_2)

	sp2_3 := bluetooth.Packet{
		Type: bluetooth.PacketTypeData,
		Data: sp1_3.Data, // THIS ID WRONG
	}
	log.Infof("Sending SP2 3: %s", sp2_3)
	p.ble.WriteData(sp2_3)

	sp2_4 := bluetooth.Packet{
		Type: bluetooth.PacketTypeData,
		Data: sp1_4.Data, // THIS ID WRONG
	}
	log.Infof("Sending SP2 4: %s", sp2_4)
	p.ble.WriteData(sp2_4)
	sp2_5 := bluetooth.Packet{
		Type: bluetooth.PacketTypeData,
		Data: sp1_5.Data, // THIS ID WRONG
	}
	log.Infof("Sending SP2 5: %s", sp2_5)
	p.ble.WriteData(sp2_5)

	p.expectCommand(bluetooth.CmdSuccess)

	p.expectCommand(bluetooth.CmdRTS)
	p.ble.WriteCmd(bluetooth.CmdCTS)

	sp3_1, _ := p.ble.Read()
	log.Infof("Got SP3 1: %s", sp3_1)
	sp3_2, _ := p.ble.Read()
	log.Infof("Got SP3 2: %s", sp3_2)
	sp3_3, _ := p.ble.Read()
	log.Infof("Got SP3 3: %s", sp3_3)
	p.ble.WriteCmd(bluetooth.CmdSuccess)

	p.ble.WriteCmd(bluetooth.CmdRTS)
	p.expectCommand(bluetooth.CmdCTS)

	log.Infof("Sending SP4 1: %s", sp3_1)
	p.ble.WriteData(sp3_1) // WRONG
	log.Infof("Sending SP4 1: %s", sp3_2)
	p.ble.WriteData(sp3_2) // WRONG
	log.Infof("Sending SP4 1: %s", sp3_3)
	p.ble.WriteData(sp3_3) // WRONG

	p.expectCommand(bluetooth.CmdSuccess)

	p.expectCommand(bluetooth.CmdRTS)
	p.ble.WriteCmd(bluetooth.CmdCTS)

	final_1, _ := p.ble.Read()
	log.Infof("Got final 1: %s", final_1)
	final_2, _ := p.ble.Read()
	log.Infof("Got final 2: %s", final_2)

	p.ble.WriteCmd(bluetooth.CmdSuccess)

	p.ble.WriteCmd(bluetooth.CmdRTS)
	p.expectCommand(bluetooth.CmdCTS)

	log.Infof("Sending final 1: %s", final_1)
	p.ble.WriteData(final_1)
	log.Infof("Sending final 2: %s", final_1)
	p.ble.WriteData(final_2)

	// what we want to see in logs:
	//  EapAkaMasterModule Reached Eap-Aka SUCCESSFULLY
}
