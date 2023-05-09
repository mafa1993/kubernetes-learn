package main

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes/scheme"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//RestClient 是执行RESTClientFor
func RestClient() {
	// 创建config
	//根据config文件或者masterUrl构建一个Config结构体，该结构体包含了，集群的主要配置信息，原来除了通过配置文件的方式连接k8s集群外，还可以通过masterurl连接集群，官方比较推荐的是通过config文件连接集群
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)

	if err != nil {
		panic(err)
	}
	conf.GroupVersion = &v1.SchemeGroupVersion
	conf.NegotiatedSerializer = scheme.Codecs.WithoutConversion() // 设置编解码器 NegotiatedSerializer is used for obtaining encoders and decoders for multiple  supported media types
	//panic: NegotiatedSerializer is required when initializing a RESTClient  不设置报错

	conf.APIPath = "/api" // url的path
	//fmt.Println(conf)

	// 根据conf创建client
	restClient, err := rest.RESTClientFor(conf)
	if err != nil {
		panic(err)
	}

	// 发送请求到api server，获取数据
	// 发送get请求，获取default 命名空间下，pod资源
	rlt := restClient.Get().Namespace("default").Resource("pods").Name("nginx-deployment-8d8c654d5-x4dbr").Do(context.TODO())

	var pod v1.Pod
	pod = v1.Pod{}
	err = rlt.Into(&pod) // 将获取到的结果放入pod
	if err != nil {
		panic(err) //"nginx" not found
	}
	fmt.Println(pod.Name)
}
