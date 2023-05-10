package main

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//ClientSet 是client实现clientset的demo
func ClientSet() {
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(conf)

	if err != nil {
		panic(err)
	}

	// 根据group来查找，CoreV1为group， Pods指定ns
	podIn := clientSet.CoreV1().Pods("default")

	// 提供GET LIST  WATCH  DELETE 等方法
	pod, err := podIn.Get(context.TODO(), "nginx-deployment-8d8c654d5-x4dbr", v1.GetOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(pod.Name)
}
