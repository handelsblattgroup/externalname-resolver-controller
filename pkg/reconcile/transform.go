package reconcile

import (
	"net"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/cli/watch/options"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/dns"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/strings/slices"
)

func ClusterIPFromExternalNameService(item *corev1.Service) *corev1.Service {

	transformed := item.DeepCopy()

	transformed.Annotations[config.AnnotationControllerManaged] = config.AnnotationControllerManagedValue
	transformed.Annotations[config.AnnotationExternalHostname] = transformed.Spec.ExternalName

	ports, protocols := FlatternPortsAndProtocols(transformed.Spec.Ports)

	transformed.Annotations[config.AnnotationExternalPorts] = ports
	transformed.Annotations[config.AnnotationExternalProtocols] = protocols
	transformed.Spec.Type = corev1.ServiceTypeClusterIP
	transformed.Spec.ExternalName = ""

	return transformed
}

func EndpointsFromExternalNameService(service *corev1.Service) (*corev1.Endpoints, error) {
	endpoints := new(corev1.Endpoints)

	endpoints.Name = service.GetName()
	endpoints.Namespace = service.GetNamespace()
	endpoints.Annotations = service.GetAnnotations()
	endpoints.Labels = filterLabels(service.GetLabels())

	endpoints.Annotations[config.AnnotationControllerManaged] = config.AnnotationControllerManagedValue

	hostname := ""
	if value, exists := service.GetAnnotations()[config.AnnotationExternalHostname]; exists {
		hostname = value
	}

	endpoints.Annotations[config.AnnotationExternalHostname] = hostname
	endpoints.Annotations[config.AnnotationExternalPorts] = service.Annotations[config.AnnotationExternalPorts]
	endpoints.Annotations[config.AnnotationExternalProtocols] = service.Annotations[config.AnnotationExternalProtocols]

	ports, err := ExpendPortsAndProtocols(
		endpoints.Annotations[config.AnnotationExternalPorts],
		endpoints.Annotations[config.AnnotationExternalProtocols],
	)
	if err != nil {
		return endpoints, errors.Wrapf(err, "failed to expend endpoinds ports from annotations for \"%s\"", service.GetName())
	}

	subset := corev1.EndpointSubset{
		Ports:     ports,
		Addresses: []corev1.EndpointAddress{},
	}

	ips, err := dns.LookupIP(hostname)
	if err != nil {
		return endpoints, errors.Wrapf(err, "failed to lookup IP for endpoint \"%s\" hostname \"%s\"", service.GetName(), hostname)
	}

	subset.Addresses = endpointAddressesFromIps(ips)

	endpoints.Subsets = []corev1.EndpointSubset{
		subset,
	}

	return endpoints, nil
}

func filterLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)

	for key, value := range labels {
		if !slices.Contains(options.Current.IgnoreLabels, key) {
			filtered[key] = value
		}
	}

	return filtered
}

func endpointAddressesFromIps(ips []net.IP) []corev1.EndpointAddress {
	list := make([]corev1.EndpointAddress, 0)

	for _, ip := range ips {
		adresse := corev1.EndpointAddress{
			IP: ip.String(),
		}

		list = append(list, adresse)
	}

	return list
}
