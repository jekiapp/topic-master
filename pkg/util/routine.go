package util

import "sync"

// ParallelForEachHost runs fn(host, topic, channel) in parallel for each host in hosts.
// Returns a slice of errors (nil if success, error if failed for that host).
func ParallelForEachHost(hosts []string, topic, channel string, fn func(host, topic, channel string) error) []error {
	var wg sync.WaitGroup
	errs := make([]error, len(hosts))
	for i, host := range hosts {
		wg.Add(1)
		go func(idx int, h string) {
			defer wg.Done()
			errs[idx] = fn(h, topic, channel)
		}(i, host)
	}
	wg.Wait()
	return errs
}
