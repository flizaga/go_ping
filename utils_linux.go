//go:build linux
// +build linux

package ping

var (
	basetime  = time.Now()
)

// Returns the length of an ICMP message.
func (p *Pinger) getMessageLength() int {
	return p.Size + 8
}

// Attempts to match the ID of an ICMP packet.
func (p *Pinger) matchID(ID int) bool {
	// On Linux we can only match ID if we are privileged.
	if p.protocol == "icmp" {
		if ID != p.id {
			return false
		}
	}
	return true
}

func durationToBytes(d time.Duration) []byte {
	nsec := d.Nanoseconds()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func bytesToDuration(b []byte) time.Duration {
	var nsec int64
	var t time.Duration
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	nstring := strconv.FormatInt(nsec, 10) + "ns"
	t, _ = time.ParseDuration(nstring)
	return t
}

func BytesToTimestamp(data []byte) uint64 {
	return uint64(bytesToDuration(data).Nanoseconds())
}

func TimestampToBytes() []byte {
	return durationToBytes(time.Since(basetime).Nanoseconds()))
}

func currentTimestamp() uint64 {
	return uint64(time.Since(basetime).Nanoseconds())
}
