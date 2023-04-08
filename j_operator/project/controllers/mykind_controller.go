/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mygroupv1beta1 "joperator/api/v1beta1"
)

// MyKindReconciler reconciles a MyKind object
type MyKindReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mygroup.mfx.co,resources=mykinds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.mfx.co,resources=mykinds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.mfx.co,resources=mykinds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyKind object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
// 消费者，主要代码逻辑写在这
func (r *MyKindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	myKind := mygroupv1beta1.MyKind{}                     // 自己定义的对象的kind
	err := r.Client.Get(ctx, req.NamespacedName, &myKind) //  r里面存储了对象的信息，这里用get获取对应key的内容，放到mykind中，NamespaceName是ns加上名称
	if err != nil {
		fmt.Println(err)
	}

	nodeList := v1.NodeList{}

	if myKind.Spec.Foo != "" {
		err = r.Client.List(ctx, &nodeList) // 获取nodeList
		if err != nil {
			fmt.Println(err)
		}
		// 遍历每个node，给每个node添加一个pod
		for _, item := range nodeList.Items {
			p := v1.Pod{
				TypeMeta: v12.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: v12.ObjectMeta{
					GenerateName: item.Name, // 命名前缀，如果没指定name，k8s会用generatename 加上一个随机数
					Namespace:    myKind.Namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: myKind.Spec.Foo,
							Name:  "container",
						},
					},
					NodeName: item.Name,
				},
			}

			err = r.Client.Create(ctx, &p)
			if err != nil {
				fmt.Println("create err", err)
			}

		}

	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// 生产者，相当于informer
func (r *MyKindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mygroupv1beta1.MyKind{}). // 箭筒mygroupv1beta1.Mykind对象的变化，最后交给Reconcile
		// 主处理逻辑
		Complete(r)
}
