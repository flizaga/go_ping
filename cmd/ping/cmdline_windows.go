//go:build windows
// +build windows

package main

import "os"

var usage_windows = `
Usage:

    ` + os.Args[0] + ` [-c count] [-i interval] [-t timeout] [--udp] [--debug] host

Examples:

    # ping google continuously
    ` + os.Args[0] + ` www.google.com

    # ping google 5 times
    ` + os.Args[0] + ` -c 5 www.google.com

    # Send an unprivileged UDP ping
    ` + os.Args[0] + ` --udp www.google.com

!WARNING!   Reading the TTL from received ipv4 packets is not yet implemented on GO :(
            -> https://github.com/golang/go/issues/7175
`

func GetOsUsage() string {
	return usage_windows
}

func GetDefaultPrivilegedFlag() bool {
	return false
}
