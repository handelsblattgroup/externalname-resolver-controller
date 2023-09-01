package reconcile

import (
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/config"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/tools"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

func EnsureAnnotations(service *corev1.Service) bool {
	modified := false

	if !tools.AnnotationExistsAndIsEqual(service.GetAnnotations(), config.AnnotationControllerManaged, config.AnnotationControllerManagedValue) {
		service.Annotations[config.AnnotationControllerManaged] = config.AnnotationControllerManagedValue
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationControllerManaged, service.GetNamespace(), service.GetName())
		modified = true
	}

	ports, protocols := FlatternPortsAndProtocols(service.Spec.Ports)

	if !tools.AnnotationExistsAndIsEqual(service.GetAnnotations(), config.AnnotationExternalPorts, ports) {
		service.Annotations[config.AnnotationExternalPorts] = ports
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationExternalPorts, service.GetNamespace(), service.GetName())
		modified = true
	}

	if !tools.AnnotationExistsAndIsEqual(service.GetAnnotations(), config.AnnotationExternalProtocols, protocols) {
		service.Annotations[config.AnnotationExternalProtocols] = protocols
		log.Debug().Msgf("updated %s annotation to %s.%s", config.AnnotationExternalProtocols, service.GetNamespace(), service.GetName())
		modified = true
	}

	return modified
}

func FindRelatedEndpoints(externalName *corev1.Service, endpoints []*corev1.Endpoints) (*corev1.Endpoints, bool) {
	for _, enpoint := range endpoints {
		if externalName.GetName() == enpoint.GetName() && externalName.GetNamespace() == enpoint.GetNamespace() {
			return enpoint, true
		}
	}

	return nil, false
}

func FilterOutOfSync() {

}
