// topic detail usecase

package detail

type NsqTopicDetailResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	EventTrigger string     `json:"event_trigger"`
	GroupOwner   string     `json:"group_owner"`
	Bookmarked   bool       `json:"bookmarked"`
	Permission   permission `json:"permission"`
	NsqdHosts    []string   `json:"nsqd_hosts"`
}

type permission struct {
	CanPause              bool `json:"can_pause"`
	CanPublish            bool `json:"can_publish"`
	CanTail               bool `json:"can_tail"`
	CanDelete             bool `json:"can_delete"`
	CanEmptyQueue         bool `json:"can_empty_queue"`
	CanUpdateEventTrigger bool `json:"can_update_event_trigger"`
}
