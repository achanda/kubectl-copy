package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var resourceName, namespace, newResourceName, resourceKind string
	flag.StringVar(&resourceName, "resource", "", "Name of the resource to copy")
	flag.StringVar(&namespace, "namespace", "", "Namespace of the resource")
	flag.StringVar(&newResourceName, "new-name", "", "New name for the copied resource")
	flag.StringVar(&resourceKind, "kind", "", "Kind of the resource (e.g., Pod, ConfigMap)")
	flag.Parse()

	if resourceName == "" || namespace == "" || newResourceName == "" || resourceKind == "" {
		slog.Error("missing required flags: --resource, --namespace, --new-name, --kind")
		os.Exit(1)
	}

	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		slog.Error("error building kubeconfig", err)
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		slog.Error("error in creating dynamic client", err)
		os.Exit(1)
	}

	gvr, _ := schema.ParseResourceArg(strings.ToLower(resourceKind) + "s")
	if gvr == nil {
		slog.Error("error parsing resource kind", err)
		os.Exit(1)
	}

	resourceClient := dynamicClient.Resource(*gvr).Namespace(namespace)
	originalResource, err := resourceClient.Get(context.TODO(), resourceName, metav1.GetOptions{})
	if err != nil {
		slog.Error("error getting original resource", err)
		os.Exit(1)
	}

	// Preparing the new resource
	newResource := originalResource.DeepCopy()
	unstructured.SetNestedField(newResource.Object, newResourceName, "metadata", "name")
	newResource.SetResourceVersion("")

	_, err = resourceClient.Create(context.TODO(), newResource, metav1.CreateOptions{})
	if err != nil {
		slog.Error("error creating new resource", err)
		os.Exit(1)
	}

	slog.Info("Copied %s %s to %s in namespace %s", resourceKind, resourceName, newResourceName, namespace)
}
