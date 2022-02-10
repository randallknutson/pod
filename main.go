package main

import (
	"flag"
	"time"

	"github.com/avereha/pod/pkg/api"
	"github.com/avereha/pod/pkg/bluetooth"
	"github.com/avereha/pod/pkg/pod"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	var stateFile = flag.String("state", "state.toml", "pod state")
	var freshState = flag.Bool("fresh", false, "start fresh. not activated, empty state")
	// if both verbose and quiet are chosen, e.g., -v -q, the verbose dominates
	var traceLevel = flag.Bool("v", false, "verbose off by default, TraceLevel")
	var infoLevel = flag.Bool("q", false, "quiet off by default, InfoLevel")

	flag.Parse()

	if *traceLevel {
		log.SetLevel(log.TraceLevel)
	} else if *infoLevel {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
		ForceColors:  true,
	})

	// TODO: This is kinda ugly, move state reader into own file and pass state to both BLE and pod
	state := &pod.PODState{
		Filename: *stateFile,
	}
	var err error
	if !(*freshState) {
		state, err = pod.NewState(*stateFile)
		if err != nil {
			log.Fatalf("pkg pod; could not restore pod state from %s: %+v", stateFile, err)
		}
	}

	log.Tracef("podId %@ %x", state.Id, state.Id)

	ble, err := bluetooth.New("hci0", state.Id)
	//defer ble.Close()
	if err != nil {
		log.Fatalf("Could not start BLE: %s", err)
	}

	p := pod.New(ble, *stateFile, *freshState)
	go func() {
		p.StartAcceptingCommands()
	}()

	log.Info("Starting API")
	s := api.New(p)
	s.Start()

	time.Sleep(9999 * time.Second)
}
