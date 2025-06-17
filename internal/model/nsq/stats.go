package nsq

// Stats represents the aggregated stats for a topic and channel across nsqd nodes
// json tags are provided for marshaling
// Only relevant fields from stats.json are included

type Stats struct {
	TopicName    string    `json:"topic_name"`
	Depth        int       `json:"depth"`
	MessageCount int       `json:"message_count"`
	Channels     []Channel `json:"channels"`
	Paused       bool      `json:"paused"`
}

type Channel struct {
	ChannelName   string   `json:"channel_name"`
	Depth         int      `json:"depth"`
	InFlightCount int      `json:"in_flight_count"`
	DeferredCount int      `json:"deferred_count"`
	MessageCount  int      `json:"message_count"`
	RequeueCount  int      `json:"requeue_count"`
	ClientCount   int      `json:"client_count"`
	Clients       []Client `json:"clients"`
}

type Client struct {
	ClientID      string `json:"client_id"`
	Hostname      string `json:"hostname"`
	Version       string `json:"version"`
	RemoteAddress string `json:"remote_address"`
	State         int    `json:"state"`
	ReadyCount    int    `json:"ready_count"`
	InFlightCount int    `json:"in_flight_count"`
	MessageCount  int    `json:"message_count"`
	FinishCount   int    `json:"finish_count"`
	RequeueCount  int    `json:"requeue_count"`
	ConnectTS     int64  `json:"connect_ts"`
	UserAgent     string `json:"user_agent"`
}
