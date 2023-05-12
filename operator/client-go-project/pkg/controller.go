package pkg

// 用于实现informer的具体逻辑

import (
	"errors"
	"reflect"
	"strings"

	knetv1 "k8s.io/api/networking/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
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
		runtime.HandleError(errors.New("ingress 断言失败"))
	}

	// 获取ownerReferences
	// 在 Kubernetes 中，使用ownerReferences来表示资源之间的从属关系。 级联删除：当删除根对象时，k8s垃圾回收器会自动删除从对象。
	svc := v13.GetControllerOf(ingress)
	if svc == nil {
		runtime.HandleError(errors.New("ingress 对应的svc未找到"))
	}

	if strings.ToLower(svc.Kind) != "service" {
		runtime.HandleError(errors.New("ingress对应的kind不是svc"))
	}
	// 获取ingress对应的svc
	con.addQueue(obj)

}

// 加入队列的方法
func (con *Controller) addQueue(obj interface{}) {
	var (
		key string
		err error
	)
	key, err = cache.MetaNamespaceKeyFunc(obj) // 获取key
	if err != nil {
		runtime.HandleError(err)
	}
	con.queue.Add(key) // 加入队列，也可以将key和obj都入队列
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
