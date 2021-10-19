package subscription

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type NamespaceSubscription struct {
	watcherInterface watch.Interface
	ClientSet        kubernetes.Interface
	Ctx              context.Context
	Completion       chan bool
}

func (ns *NamespaceSubscription) Reconcile(object runtime.Object, event watch.EventType) {

	// klog.Infof("%v", object)
	namespace := object.(*v1.Namespace)
	klog.Infof("NamespaceSubscription event type %s for %s", event, namespace.Name)

	switch event {
	case watch.Added:
		updatedNs := namespace.DeepCopy()
		if updatedNs.Labels == nil {
			updatedNs.Labels = make(map[string]string)
		}
		updatedNs.Labels["kubernetes.io/managed-by"] = "operator"
		klog.Info("Taking ownership of the 'custom' namespace.")
		_, err := ns.ClientSet.CoreV1().Namespaces().Update(context.TODO(), updatedNs, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		}

	case watch.Deleted:
		deletedNs := namespace.DeepCopy()
		klog.Info("Do you think you can delete the 'custom' namespace? THINK TWICE!")
		deletedNs.ResourceVersion = ""
		_, err := ns.ClientSet.CoreV1().Namespaces().Create(ns.Ctx, deletedNs, metav1.CreateOptions{})
		if err != nil {
			klog.Error(err)
			break
		}
		klog.Info("'custom' namespace created.")
	case watch.Modified:

		if namespace.Labels["type"] == "sre" {
			klog.Info("This could be some custom behaviour beyond just a CRUD")
		}

	}

}

func (ns *NamespaceSubscription) Subscribe() (watch.Interface, error) {
	var err error

	_, err = ns.ClientSet.CoreV1().Namespaces().Get(ns.Ctx, "custom", metav1.GetOptions{})
	if err != nil {
		klog.Infof("%v", err)
		customNs := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "custom",
				Labels: map[string]string{
					"kubernetes.io/managed-by": "operator",
				},
			},
		}
		_, err = ns.ClientSet.CoreV1().Namespaces().Create(ns.Ctx, customNs, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		klog.Info("'custom' namespace created.")
	} else {
		klog.Info("'custom' namespace already exists.")
	}

	ns.watcherInterface, err = ns.ClientSet.CoreV1().Namespaces().Watch(ns.Ctx, metav1.ListOptions{FieldSelector: "metadata.name=custom"})
	if err != nil {
		return nil, err
	}

	return ns.watcherInterface, nil
}

func (ns *NamespaceSubscription) IsComplete() <-chan bool {

	return ns.Completion
}
