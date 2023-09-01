package reconcile

import (
	"net"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/dns"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/config"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/tools"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

func EnsureEndpointAnnotations(service *corev1.Service, endpoints *corev1.Endpoints) bool {
	modified := false

	if !tools.AnnotationExistsAndIsEqual(endpoints.GetAnnotations(), config.AnnotationControllerManaged, config.AnnotationControllerManagedValue) {
		endpoints.Annotations[config.AnnotationControllerManaged] = config.AnnotationControllerManagedValue
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationControllerManaged, endpoints.GetNamespace(), endpoints.GetName())
		modified = true
	}

	ports := service.Annotations[config.AnnotationExternalPorts]
	protocols := service.Annotations[config.AnnotationExternalProtocols]
	hostname := service.Annotations[config.AnnotationExternalHostname]

	if !tools.AnnotationExistsAndIsEqual(endpoints.GetAnnotations(), config.AnnotationExternalPorts, ports) {
		endpoints.Annotations[config.AnnotationExternalPorts] = ports
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationExternalPorts, endpoints.GetNamespace(), endpoints.GetName())
		modified = true
	}

	if !tools.AnnotationExistsAndIsEqual(endpoints.GetAnnotations(), config.AnnotationExternalProtocols, protocols) {
		endpoints.Annotations[config.AnnotationExternalProtocols] = protocols
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationExternalProtocols, endpoints.GetNamespace(), endpoints.GetName())
		modified = true
	}

	if !tools.AnnotationExistsAndIsEqual(endpoints.GetAnnotations(), config.AnnotationExternalHostname, hostname) {
		endpoints.Annotations[config.AnnotationExternalHostname] = hostname
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationExternalHostname, endpoints.GetNamespace(), endpoints.GetName())
		modified = true
	}

	return modified
}

func CheckEndpointIps(endpoints *corev1.Endpoints) (bool, error) {
	modified := false

	ips, err := dns.LookupIP(endpoints.Annotations[config.AnnotationExternalHostname])
	if err != nil {
		return modified, errors.Wrapf(err, "failed to lookup IP for endpoint \"%s\" hostname \"%s\"", endpoints.GetName(), endpoints.Annotations[config.AnnotationExternalHostname])
	}

	if !ipBlockEqual(endpoints.Subsets[0].Addresses, ips) {
		endpoints.Subsets[0].Addresses = endpointAddressesFromIps(ips)
		modified = true
	}

	return modified, nil
}

func ipBlockEqual(addresses []corev1.EndpointAddress, ips []net.IP) bool {
	for _, address := range addresses {
		found := false

		for _, ip := range ips {
			if address.IP == ip.String() {
				found = true
			}
		}

		if !found {
			log.Debug().Msgf("adress %s not found in ip list", address.IP)
			return false
		}
	}

	return len(addresses) == len(ips)
}
