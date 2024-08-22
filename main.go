package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Setup config for connecting to the Kubernetes cluster
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", home+"/.kube/config", "kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Watch Deployments in the default namespace
	watcher, err := clientset.AppsV1().Deployments("default").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Process watch events
	for event := range watcher.ResultChan() {
		if event.Type == watch.Modified {
			deployment, ok := event.Object.(*v1.Deployment)
			if !ok {
				fmt.Println("error decoding object")
				continue
			}

			// Clear node affinity settings
			if deployment.Spec.Template.Spec.Affinity != nil {
				deployment.Spec.Template.Spec.Affinity.NodeAffinity = nil
			}

			// Update the deployment
			updatedDeployment, err := clientset.AppsV1().Deployments("default").Update(context.TODO(), deployment, metav1.UpdateOptions{})
			if err != nil {
				fmt.Printf("Error updating deployment: %v\n", err)
				continue
			}

			fmt.Printf("Removed node affinity from deployment: %s\n", updatedDeployment.Name)
		}
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
