/*
* Copyright (c) 2018, salesforce.com, inc.
* All rights reserved.
* SPDX-License-Identifier: BSD-3-Clause
* For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */
package main

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta2"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	sentinalv1alpha1 "github.com/karlhungus/sentinal-controller/pkg/apis/sentinalcontroller/v1alpha1"
	clientset "github.com/karlhungus/sentinal-controller/pkg/client/clientset/versioned"
	sentinalscheme "github.com/karlhungus/sentinal-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/karlhungus/sentinal-controller/pkg/client/informers/externalversions"
	listers "github.com/karlhungus/sentinal-controller/pkg/client/listers/sentinalcontroller/v1alpha1"
)

const controllerAgentName = "sentinal-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a SentinalDeployment is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a SentinalDeployment fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by SentinalDeployment"
	// MessageResourceSynced is the message used for an Event fired when a SentinalDeployment
	// is synced successfully
	MessageResourceSynced = "SentinalDeploy synced successfully"
)

// Controller is the controller implementation for SentinalDeploy resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sentinalclientset is a clientset for our own API group
	sentinalclientset clientset.Interface

	deploymentsLister         appslisters.DeploymentLister
	deploymentsSynced         cache.InformerSynced
	sentinalDeploymentsLister listers.SentinalDeploymentLister
	sentinalDeploymentsSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sentinal controller
