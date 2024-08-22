package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	fmt.Println("kubeconfig", *kubeconfig)
	fmt.Println("uid", os.Getuid())

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
		return
	}

	selector := fields.Everything()
	watchlist := cache.NewListWatchFromClient(clientset.AppsV1().RESTClient(), "deployments",
		metav1.NamespaceAll, selector)
	_, controller := cache.NewInformer(
		watchlist,
		&appsv1.Deployment{},
		0,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				fmt.Println("DeleteFunc")
			},
			AddFunc: func(newObj interface{}) {
				newDeployment := newObj.(*appsv1.Deployment)
				if hasNodeAffinity(&newDeployment.Spec.Template.Spec) {
					newDeploymentCopy := newDeployment.DeepCopy()
					newDeploymentCopy.Spec.Template.Spec.Affinity = nil
					_, err := clientset.AppsV1().Deployments(newDeployment.Namespace).Update(context.Background(), newDeploymentCopy, metav1.UpdateOptions{})
					if err != nil {
						log.Printf("Error updating deployment: %v", err)
					} else {
						log.Printf("Removed node affinity from deployment %s", newDeploymentCopy.Name)
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newDeployment := newObj.(*appsv1.Deployment)

				if hasNodeAffinity(&newDeployment.Spec.Template.Spec) {
					newDeploymentCopy := newDeployment.DeepCopy()
					newDeploymentCopy.Spec.Template.Spec.Affinity = nil
					_, err := clientset.AppsV1().Deployments(newDeployment.Namespace).Update(context.Background(), newDeploymentCopy, metav1.UpdateOptions{})
					if err != nil {
						log.Printf("Error updating deployment: %v", err)
					} else {
						log.Printf("Removed node affinity from deployment %s", newDeploymentCopy.Name)
					}
				}
			},
		},
	)

	stop := make(chan struct{})
	defer close(stop)

	go controller.Run(stop)
	if !cache.WaitForCacheSync(stop, controller.HasSynced) {
		log.Fatal("Timed out waiting for caches to sync")
	}

	<-stop
}

func hasNodeAffinity(spec *corev1.PodSpec) bool {
	return spec.Affinity != nil && spec.Affinity.NodeAffinity != nil
}
