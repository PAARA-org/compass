# compass

This application reads bearings and plots them on a compass

In its current implementation, it reads 23 (or 25 if padding is used) character lines from STDIN and plots bearing circles on a compass.

## Building

```bash
go build server.go
```

Alternatively, you can set the **OS** and **ARCH** when compiling the code. With GO version 1.21.0, the following OS and ARCH combinations are allowed as seen in `go tool dist list`:

| OS \ ARCH | 386 | amd64 | arm | arm64 | loong64 | mips | mips64 | mips64le | mipsle | ppc64 | ppc64le | riscv64 | s390x | WASM |
| --------- | --- | ----- | --- | ----- | ------- | ---- | ------ | -------- | ------ | ----- | ------- | ------- | ----- | ---- |
| aix       |     |       |     |       |         |      |        |          |        | X     |         |         |       |      |
| android   | X   | X     | X   | X     |         |      |        |          |        |       |         |         |       |      |
| darwin    |     | X     |     | X     |         |      |        |          |        |       |         |         |       |      |
| dragonfly |     | X     |     |       |         |      |        |          |        |       |         |         |       |      |
| freebsd   | X   | X     | X   | X     |         |      |        |          |        |       |         | X       |       |      |
| illumos   |     | X     |     |       |         |      |        |          |        |       |         |         |       |      |
| ios       |     | X     |     | X     |         |      |        |          |        |       |         |         |       |      |
| js        |     |       |     |       |         |      |        |          |        |       |         |         |       | X    |
| linux     | X   | X     | X   | X     | X       | X    | X      | X        | X      | X     | X       | X       | X     |      |
| netbsd    | X   | X     | X   | X     |         |      |        |          |        |       |         |         |       |      |
| openbsd   | X   | X     | X   | X     |         |      |        |          |        |       |         |         |       |      |
| plan9     | X   | X     | X   |       |         |      |        |          |        |       |         |         |       |      |
| solaris   |     | X     |     |       |         |      |        |          |        |       |         |         |       |      |
| wasip1    |     |       |     |       |         |      |        |          |        |       |         |         |       | X    |
| windows   | X   | X     | X   | X     |         |      |        |          |        |       |         |         |       |      |

- Build for Raspberry PI (ARM) Running 32 bit Linux

  ```bash
  GOOS=linux GOARCH=arm go build server.go
  ```

- Build for RaspBerry PI (ARM) running 64 bit Linux

  ```bash
  GOOS=linux GOARCH=arm64 go build server.go
  ```

## Usage

```bash
% ./server --helpshort
flag provided but not defined: -helpshort
Usage of ./server:
  -bearings int
    	Max bearings to cache (default 20)
  -paddedTimestamp
    	Pad timestamps to 15 digits
  -refresh int
    	Refresh interval in seconds (default 5)
```

## Generating fake input data

13 digit millisecond unix timestamps:

`for i in $(seq -f "%03g" 0 359) ; do echo -n "C$i" ; echo -n "0000000" ; printf "%010d" "$(($(date +%s)))" ; echo "000" ; sleep 0.5 ; done  >> bearings`

15 digit millisecond unix timestamps:

`for i in $(seq -f "%03g" 0 359) ; do echo -n "C$i" ; echo -n "0000000" ; printf "%012d" "$(($(date +%s)))" ; echo "000" ; sleep 0.5 ; done  >> bearings`

## Example run

```bash
tail -F bearings | ./server -bearings=30 -refresh=1
```
