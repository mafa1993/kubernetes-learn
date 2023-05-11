package main

import (
	"fmt"
	"log"

	"client-go/project/pkg"

	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	networkingv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		conf      *rest.Config
		err       error
		clientSet *kubernetes.Clientset
	)
	// 获取配置，第一个参数为空，则使用配置文件里配置的地址
	conf, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		// 使用集群内部的配置获取，通过serviceAccount
		conf, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

	// 创建client，这里是都是内部资源，所以使用了client set
	clientSet, err = kubernetes.NewForConfig(conf)

	if err != nil {
		log.Fatal(err)
	}

	var (
		shardeFactory   informers.SharedInformerFactory
		svcInformer     v1.ServiceInformer
		ingressInformer networkingv1.IngressInformer
	)
	// 创建informer
	shardeFactory = informers.NewSharedInformerFactoryWithOptions(clientSet, 0, informers.WithNamespace("default"))

	// 创建svc的informer
	svcInformer = shardeFactory.Core().V1().Services()
	// 创建ingress的informer
	ingressInformer = shardeFactory.Networking().V1().Ingresses()
	fmt.Println(svcInformer)
	fmt.Println(ingressInformer)

	// 增加事件
	var con pkg.Controller
	con = pkg.NewController(clientSet, svcInformer, ingressInformer)

	// 启动informer
	var stopCh chan struct{}
	// controller启动
	stopCh = make(chan struct{})

	con.Run()

	shardeFactory.Start(stopCh) // 定义的所有infromer都会Run
	shardeFactory.WaitForCacheSync(stopCh)

	<-stopCh
}
