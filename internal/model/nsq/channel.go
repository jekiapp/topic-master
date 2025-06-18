package nsq

// ChannelStats represents the statistics for an NSQ channel
type ChannelStats struct {
	Depth    int  `json:"depth"`
	Messages int  `json:"messages"`
	InFlight int  `json:"in_flight"`
	Requeued int  `json:"requeued"`
	Deferred int  `json:"deferred"`
	Paused   bool `json:"paused"`
}
