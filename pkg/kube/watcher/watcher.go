package watcher

import (
	"sync"
	"time"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/cli/watch/options"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/dns"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/entpoints"
	externalname "github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/service"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/reconcile"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

func New(options *options.Options) (*Watcher, error) {
	watcher := new(Watcher)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Info().Msgf("cluster kube config not found, trying local config: %s", options.Kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
		if err != nil {
			return nil, errors.Wrapf(err, "could not retreive kube config")
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retreive client set from kube config")
	}

	watcher.resyncInterval = options.ResyncInterval
	watcher.clientset = clientset

	watcher.externalNameChan = make(chan watch.Event)
	watcher.endpointChan = make(chan watch.Event)

	watcher.externalNameWatcher = externalname.New(watcher.clientset, watcher.externalNameChan, watcher.resyncInterval)
	watcher.endpointsWatcher = entpoints.New(watcher.clientset, watcher.endpointChan, watcher.resyncInterval)

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	watcher.broadcaster = broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "externalname-resolver"})

	return watcher, nil
}

type Watcher struct {
	kubeconfig          string
	clientset           *kubernetes.Clientset
	externalNameWatcher *externalname.Watcher
	endpointsWatcher    *entpoints.Watcher
	externalNameChan    chan watch.Event
	endpointChan        chan watch.Event
	resyncInterval      time.Duration
	broadcaster         record.EventRecorder
	wg                  sync.WaitGroup
}

func (i *Watcher) Watch() {
	go i.do()
	i.endpointsWatcher.Watch()
	i.externalNameWatcher.Watch()

	i.wg.Add(3)
	i.wg.Wait()
}

func (i *Watcher) do() {
	for {
		select {
		case endpointsEvent := <-i.endpointChan:
			endpoints := endpointsEvent.Object.(*corev1.Endpoints)

			if i.endpointsWatcher.Relevant(endpoints) {
				eventType := endpointsEvent.Type
				log.Debug().Msgf("event %s for endpoints %s", eventType, endpoints.GetName())
			}

		case externalNameEvent := <-i.externalNameChan:
			service := externalNameEvent.Object.(*corev1.Service)

			if i.externalNameWatcher.Relevant(service) {
				eventType := externalNameEvent.Type

				switch eventType {
				case watch.Added:
					log.Debug().Msgf("event %s for service %s", eventType, service.GetName())
					i.reconcile(service, eventType)
				case watch.Modified:
					log.Info().Msgf("event %s for service %s", eventType, service.GetName())
					i.reconcile(service, eventType)
				case watch.Deleted:
					log.Info().Msgf("event %s for service %s", eventType, service.GetName())
					i.delete(service, eventType)
				}
			}
		}
	}
}

func (i *Watcher) getClusterDnsServer() error {
	if options.Current.ClusterDnsIP != "" {
		dns.ClusterServerIP = options.Current.ClusterDnsIP
		return nil
	}

	services, err := i.externalNameWatcher.List()
	if err != nil {
		return errors.Wrapf(err, "failed to load service list")
	}

	clusterDnsIP, found := dns.GetClusterDnsServer(services)
	if !found {
		return errors.Wrapf(err, "failed to find cluster DNS server IP")
	}

	dns.ClusterServerIP = clusterDnsIP

	log.Info().Msgf("found cluster DNS server IP : %s", dns.ClusterServerIP)
	return nil
}

func (i *Watcher) delete(service *corev1.Service, event watch.EventType) {
	endpoints, err := i.endpointsWatcher.List()
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve endpoints list")
	}

	endpoint, found := reconcile.FindRelatedEndpoints(service, endpoints)
	if found {
		err = i.endpointsWatcher.Create(endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("failed create correlated Endpoints for %s.%s", service.GetNamespace(), service.GetName())
			i.broadcaster.Event(service, corev1.EventTypeWarning, err.Error(), "failed to delete correlated Endpoints")
			return
		}

		log.Debug().Msgf("endpoint deleted %s", endpoint.GetName())
		i.broadcaster.Event(service, corev1.EventTypeNormal, "", "correlated Endpoints deleted")
	}
}

