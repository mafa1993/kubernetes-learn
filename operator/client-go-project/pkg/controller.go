package pkg

// 用于实现informer的具体逻辑

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	kcorev1 "k8s.io/api/core/v1"
	knetv1 "k8s.io/api/networking/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/informers/core/v1"
	networkingv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	netlisterv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

/*Controller 是 informer 的实现
 * 1. 给service informer添加AddFunc和UpdateFunc
 * 2. 给ingress informer添加deletefunc
 * 3. 从ingress、service的indexer中获取数据
 * 4. 查询apiserver获取一些信息，用于处理
 */
type Controller struct {
	client        kubernetes.Interface
	svcLister     corelisterv1.ServiceLister // 用于从indexer中获取数据
	ingressLister netlisterv1.IngressLister  // 用于从indexer中获取数据
	queue         workqueue.RateLimitingInterface
}

const (
	workerNum int = 5
	maxRetry  int = 3
)

// 添加svc时的处理逻辑
func (con *Controller) svcAdd(obj interface{}) {
	con.addQueue(obj)
}

// 修改svc时的处理逻辑
func (con *Controller) svcEdit(old, new interface{}) {
	// 如果两个一致不做处理
	if reflect.DeepEqual(old, new) {
		return
	}

	// 如果是修改了annotation，则进行删除ingress，其他修改不做处理
	con.addQueue(new)
}

// ingress删除时的处理逻辑
func (con *Controller) ingressDel(obj interface{}) {
	// 如果ingress删除了，但是svc有annotation ingress/http=true 则重新创建ingress

	// ingress,ok := obj.(knetv1.ingress)
	// if !ok {
	// 	runtime.HandleError(errors.New("ingress 断言失败"))

	// }
	ingress, ok := obj.(*knetv1.Ingress) // 包里进行方法实现时，都是用的指针，所以这里断言指针，这里如果不为指针，下方传参传地址也可以

	if !ok {
		con.handlerError(ingress.Namespace+"/"+ingress.Name, errors.New("ingress 断言失败"))
		//runtime.HandleError(errors.New("ingress 断言失败"))
	}

	// 获取ownerReferences
	// 在 Kubernetes 中，使用ownerReferences来表示资源之间的从属关系。 级联删除：当删除根对象时，k8s垃圾回收器会自动删除从对象。
	svc := v13.GetControllerOf(ingress)

	if svc == nil {
		runtime.HandleError(errors.New("ingress 对应的svc未找到"))
	}

	if strings.ToLower(svc.Kind) != "service" {
		con.handlerError(ingress.Namespace+"/"+ingress.Name, errors.New("ingress对应的kind不是svc"))
		//runtime.HandleError(errors.New("ingress对应的kind不是svc"))
	}

	// 获取ingress对应的svc
	con.addQueue(ingress)

}

// 加入队列的方法
func (con *Controller) addQueue(obj interface{}) {
	var (
		key string
		err error
	)

	key, err = cache.MetaNamespaceKeyFunc(obj) // 获取key
	if err != nil {
		con.handlerError(key, err)
	}
	con.queue.Add(key) // 加入队列，也可以将key和obj都入队列
}

// 错误处理，有错进行重试
func (con *Controller) handlerError(key string, err error) {
	if con.queue.NumRequeues(key) <= maxRetry { // 小于重试次数
		con.queue.AddRateLimited(key) // 加入限速队列
		return
	}

	runtime.HandleError(err)
	con.queue.Forget(key) // 清空
}

// Run 是消费者，informer处理逻辑中加入了队列，定义方法对队列数据进行消费
func (con Controller) Run(stopCh chan struct{}) {

	// 控制worker数量
	for i := 0; i < workerNum; i++ {
		// Until loops until stop channel is closed, running f every period.
		go wait.Until(con.queueConsumer, 0, stopCh) // 里面的实现是一个死循环，不断的执行回调函数
	}

	<-stopCh // 用于标记程序结束
}

