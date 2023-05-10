package main

import (
	"fmt"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// informer demo

// InformerDemo is demo of use informer
func InformerDemo() {
	// 创建conf，用于api server的连接
	// clientcmd.RecommendedHomeFile获取home下的.kube
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		panic(err)
	}

	// 创建client
	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		panic(err)
	}

	// 创建informer
	factory := informers.NewSharedInformerFactory(clientSet, 0)
	informer := factory.Core().V1().Pods().Informer()

	// 给informer添加处理事件
	/*
			三个属性，对应不同的事件，添加，更新，删除
			type ResourceEventHandlerFuncs struct {
			AddFunc    func(obj interface{})
			UpdateFunc func(oldObj, newObj interface{})
			DeleteFunc func(obj interface{})
		}
	*/
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("Add")
			fmt.Printf("%v", obj)
		},
		UpdateFunc: func(oldobj interface{}, newobj interface{}) {
			fmt.Println("Update")
			fmt.Printf("老配置%v", oldobj)
			fmt.Printf("新配置%v", newobj)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("Del")
			fmt.Printf("%v", obj)
		},
	})

	// 启动informer
	stopchan := make(chan struct{})
	factory.Start(stopchan)            // 里面最终调用的是infromer.Run
	factory.WaitForCacheSync(stopchan) //等待所有已启动的通知器的缓存被同步。
	<-stopchan                         // 等待结束
}
