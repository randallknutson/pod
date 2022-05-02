# pod

Fake pod implementation

* The original 0pen-dash repository from which this was forked was removed by the owner.

* This fork has diverged from the original implementation, which was based on hardcoded responses. This version mimics more pod details like reservoir level, total delivery, alerts, and faults, and builds dynamic responses based on that state.

* This version also mimics a behavior we see in some DASH pods where the pod disconnects every 3 minutes; this can be used with iOS hooks to make a heartbeat to run Loop in situations where a BLE CGM is not available. 

* It also has a websocket based API that can used by a separate [NodeJS/React frontend](https://github.com/ps2/pod_simulator_frontend), that is installed and run separately for now.

Requirements:
1. Version of iOS code (Loop app) that will interact with this simulator - Loop dev branch or FreeAPS freeaps_dev branch
2. Raspberry pi with Bluetooth BLE (using a pi3b or pi4 right now)
3. The user must have sudo privilege on the pi
4. Install the go language on your device (search internet for procedure)

You can build on pi directly or use cross-compiler and scp the executable

## Build on the pi

Log on the pi and type the following commands, starting at {your_path}/pod:
```
go build
sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod
```

## Run simulator on the pi

The simulator runs until:
* aborted with a control-C
* pod is deactivated on the phone
* quit out of phone app after establishing BLE connection

The simulator may error out unexpectedly. Just restart it and it should reconnect with the app (do not use the `-fresh` flag in this case.)

When in doubt, control-C and restart it.

To pair a new simulated dash pod, use this command on the pi:
```
./pod -q -fresh 
```

Timing for a new pod:
* start the pairing process in the app (by tapping on Pair Pod button)
* hit return on `./pod -q -fresh` command line on the pi
* if it fails to connect, wait for pairing attempt to fail and try again
* if it still fails, you can switch from Omnipod DASH on the app, add it back and try again

Working with an active pod simulator:
* If you need to quit and restart the app or to build the app fresh, it is best to control-C out of the pod simulator on the pi
* Restart or rebuild the app, then shortly after app opens, try to restore communication
* Otherwise, the app and simulator may stop being able to communicate
* In this case, you need to Deactivate Pod using app and add a new one

To restore communication between the app and an existing simulated dash pod, issue this command on the pi as soon as possible after resuming the app:
```
./pod -q
```

To change the reporting level, add one of these two flags:
* `-v` to make reporting more verbose (Trace Level)
* `-q` to make reporting less verbose (recommended)
* no extra flag - medium verbose (Debug Level)

Note that quitting the app will cause the following message:
```
FATA[####] pkg bluetooth; ** disconnect:
```

Simple restore communication as stated above.

# Original README.md

We maintained the original README file below. It may be helpful if someone plans to cross-compile the code and just transfer the executable.

The "scripts" folder contains bits and small pieces that I [original owner] used trying to figure out the protocol with data from logs.

## How to build

```
go build
```

building for Raspberry Pi:
```
GOARCH=arm go build`
```

## How to run

This was tested so far only on Linux.
Before running, bring the BLE device down and stop bluetooth daemon
```
sudo hciconfig
sudo hciconfig hci0 down

sudo service bluetooth stop
sudo  systemctl  disable bluetooth
```

Before running, the executable must be granted capabilities(or run as root):
```
sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod
```
And then run

```
./pod -fresh
```

## Flags

```
$ ./pod  --help
Usage of ./pod:
  -fresh
        start fresh. not activated, empty state
  -state string
        pod state (default "state.toml")

```

When running with `-fresh`, the state will be saved, so running it twice(first with `-fresh`, then without) should work.

## How to build & run for Raspberry pi
Tested on `Raspberry Pi 3B+` running `Raspbian 10`

```
GOARCH=arm go build;
ssh pi 'killall pod';
scp pod pi:~/  
ssh pi " sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod; ./pod"
```
