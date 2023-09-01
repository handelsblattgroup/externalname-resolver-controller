package config

var (
	AnnotationControllerManaged      = "k8s.externalname.endpoints/managed"
	AnnotationControllerManagedValue = "true"
	AnnotationExternalHostname       = "k8s.externalname.endpoints/hostname"
	AnnotationExternalProtocols      = "k8s.externalname.endpoints/protocols"
	AnnotationExternalPorts          = "k8s.externalname.endpoints/ports"
)
