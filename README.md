# ip-discovery

Find other devices' ip through UDP broadcast.

## Usage

First run server on device A and B:

```sh
$ ip-discovery -s
Listening UDP at 0.0.0.0:25615
```

Then find A and B on device C:

```sh
$ ip-discovery
192.168.1.7
192.168.2.3
```

Help:

```sh
$ ip-discovery --help
Usage of ip-discovery:
  -k string
    	payload key
  -p int
    	server port (default 25615)
  -s	server mode
  -t int
    	timeout(ms) (default 1000)
```

## Installation

```sh
go install github.com/urie96/ip-discovery@latest
```