func (i *Watcher) reconcile(externalName *corev1.Service, event watch.EventType) {
	log.Debug().Msgf("reconcile service %s for event %s", externalName.GetName(), event)

	if dns.ClusterServerIP == "" {
		err := i.getClusterDnsServer()
		if err != nil {
			log.Error().Err(err).Msg("failed to locate cluster DNS server")
			i.broadcaster.Event(externalName, corev1.EventTypeWarning, err.Error(), "failed to locate cluster DNS server")
			return
		}
	}

	modified := reconcile.EnsureAnnotations(externalName)

	if modified {
		err := i.externalNameWatcher.Update(externalName)
		if err != nil {
			log.Error().Err(err).Msgf("failed to update service %s.%s", externalName.GetNamespace(), externalName.GetName())
			i.broadcaster.Event(externalName, corev1.EventTypeWarning, err.Error(), "failed to ensure all annotations")
			return
		}

		log.Info().Msgf("successfully updated service %s.%s", externalName.GetNamespace(), externalName.GetName())
		i.broadcaster.Event(externalName, corev1.EventTypeNormal, "", "successfully updated service")
	}

	// ----
	// check correlated Endpoints
	endpoints, err := i.endpointsWatcher.List()
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve endpoints list")
	}

	endpoint, found := reconcile.FindRelatedEndpoints(externalName, endpoints)
	action := watch.Bookmark // nothing to do

	if found {
		log.Debug().Msgf("correlated endpoint %s found", endpoint.GetName())

		modified := reconcile.EnsureEndpointAnnotations(externalName, endpoint)
		ipChange, err := reconcile.CheckEndpointIps(endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("failed to check hostname IP change %s.%s", endpoint.GetNamespace(), endpoint.GetName())
			i.broadcaster.Event(endpoint, corev1.EventTypeWarning, err.Error(), "failed to check hostname IP change")
			return
		}

		if modified || ipChange {
			action = watch.Modified
		}

	} else {
		log.Debug().Msg("correlated endpoint does not exist")
		endpoint, err = reconcile.EndpointsFromExternalNameService(externalName)
		if err != nil {
			log.Error().Err(err).Msgf("failed generate correlated Endpoints for %s.%s", externalName.GetNamespace(), externalName.GetName())
			i.broadcaster.Event(externalName, corev1.EventTypeWarning, err.Error(), "failed to generate correlated Endpoints")
			return
		}

		log.Debug().Msgf("endpoint generated %s", endpoint.GetName())
		action = watch.Added
	}

	switch action {
	case watch.Added:
		err = i.endpointsWatcher.Create(endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("failed create correlated Endpoints for %s.%s", externalName.GetNamespace(), externalName.GetName())
			i.broadcaster.Event(externalName, corev1.EventTypeWarning, err.Error(), "failed to generate correlated Endpoints")
			return
		}

		log.Debug().Msgf("endpoint created %s", endpoint.GetName())
		i.broadcaster.Event(externalName, corev1.EventTypeNormal, "", "created correlated Endpoints")

	case watch.Modified:
		err := i.endpointsWatcher.Update(endpoint)
		if err != nil {
			log.Error().Err(err).Msgf("failed to update endpoints %s.%s", endpoint.GetNamespace(), endpoint.GetName())
			i.broadcaster.Event(endpoint, corev1.EventTypeWarning, err.Error(), "failed ensure all annotations")
			return
		}

		log.Info().Msgf("successfully updated endpoints %s.%s", endpoint.GetNamespace(), endpoint.GetName())
		i.broadcaster.Event(endpoint, corev1.EventTypeNormal, "", "successfully updated endpoints")
	}

}
