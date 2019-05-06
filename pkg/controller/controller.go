package controller

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/walk1ng/k8swatch/pkg/utils"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/Sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/walk1ng/k8swatch/pkg/config"
	"github.com/walk1ng/k8swatch/pkg/handlers"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const maxRetries = 5

var serverStartTime time.Time

// Event describe the informer event
type Event struct {
	key          string
	eventType    string
	resourceType string
	namespace    string
}

// Controller object
type Controller struct {
	logger       *logrus.Entry
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler handlers.Handler
}

// Start starts controller entry
func Start(conf *config.Config, eventHandler handlers.Handler) {
	var clientset kubernetes.Interface
	_, err := rest.InClusterConfig()
	if err == rest.ErrNotInCluster {
		// run out of cluster
		clientset = utils.GetClientOutOfCluster()
	} else {
		// run in cluster
		clientset = utils.GetClient()
	}

	if conf.Resource.Pod {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return clientset.CoreV1().Pods(conf.Namespace).List(options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return clientset.CoreV1().Pods(conf.Namespace).Watch(options)
				},
			},
			&api_v1.Pod{},
			0,
			cache.Indexers{},
		)

		c := newController(clientset, informer, eventHandler, "pod")
		stopCh := make(chan struct{})
		defer close(stopCh)

		go c.Run(stopCh)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT)
	signal.Notify(sigterm, syscall.SIGTERM)
	<-sigterm

}

func newController(clientset kubernetes.Interface, informer cache.SharedIndexInformer, eventHandler handlers.Handler, resourceType string) *Controller {
	// queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	var newEvent Event
	var err error

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.eventType = "create"
			newEvent.resourceType = resourceType
			logrus.WithField("pkg", "k8swatch-"+resourceType).Infof("Processing add to %s: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(old)
			newEvent.eventType = "update"
			newEvent.resourceType = resourceType
			logrus.WithField("pkg", "k8swatch-"+resourceType).Infof("Processing update to %s: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newEvent.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			newEvent.eventType = "delete"
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(obj).Namespace
			logrus.WithField("pkg", "k8swatch-"+resourceType).Infof("Processing delete to %s: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
	})

	return &Controller{
		logger:       logrus.WithField("pkg", "k8swatch-"+resourceType),
		clientset:    clientset,
		queue:        queue,
		informer:     informer,
		eventHandler: eventHandler,
	}
}

// Run runs the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting k8swatch controller")
	serverStartTime = time.Now().Local()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.hasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timeout waiting for caches to sync"))
		return
	}

	c.logger.Info("k8swatch controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)

}

func (c *Controller) hasSynced() bool {
	return c.informer.HasSynced()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	newEvent, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(newEvent)
	err := c.processItem(newEvent.(Event))
	if err == nil {
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < maxRetries {
		c.logger.Errorf("Error processing %s (will retry): %v", newEvent.(Event).key, err)
		c.queue.AddRateLimited(newEvent)
	} else {
		// too many retries and err != nil
		c.logger.Errorf("Error processing %s (give up): %v", newEvent.(Event).key, err)
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(newEvent Event) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(newEvent.key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", newEvent.key, err)
	}

	// get object's metadata
	objMeta := utils.GetObjectMetaData(obj)

	// process event based on its type
	switch newEvent.eventType {
	case "create":
		if objMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			c.eventHandler.ObjCreated(obj)
			return nil
		}

	}
	return nil
}
