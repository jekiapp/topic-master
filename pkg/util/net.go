package util

import (
	"net"
	"os"
)

// ReplaceDockerIPWithLocalhost is a helper function to replace the docker IP
// with the localhost IP when detected to run in local environment.
func ReplaceDockerIPWithLocalhost(address string) string {
	if os.Getenv("IN_LOCAL") == "" {
		return address
	}
	// Try to split host and port using net.SplitHostPort, which handles IPv6
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		// If error, it might be because there's no port, so treat the whole as host
		port = ""
	}
	return "127.0.0.1:" + port
}
