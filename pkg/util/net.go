package util

import (
	"net"
	"strings"
)

// ReplaceDockerHostWithLocalhost replaces any non-IP host in the input slice with 127.0.0.1, keeping the port.
func ReplaceDockerHostWithLocalhost(hosts []string) []string {
	result := make([]string, len(hosts))
	for i, h := range hosts {
		host, port := h, ""
		if idx := strings.LastIndex(h, ":"); idx > 0 {
			host = h[:idx]
			port = h[idx+1:]
		}
		if net.ParseIP(host) == nil {
			if port != "" {
				result[i] = "127.0.0.1:" + port
			} else {
				result[i] = "127.0.0.1"
			}
		} else {
			result[i] = h
		}
	}
	return result
}
