package entpoints

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/config"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/tools"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
)

func New(clientset *kubernetes.Clientset, listener chan<- watch.Event, resyncInterval time.Duration) *Watcher {
	instance := new(Watcher)

	instance.clientset = clientset
	instance.resyncInterval = resyncInterval
	instance.channel = listener

	return instance
}

type Watcher struct {
	clientset      *kubernetes.Clientset
	resyncInterval time.Duration
	channel        chan<- watch.Event
	wg             sync.WaitGroup
}

func (i *Watcher) Watch() {
	go i.do()
}

func (i *Watcher) do() {

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(60)
		return i.clientset.CoreV1().Endpoints(metav1.NamespaceAll).Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeOut})
	}

	watcher, err := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		panic(fmt.Sprintf("could not initialise endpoint watcher: %s\n", err.Error()))
	}

	for event := range watcher.ResultChan() {
		i.channel <- event
	}
}

func (i *Watcher) List() ([]*corev1.Endpoints, error) {
	list, err := i.clientset.CoreV1().Endpoints(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve enpoints")
	}

	flattened := make([]*corev1.Endpoints, 0)
	for idx, _ := range list.Items {
		flattened = append(flattened, &list.Items[idx])
	}

	return flattened, err
}

func (i *Watcher) FilterRelevant(items []*corev1.Endpoints) []*corev1.Endpoints {
	filtered := make([]*corev1.Endpoints, 0)

	for _, item := range items {
		if i.Relevant(item) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (i *Watcher) Relevant(item *corev1.Endpoints) bool {
	return tools.AnnotationExists(item.GetAnnotations(), config.AnnotationControllerManaged)
}

func (i *Watcher) Delete(item *corev1.Endpoints) error {
	return i.clientset.CoreV1().Endpoints(item.GetNamespace()).Delete(context.Background(), item.GetName(), *metav1.NewDeleteOptions(0))
}

func (i *Watcher) Create(item *corev1.Endpoints) error {
	_, err := i.clientset.CoreV1().Endpoints(item.GetNamespace()).Create(context.Background(), item, *&metav1.CreateOptions{})

	return err
}

func (i *Watcher) Update(item *corev1.Endpoints) error {
	_, err := i.clientset.CoreV1().Endpoints(item.GetNamespace()).Update(context.Background(), item, *&metav1.UpdateOptions{})

	return err
}