// 真正消费者
func (con Controller) queueConsumer() {
	item, shutDown := con.queue.Get() // 获取数据进行处理
	if shutDown {
		return
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(item.(string))

	if err != nil {
		runtime.HandleError(err)
	}

	// svc 新增、修改、删除
	// 获取svc
	svc, err := con.svcLister.Services(namespace).Get(name)
	// 获取ingress
	_, inerr := con.ingressLister.Ingresses(namespace).Get(name)

	if kerr.IsNotFound(err) && kerr.IsNotFound(inerr) {
		con.queue.Done(item) // 完成 在队列删除
		return
	}
	// 没有svc，在从svc中找annotations 会报错 Observed a panic: "invalid memory address or nil pointer dereference" (runtime error: invalid memory address or nil pointer dereference)

	if kerr.IsNotFound(err) {

		err2 := con.client.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, v13.DeleteOptions{}) // ingress 删除
		// 删除失败
		if err2 != nil {
			runtime.HandleError(err2)
		}
		con.queue.Done(item) // 完成 在队列删除
		return
	}
	annotation := svc.GetAnnotations()["ingress/http"] // 获取备注


	// 配置了true，并且没有创建ingress (包含两种情况，1是只建立了svc，还没建立ingress，2是删除ingress)
	if annotation == "true" && kerr.IsNotFound(inerr) {
		ingressOptions := con.createIngressOptions(svc)
		//创建ingress
		_, err2 := con.client.NetworkingV1().Ingresses(namespace).Create(context.TODO(), ingressOptions, v13.CreateOptions{})
		// 创建失败
		if err2 != nil {
			runtime.HandleError(err2)
		}

	}

	// 没找到svc，或者标记为false，但是有ingress，将ingress 删除
	if kerr.IsNotFound(err) && inerr == nil || annotation == "false" && inerr == nil {
		err2 := con.client.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, v13.DeleteOptions{}) // ingress 删除
		if err2 != nil {
			runtime.HandleError(err2)
		}
	}

	con.queue.Done(item) // 完成 在队列删除

}

// 创建ingress的配置
func (con *Controller) createIngressOptions(svc *kcorev1.Service) *knetv1.Ingress {
	var ingressOpts knetv1.Ingress
	ingressClass := "nginx"
	pathType := knetv1.PathTypeExact


	ingressOpts = knetv1.Ingress{
		// TypeMeta: v13.TypeMeta{  不需要定义typeMeta，用默认的即可
		// 	Kind: ,
		// },
		ObjectMeta: v13.ObjectMeta{
			Name:      svc.Name,      // 设置ingress name
			Namespace: svc.Namespace, // ingress namespace
			OwnerReferences: []v13.OwnerReference{ // 创建关联关系，删除ingress时，判断了svc，是从reference中找的
				*v13.NewControllerRef(svc, kcorev1.SchemeGroupVersion.WithKind("Service")),
				// *v13.NewControllerRef(svc, schema.GroupVersionKind{
				// 	Group:   svc.GroupVersionKind().Group,
				// 	Version: svc.GroupVersionKind().Version,
				// 	Kind:    svc.Kind,
				// }),
			},
		},
		Spec: knetv1.IngressSpec{
			IngressClassName: &ingressClass,
			Rules: []knetv1.IngressRule{
				{
					Host: "test.com", //访问的url
					IngressRuleValue: knetv1.IngressRuleValue{
						HTTP: &knetv1.HTTPIngressRuleValue{
							Paths: []knetv1.HTTPIngressPath{
								{
									Path:     "/" + svc.Name, //访问的路径
									PathType: &pathType,
									Backend: knetv1.IngressBackend{
										Service: &knetv1.IngressServiceBackend{
											Name: svc.Name, // 后端部署的service name
											Port: knetv1.ServiceBackendPort{
												Number: svc.Spec.Ports[0].Port, // 后端部署的service port
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingressOpts
}

// NewController 实例化controller
func NewController(client kubernetes.Interface, svcInformer v1.ServiceInformer, ingressInformer networkingv1.IngressInformer) Controller {
	var controllerObj Controller
	controllerObj = Controller{
		client:        client,
		svcLister:     svcInformer.Lister(),
		ingressLister: ingressInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "default"),
	}

	var (
		svcInformerObj, ingressInformerObj cache.SharedIndexInformer
	)
	// 获得service的informer进行处理事件添加
	svcInformerObj = svcInformer.Informer()
	svcInformerObj.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controllerObj.svcAdd,
		UpdateFunc: controllerObj.svcEdit,
	})

	ingressInformerObj = ingressInformer.Informer()
	// 监听ingress的删除逻辑
	ingressInformerObj.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controllerObj.ingressDel,
	})

	return controllerObj
}
