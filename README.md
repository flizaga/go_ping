# go-ping
[![PkgGoDev](https://pkg.go.dev/badge/github.com/go-ping/ping)](https://pkg.go.dev/github.com/go-ping/ping)
[![Circle CI](https://circleci.com/gh/go-ping/ping.svg?style=svg)](https://circleci.com/gh/go-ping/ping)

A simple but powerful ICMP echo (ping) library for Go, inspired by
[go-fastping](https://github.com/tatsushid/go-fastping).

Here is a very simple example that sends and receives three packets:

```go
pinger, err := ping.NewPinger("www.google.com")
if err != nil {
	panic(err)
}
pinger.Count = 3
err = pinger.Run() // Blocks until finished.
if err != nil {
	panic(err)
}
stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
```

Here is an example that emulates the traditional UNIX ping command:

```go
pinger, err := ping.NewPinger("www.google.com")
if err != nil {
	panic(err)
}

// Listen for Ctrl-C.
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt)
go func() {
	for _ = range c {
		pinger.Stop()
	}
}()

pinger.OnRecv = func(pkt *ping.Packet) {
	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
}

pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
}

pinger.OnFinish = func(stats *ping.Statistics) {
	fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
	fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
}

fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
err = pinger.Run()
if err != nil {
	panic(err)
}
```

It sends ICMP Echo Request packet(s) and waits for an Echo Reply in
response. If it receives a response, it calls the `OnRecv` callback
unless a packet with that sequence number has already been received,
in which case it calls the `OnDuplicateRecv` callback. When it's
finished, it calls the `OnFinish` callback.

For a full ping example, see
[cmd/ping/ping.go](https://github.com/go-ping/ping/blob/master/cmd/ping/ping.go).

## Installation

```
go get -u github.com/go-ping/ping
```

To install the native Go ping executable:

```bash
go get -u github.com/go-ping/ping/...
$GOPATH/bin/ping
```

## Supported Operating Systems

### Linux
This library attempts to send an "unprivileged" ping via UDP. On Linux,
this must be enabled with the following sysctl command:

```
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

If you do not wish to do this, you can call `pinger.SetPrivileged(true)`
in your code and then use setcap on your binary to allow it to bind to
raw sockets (or just run it as root):

```
setcap cap_net_raw=+ep /path/to/your/compiled/binary
```

See [this blog](https://sturmflut.github.io/linux/ubuntu/2015/01/17/unprivileged-icmp-sockets-on-linux/)
and the Go [x/net/icmp](https://godoc.org/golang.org/x/net/icmp) package
for more details.

### Windows

You must use `pinger.SetPrivileged(true)`, otherwise you will receive
the following error:

```
socket: The requested protocol has not been configured into the system, or no implementation for it exists.
```

Despite the method name, this should work without the need to elevate
privileges and has been tested on Windows 10. Please note that accessing
packet TTL values is not supported due to limitations in the Go
x/net/ipv4 and x/net/ipv6 packages.

### Plan 9 from Bell Labs

There is no support for Plan 9. This is because the entire `x/net/ipv4`
and `x/net/ipv6` packages are not implemented by the Go programming
language.

## Maintainers and Getting Help:

This repo was originally in the personal account of
[sparrc](https://github.com/sparrc), but is now maintained by the
[go-ping organization](https://github.com/go-ping).

For support and help, you usually find us in the #go-ping channel of
Gophers Slack. See https://invite.slack.golangbridge.org/ for an invite
to the Gophers Slack org.

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/go-ping/ping/blob/master/CONTRIBUTING.md)

## Flizaga changes

- Changed the way we get the ping latency
	- The current go_ping project stores a date in the ping packet data, then compare this date to the current reception date
	- On Linux/Unix, this is working, but it's not optimal, because if the clock changed in the meantime, 
		for time adjustment via ntp or daylight time change, it can give an inaccurate latency. 
		So I changed on unix/linux/macos to store a time object on the program side instead of the packet. 
		On GOlang level, the time obj stores two time values : a date/timestamp and a monotonic clock value. 
		When you use a substraction in order to get a duration, then Golang will use the monotonic clock values, 
		so it's not affected by any modification on system date and time. 
	- On Windows, the system clock has a very bad precision (approx 18ms precision) and monotonic clock is also no very accurate (1ms). 
		So the best way is to use performance counters, it has higher resolution. Accuracy is 1 CPU Cycle which would be 3ns for a 3Ghz CPU. 
		As far as I know it's a bit more resources consuming, and we get a little overhead to do the kernel32 syscall, but it's a lot better 
		than the other clocks. Anyway we are not supposed to send thousand of ping per seconds, so it's not a big issue on the resources usage. 
		https://learn.microsoft.com/en-us/windows/win32/sysinfo/acquiring-high-resolution-time-stamps 
- Added a JSON output, which helps to use it on scripts rather than parsing the current output 
- Changed the default human reable output for it to look like more than the standard RedHat ping command which comes from iputils package. 
	It could be interesting to add a parameter to choose how the output would look like (windows, gnu ping, iputils ping, etc.) 
- Changed the status of the privileged flag to true by defaut when running on windows 
- Added a packet timeout value, which is more interesting than a whole command timeout 
	In order to do this, I store a time object in the awaitingSequences object instead of an empty struct. 
	As I've added also a packet timeout, we are supposed to have a not so big awaitingSequences object as it is cleaned at each packet timeout OnTimeout event 
- Improved the usage display on the cmd/ping command 
- Changed a bit the parameters from the command 

P.S. : this is the first time I do some GO coding, so please don't be mad at me for the code quality
