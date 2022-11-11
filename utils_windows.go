//go:build windows
// +build windows

package ping

import (
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	modkernel32                = syscall.NewLazyDLL("kernel32.dll")
	_QueryPerformanceFrequency = modkernel32.NewProc("QueryPerformanceFrequency")
	_QueryPerformanceCounter   = modkernel32.NewProc("QueryPerformanceCounter")
	cpufreq                    = QPCFrequency()
)

// Returns an int64 of the number of ticks since start using queryPerformanceCounter
func QPC() uint64 {
	var now uint64
	r1, _, _ := syscall.SyscallN(_QueryPerformanceCounter.Addr(), uintptr(unsafe.Pointer(&now)))
	if r1 == 0 {
		panic("call failed")
	}
	return now
}

// QPCFrequency returns frequency in ticks per second
func QPCFrequency() uint64 {
	var freq uint64
	r1, _, _ := syscall.SyscallN(_QueryPerformanceFrequency.Addr(), uintptr(unsafe.Pointer(&freq)))
	if r1 == 0 {
		panic("call failed")
	}
	return freq
}

// Get duration of a CPU cycle (tick)
func GetTickDuration() time.Duration {
	out, _ := time.ParseDuration(strconv.FormatUint(uint64(time.Second.Nanoseconds())/QPCFrequency(), 10) + "ns")
	return out
}

// Returns the length of an ICMP message, plus the IP packet header.
func (p *Pinger) getMessageLength() int {
	if p.ipv4 {
		return p.Size + 8 + ipv4.HeaderLen
	}
	return p.Size + 8 + ipv6.HeaderLen
}

// Attempts to match the ID of an ICMP packet.
func (p *Pinger) matchID(ID int) bool {
	return ID == p.id
}

func QpcToBytes(nticks uint64) []byte {
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nticks >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func bytesToQpc(b []byte) uint64 {
	var nticks uint64
	for i := uint8(0); i < 8; i++ {
		nticks += uint64(b[i]) << ((7 - i) * 8)
	}
	return nticks
}

func BytesToTimestamp(data []byte) uint64 {
	return bytesToQpc(data)
}

func TimestampToBytes() []byte {
	return QpcToBytes(currentTimestamp())
}

func currentTimestamp() uint64 {
	return QPC() * uint64(time.Second.Nanoseconds()) / cpufreq
}
