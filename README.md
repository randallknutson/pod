# pod

Fake pod implementation
The "packet" folder it's just bits and small pieces that I used trying to figure out the protocol with data from logs.

## How to build

```
go build
```

## How to run

```
./pod
```

## How to build&run for pi

```
GOARCH=arm go build; ssh pi 'killall pod'; scp pod pi:~/ && ssh pi " sudo setcap 'cap_net_raw,cap_net_admin=eip' ./pod; ./pod"
```
