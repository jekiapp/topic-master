package nsq

import (
	"fmt"

	nsqmodel "github.com/jekiapp/topic-master/internal/model/nsq"
	nsqrepo "github.com/jekiapp/topic-master/internal/repository/nsq"
	"github.com/jekiapp/topic-master/pkg/util"
)

func GetNsqdHosts(lookupdURL, topicName string) ([]nsqmodel.SimpleNsqd, error) {
	nsqds, err := nsqrepo.GetNsqdsForTopic(lookupdURL, topicName)
	if err != nil {
		return nil, fmt.Errorf("error getting nsqds for topic: %v", err)
	}

	hosts := make([]nsqmodel.SimpleNsqd, 0, len(nsqds))
	for _, n := range nsqds {

		host := fmt.Sprintf("%s:%d", n.BroadcastAddress, n.HTTPPort)
		hosts = append(hosts, nsqmodel.SimpleNsqd{
			Address:  util.ReplaceDockerIPWithLocalhost(host),
			HostName: n.Hostname,
		})
	}

	return hosts, nil
}
