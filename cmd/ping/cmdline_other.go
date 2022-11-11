//go:build !windows
// +build !windows

package main

var usage_other = `
Usage:

    ` + os.Args[0] + ` [-c count] [-i interval] [-t timeout] [--udp] [--debug] host

Examples:

    # ping google continuously
    sudo ` + os.Args[0] + ` www.google.com

    # ping google 5 times
    sudo ` + os.Args[0] + ` -c 5 www.google.com

    # Send an unprivileged UDP ping
    ` + os.Args[0] + ` --udp www.google.com
`

func GetOsUsage() string {
	return usage_other
}

func GetDefaultPrivilegedFlag() bool {
	return true
}
