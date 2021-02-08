# pod
Fake pod

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
GOARCH=arm go build; ssh pi 'killall pod'; scp pod pi:~/ && ssh pi ./pod
```
