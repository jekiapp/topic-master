package nsq

// Nsqd represents an NSQD node as returned by lookupd
// Only include fields relevant for discovery/connection
// (add json tags for possible future marshaling)
type Nsqd struct {
	BroadcastAddress string `json:"broadcast_address"`
	Hostname         string `json:"hostname"`
	RemoteAddress    string `json:"remote_address"`
	TCPPort          int    `json:"tcp_port"`
	HTTPPort         int    `json:"http_port"`
	Version          string `json:"version"`
}
