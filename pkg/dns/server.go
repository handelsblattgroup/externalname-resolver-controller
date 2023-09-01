package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

var ClusterServerIP = ""
var ClusterServer *net.Resolver

func ResolverFromIP(ip string) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(2000),
			}
			return d.DialContext(ctx, network, fmt.Sprintf("%s:53", ip))
		},
	}

	ClusterServer = r
	log.Debug().Msgf("created cluster resolver from DNS server IP %s => %+#v", ClusterServerIP, ClusterServer)
}

func GetClusterDnsServer(services []*corev1.Service) (string, bool) {
	for _, service := range services {
		if hasDnsServerPorts(service) {
			return service.Spec.ClusterIP, true
		}
	}

	return "", false
}

func hasDnsServerPorts(service *corev1.Service) bool {
	for _, entry := range service.Spec.Ports {
		if entry.Port == int32(53) {
			return true
		}
	}

	return false
}
