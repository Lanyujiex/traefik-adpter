package traefik

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func UseDynamic(config *rest.Config) {
	dynamicSet, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dynamicSet: %s\n", err.Error())
		os.Exit(1)
	}
	middSchema := schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "middlewares",
	}
	list, err := dynamicSet.Resource(middSchema).List(context.Background(), v1.ListOptions{})
	if err != nil {
		os.Exit(1)
	}
	for _, middleware := range list.Items {
		fmt.Println(middleware)
	}
}
