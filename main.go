package main

import (
	"flag"
	"time"

	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/avereha/pod/pkg/pod"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	var stateFile = flag.String("state", "state.toml", "pod state")
	var freshState = flag.Bool("fresh", false, "start fresh. not activated, empty state")

	flag.Parse()


	log.SetLevel(log.TraceLevel)

	log.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
		ForceColors:  true,
	})

	ble, err := bluetooth.New("hci0")
	//defer ble.Close()
	if err != nil {
		log.Fatalf("Could not start BLE: %s", err)
	}

	p := pod.New(ble, *stateFile, *freshState)
	p.StartAcceptingCommands()

	time.Sleep(9999 * time.Second)
}
