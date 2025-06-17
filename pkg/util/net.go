package util

import (
	"net"
	"strings"
)

// ReplaceDockerHostWithLocalhost replaces any non-IP host in the input slice with 127.0.0.1, keeping the port.
func ReplaceDockerIPWithLocalhost(ip string) string {
	// Try to split host and port using net.SplitHostPort, which handles IPv6
	host, port, err := net.SplitHostPort(ip)
	if err != nil {
		// If error, it might be because there's no port, so treat the whole as host
		host = ip
		port = ""
	}

	// Remove brackets for IPv6 if present
	trimmedHost := strings.Trim(host, "[]")

	if net.ParseIP(trimmedHost) == nil {
		if port != "" {
			return "127.0.0.1:" + port
		} else {
			return "127.0.0.1"
		}
	} else {
		if port != "" {
			// Reconstruct with brackets for IPv6 if needed
			if strings.Contains(trimmedHost, ":") {
				return "[" + trimmedHost + "]:" + port
			}
			return trimmedHost + ":" + port
		}
		return trimmedHost
	}
}
