package pkg

// 用于实现informer的具体逻辑

import (
	v1 "k8s.io/client-go/informers/core/v1"
	networkingv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	netlisterv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
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
}

// 添加svc时的处理逻辑
func (con Controller) svcAdd(obj interface{}) {

}

// 修改svc时的处理逻辑
func (con Controller) svcEdit(old, new interface{}) {

}

// ingress删除时的处理逻辑
func (con Controller) ingressDel(obj interface{}) {

}

// Run 是消费者，informer处理逻辑中加入了队列，定义方法对队列数据进行消费
func (con Controller) Run() {

}

// NewController 实例化controller
func NewController(client kubernetes.Interface, svcInformer v1.ServiceInformer, ingressInformer networkingv1.IngressInformer) Controller {
	var controllerObj Controller
	controllerObj = Controller{
		client:        client,
		svcLister:     svcInformer.Lister(),
		ingressLister: ingressInformer.Lister(),
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
