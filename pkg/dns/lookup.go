package dns

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
)

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

func LookupIP(hostname string) ([]net.IP, error) {
	ips := make([]net.IP, 0)

	log.Debug().Msgf("looking up IP for hostname %s", hostname)
	if false && strings.Contains(hostname, "cluster.local") {
		if ClusterServer == nil {
			ResolverFromIP(ClusterServerIP)
		}

		log.Debug().Msgf("looking up IP for hostname %s in the cluster against DNS server %s", hostname, ClusterServerIP)
		_, iplist, err := ClusterServer.LookupSRV(context.Background(), "", "", hostname)
		if err != nil {
			return ips, err
		}

		for _, ip := range iplist {
			//ips = append(ips, ip)
			fmt.Printf("  - SRV Record %+#v\n", ip)
		}
	} else {
		iplist, err := net.LookupIP(hostname)
		if err != nil {
			return ips, err
		}

		ips = append(ips, iplist...)
	}

	return ips, nil
}

func PurgeIPv6(ips []net.IP) []net.IP {
	filtered := make([]net.IP, 0)

	for _, ip := range ips {
		if IsIPv4(ip.String()) {
			filtered = append(filtered, ip)
		}
	}

	return filtered
}
