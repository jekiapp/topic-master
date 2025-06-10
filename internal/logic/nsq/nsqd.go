package nsq

import (
	"fmt"

	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
)

func GetNsqdHosts(lookupdURL, topicName string) ([]string, error) {
	nsqds, err := nsqrepo.GetNsqdsForTopic(lookupdURL, topicName)
	if err != nil {
		return nil, fmt.Errorf("error getting nsqds for topic: %v", err)
	}

	hosts := make([]string, 0, len(nsqds))
	for _, n := range nsqds {
		hosts = append(hosts, fmt.Sprintf("%s:%d", n.BroadcastAddress, n.HTTPPort))
	}
	return util.ReplaceDockerHostWithLocalhost(hosts), nil
}