func NewController(
	kubeclientset kubernetes.Interface,
	sentinalclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sentinalInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and SentinalDeployment
	// types.
	deploymentInformer := kubeInformerFactory.Apps().V1beta2().Deployments()
	sentinalInformer := sentinalInformerFactory.Sentinalcontroller().V1alpha1().SentinalDeployments()

	// Create event broadcaster
	// Add sentinal-controller types to the default Kubernetes Scheme so Events can be
	// logged for sentinal-controller types.
	sentinalscheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:             kubeclientset,
		sentinalclientset:         sentinalclientset,
		deploymentsLister:         deploymentInformer.Lister(),
		deploymentsSynced:         deploymentInformer.Informer().HasSynced,
		sentinalDeploymentsLister: sentinalInformer.Lister(),
		sentinalDeploymentsSynced: sentinalInformer.Informer().HasSynced,
		workqueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SentinalDeployments"),
		recorder:                  recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when SentinalDeployment resources change
	sentinalInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSentinalDeployment,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueSentinalDeployment(new)
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a SentinalDeployment resource will enqueue that SentinalDeployment resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1beta2.Deployment)
			oldDepl := old.(*appsv1beta2.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting SentinalDeployment controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.sentinalDeploymentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process SentinalDeployment resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// SentinalDeployment resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the SentinalDeployment resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the SentinalDeployment resource with this namespace/name
	sd, err := c.sentinalDeploymentsLister.SentinalDeployments(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("sentinal deployment '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := sd.Spec.StableDeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: stableDeployment name must be specified", key))
		return nil
	}

	canaryName := sd.Spec.CanaryDeploymentName
	if canaryName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: canaryDeployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in SentinalDeployment.spec
	deployment, err := c.deploymentsLister.Deployments(sd.Namespace).Get(deploymentName)
	// We couldn't find the deployment, so error out
	// TODO does this just take care of it the next time?
	// TODO should deployments be marked as owned before we take them over, so two sentinals can't exist at the same
	// time?
	if err != nil {
		return err
	}

	// Get the canary deployment with the name specified in SentinalDeployment.spec
	canary, err := c.deploymentsLister.Deployments(sd.Namespace).Get(canaryName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		canary, err = c.kubeclientset.AppsV1beta2().Deployments(sd.Namespace).Create(newCanaryDeployment(sd, deployment))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// TODO
	// If the Canary Deployment is not controlled by this SentinalDeployment resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(canary, sd) {
		msg := fmt.Sprintf(MessageResourceExists, canary.Name)
		c.recorder.Event(sd, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of replicas on the deployment resource is specified and not
	// equal to the desired replicas update the deployment with the stableDesiredReplcas
	if sd.Spec.StableDesiredReplicas != *deployment.Spec.Replicas {
		glog.V(4).Infof("SentinalDeployer Dep replicas: %d, deployR: %d, SentinalDeployer Canary replicas:%d, canaryR: %d", sd.Spec.StableDesiredReplicas, *deployment.Spec.Replicas, sd.Spec.CanaryDesiredReplicas, *canary.Spec.Replicas)

		// TODO handle errors
		c.updateReplicas(sd.Spec.StableDesiredReplicas, deployment)
		c.updateReplicas(sd.Spec.CanaryDesiredReplicas, canary)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the SentinalDeployment resource to reflect the
	// current state of the world
	err = c.updateSentinalDeploymentStatus(sd, deployment, canary)
	if err != nil {
		return err
	}

	c.recorder.Event(sd, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateSentinalDeploymentStatus(sd *sentinalv1alpha1.SentinalDeployment, deployment *appsv1beta2.Deployment, canary *appsv1beta2.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	sdCopy := sd.DeepCopy()
	sdCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	sdCopy.Status.CanaryReplicas = canary.Status.AvailableReplicas
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.sentinalclientset.SentinalcontrollerV1alpha1().SentinalDeployments(sd.Namespace).Update(sdCopy)
	return err
}

// enqueueSentinalDeployment takes a SentinalDeployment resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than SentinalDeployment.
func (c *Controller) enqueueSentinalDeployment(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the SentinalDeployment resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that SentinalDeployment resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a SentinalDeployment, we should not do anything more
		// with it.
		if ownerRef.Kind != "SentinalDeployment" {
			return
		}

		sd, err := c.sentinalDeploymentsLister.SentinalDeployments(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of sentinal deployment '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueSentinalDeployment(sd)
		return
	}
}

// updateReplicas brings the target closer to desired
func (c *Controller) updateReplicas(desired int32, target *appsv1beta2.Deployment) error {
	deployment := target.DeepCopy()
	//TODO HPS or nil Replicas
	//TODO only update replicas if deployment status avaiable matches current metrics
	if desired > *deployment.Spec.Replicas {
		newReplicas := *target.Spec.Replicas + 1
		deployment.Spec.Replicas = &newReplicas
	} else if desired < *deployment.Spec.Replicas {
		newReplicas := *target.Spec.Replicas - 1
		deployment.Spec.Replicas = &newReplicas
	}
	_, err := c.kubeclientset.AppsV1beta2().Deployments(deployment.Namespace).Update(deployment)
	if err != nil {
		//TODO handle this
		glog.V(4).Infof("failed to update deployment: %s replicas, %d", deployment.ObjectMeta.Name, *deployment.Spec.Replicas)
		return err
	}
	return nil
}

// newCanaryDeployment creates a new CanaryDeployment for a SentinalDeployment resource.
// It also sets the appropriate OwnerReferences on the resource so handleObject can discover the SD resource that 'owns' it.
func newCanaryDeployment(sd *sentinalv1alpha1.SentinalDeployment, dep *appsv1beta2.Deployment) *appsv1beta2.Deployment {
	canary := dep.DeepCopy()
	canary.ObjectMeta.ResourceVersion = ""
	canary.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(sd, schema.GroupVersionKind{
			Group:   sentinalv1alpha1.SchemeGroupVersion.Group,
			Version: sentinalv1alpha1.SchemeGroupVersion.Version,
			Kind:    "SentinalDeployment",
		}),
	}

	//TODO label this
	canary.ObjectMeta.Name = sd.Spec.CanaryDeploymentName
	//TODO calculate this based on how many dep's been decremented by
	canary.Spec.Replicas = &sd.Spec.CanaryDesiredReplicas
	for _, canaryContainer := range canary.Spec.Template.Spec.Containers {
		for _, updatedContainer := range sd.Spec.CanaryContainers {
			if canaryContainer.Name == updatedContainer.Name {
				canaryContainer.Image = updatedContainer.Image
			}
		}
	}
	return canary
}
