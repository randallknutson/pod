# pod

Fake pod implementation
The "scripts" folder is just bits and small pieces that I used trying to figure out the protocol with data from logs.

## How to build

```
go build
```

building for Raspberry Pi:
```
GOARCH=arm go build`

```

## How to run

Tested so far only on Linux.
Before running, bring the BLE device down and stop bluetooth daemon
```
sudo hciconfig
sudo hciconfig hci0 down

sudo service bluetooth stop
```

Before running, the executable must be granted capabilities(or run as root):
```
sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod
```
And then run

```
./pod -fresh
```

## How to build&run for Raspberry pi

```
GOARCH=arm go build; ssh pi 'killall pod'; scp pod pi:~/ && ssh pi " sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod; ./pod"
```
