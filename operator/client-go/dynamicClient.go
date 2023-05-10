package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// DynamicClient 是
func DynamicClient() {
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(conf)

	if err != nil {
		panic(err.Error())
	}

	// 设置查询的资源
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "pods"}

	// 使用dynamicClient的查询列表方法，查询指定namespace下的所有pod，限定为100条
	unstructObj, err := dynamicClient.Resource(gvr).Namespace("default").Get(context.TODO(), "nginx-deployment-8d8c654d5-x4dbr", mv1.GetOptions{}, "")
	//List(context.TODO(), mv1.ListOptions{Limit: 100})

	if err != nil {
		panic(err)
	}

	// 通过runtime.DefaultUnstructuredConverter函数将unstructured.UnstructuredList转为PodList类型
	pod := &v1.Pod{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), pod)
	if err != nil {
		panic(err)
	}
	fmt.Println(pod.Name)

	// for _, d := range podList.Items {
	// 	fmt.Printf("NAMESPACE: %v NAME:%v \t STATUS: %+v\n", d.Namespace, d.Name, d.Status.Phase)
	// }

}
