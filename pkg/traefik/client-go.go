package traefik

import (
	"context"
	"fmt"
	"log"
	"os"

	traefikType "github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikClient "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	v1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func UseClient(config *rest.Config) {
	log.Println("use client")
	traefikClient, err := traefikClient.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating traefikClient: %s\n", err.Error())
		os.Exit(1)
	}
	middles, err := traefikClient.TraefikV1alpha1().Middlewares("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		os.Exit(1)
	}
	for _, middleware := range middles.Items {
		fmt.Println(middleware)
	}
}

func CreateRewriteMiddleware(namespace string, path string, config *rest.Config) (err error) {
	if config == nil {
		config, err = ctrl.GetConfig()
		if err != nil {
			log.Printf("Error building in-cluster config: %s\n", err.Error())
			return err
		}
	}
	if namespace == "" {
		return fmt.Errorf("Error: Create default rewrite-root Middleware failed, namespace cannot be empty!")
	}
	mid := v1alpha1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rewrite-root",
			Namespace: namespace,
		},
		Spec: v1alpha1.MiddlewareSpec{
			ReplacePath: &traefikType.ReplacePath{
				Path: path,
			},
		},
	}
	traefikClient, err := traefikClient.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating traefikClient: %s\n", err.Error())
		return err
	}
	_, err = traefikClient.TraefikV1alpha1().Middlewares(namespace).Create(context.Background(), &mid, metav1.CreateOptions{})
	if err != nil && errors.IsAlreadyExists(err) {
		log.Printf("Error creatingrewrite-root Middleware failed: %s\n", err.Error())
		return err
	}
	return nil
}
