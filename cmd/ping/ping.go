package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-ping/ping"
)

type outputjson struct {
	Packets []ping.Packet
	Stats   *ping.Statistics
}

func main() {
	timeout := flag.Duration("t", time.Second*10, "Packet timeout.")
	interval := flag.Duration("i", time.Second, "Duration between two ping")
	count := flag.Int("c", -1, "Stop after x sent ping")
	size := flag.Int("s", 56, "Payload size")
	ttl := flag.Int("l", 64, "TTL")
	udp := flag.Bool("udp", GetDefaultPrivilegedFlag(), "Send UDP ping (Does not require admin privileges")
	debugmode := flag.Bool("debug", false, "Activate debug mode")
	printjson := flag.Bool("json", false, "Output JSON instead of normal ping")

	oldusage := flag.Usage
	flag.Usage = func() {
		oldusage()
		fmt.Print(GetOsUsage())
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	host := flag.Arg(0)
	pinger, err := ping.NewPinger(host)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	// listen for ctrl-C signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pinger.Stop()
		}
	}()

	var outputpacketsarray []ping.Packet
	pinger.OnRecv = func(pkt *ping.Packet) {
		if *printjson {
			outputpacketsarray = append(outputpacketsarray, *pkt)
		} else {
			fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%v time=%v\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Ttl, pkt.Rtt)
		}
	}
	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		if !*printjson {
			fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%v time=%v (DUP!)\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Ttl, pkt.Rtt)
		}
	}
	pinger.OnTimeout = func(pid int, pseq int) {
		inPkt := ping.Packet{
			Rtt:    *timeout,
			Ttl:    -1,
			ID:     pid,
			Seq:    pseq,
			IPAddr: pinger.IPAddr(),
			Addr:   pinger.Addr(),
		}
		if *printjson {
			outputpacketsarray = append(outputpacketsarray, inPkt)
		} else {
			fmt.Printf("Timeout! : icmp_seq=%d  ttl=%v time=>%v\n", inPkt.Seq, inPkt.Ttl, inPkt.Rtt)
		}
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		if *printjson {
			outputarray := &outputjson{
				Packets: outputpacketsarray,
				Stats:   stats,
			}
			jsonstring, _ := json.Marshal(outputarray)
			fmt.Printf("%s", jsonstring)
		} else {
			fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
			fmt.Printf("%d packets transmitted, %d packets received, %d duplicates, %v%% packet loss\n",
				stats.PacketsSent, stats.PacketsRecv, stats.PacketsRecvDuplicates, stats.PacketLoss)
			fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
				stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		}
	}

	pinger.Count = *count
	pinger.Size = *size
	pinger.Interval = *interval
	pinger.PktTimeout = *timeout
	pinger.TTL = *ttl
	pinger.SetPrivileged(!*udp)

	if *debugmode {
		fmt.Printf("Precision +/- %s\n", ping.GetTickDuration())
	}
	if !*printjson {
		fmt.Printf("PING %s (%s) %v(%v) bytes of data:\n", pinger.Addr(), pinger.IPAddr(), *size+8, pinger.Size)
	}
	err = pinger.Run()
	if err != nil {
		fmt.Println("Failed to ping target host:", err)
	}
}
