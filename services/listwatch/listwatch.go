package listwatch

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"traefik-adpter/pkg/traefik"
	"traefik-adpter/utils"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	IngressPathStriKey   = "traefik.ingress.kubernetes.io/rule-type"
	PathPrefixStripValue = "PathPrefixStrip"
	IngressMiddleWareKey = "traefik.ingress.kubernetes.io/router.middlewares"
)

func ListIngress(stopCh chan struct{}) {
	config, err := ctrl.GetConfig()
	if err != nil {
		log.Printf("Error building in-cluster config: %s\n", err.Error())
		os.Exit(1)
	}
	// // 创建 Kubernetes 客户端配置
	// config, err = rest.InClusterConfig()
	// if err != nil {
	// 	log.Printf("Error building in-cluster config: %s\n", err.Error())
	// 	os.Exit(1)
	// }
	// traefik.UseClient(config)
	// traefik.UseDynamic(config)

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating clientset: %s\n", err.Error())
		os.Exit(1)
	}
	// 创建 Ingress 资源的 ListWatch
	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.NetworkingV1().Ingresses("").List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.NetworkingV1().Ingresses("").Watch(context.TODO(), options)
		},
	}

	// 设置监听器
	_, controller := cache.NewInformer(
		listWatch,
		&networkingv1.Ingress{},
		// 重试策略
		time.Second*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ingress := obj.(*networkingv1.Ingress)
				log.Printf("Ingress created: %s/%s, and the annotations: %v\n", ingress.Namespace, ingress.Name, utils.MapToJsonString(ingress.Annotations))
				PatchIngressRewriteRoot(clientset, ingress)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newIngress := newObj.(*networkingv1.Ingress)
				oldIngress := oldObj.(*networkingv1.Ingress)
				if reflect.DeepEqual(newIngress.Annotations, oldIngress.Annotations) {
					return
				}
				log.Printf("Ingress %s/%s, annotation is not equal, skipped", newIngress.Namespace, newIngress.Name)
				log.Printf("Ingress %s/%s, new annotation is %v", newIngress.Namespace, newIngress.Name, utils.MapToJsonString(newIngress.Annotations))
				log.Printf("Ingress %s/%s, old annotation is %v", oldIngress.Namespace, oldIngress.Name, utils.MapToJsonString(oldIngress.Annotations))
				PatchIngressRewriteRoot(clientset, newIngress)
			},
			DeleteFunc: func(obj interface{}) {},
		},
	)

	// 启动控制器
	go controller.Run(stopCh)
}

func PatchIngressRewriteRoot(clientset *kubernetes.Clientset, opIngress *networkingv1.Ingress) {
	newAnnotatins := opIngress.Annotations
	namespace := opIngress.Namespace
	if _, ok := newAnnotatins[IngressPathStriKey]; !ok {
		return
	}
	if newAnnotatins[IngressPathStriKey] == PathPrefixStripValue {
		middleValue := newAnnotatins[IngressMiddleWareKey]
		rewriteRoot := fmt.Sprintf("%s-rewrite-root@kubernetescrd", opIngress.Namespace)
		if strings.Contains(middleValue, rewriteRoot) {
			return
		}
		err := traefik.CreateRewriteMiddleware(namespace, "/", nil)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = clientset.NetworkingV1().Ingresses(namespace).Update(context.Background(), opIngress, metav1.UpdateOptions{})
		if err != nil {
			log.Println(err)
			return
		}
	}
}
