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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	crdexamplecomv1 "kubebuilder/api/v1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=crd.example.com.crd.example.com,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.example.com.crd.example.com,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.example.com.crd.example.com,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
// controller业务逻辑的实现
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil  // result有两个属性，一个是是否重试，一个是间隔多久重试
}

// SetupWithManager sets up the controller with the Manager.
// 将controller注册到manager中
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).  // 创建builder
		For(&crdexamplecomv1.App{}). // 将自定义对象放入builder的ForInput属性，ownsInput管理 ownref 的关联子对象 Watches 用来设置其他关注的对象  withFilter是用来设置prdicate withoptions 用来设置controller的一些配置，例如worker数量等
		Complete(r)  // 这里面会调用Reconcile，里面对controller进行了初始化，并将controller add到了manager，controller.do 存放了Reconcile方法
}
// complete中包含两个主要过程，do controller 和 do watch
// do watch 中实现了事件的监听和处理  predicate用来过滤事件，controller start来开始事件的监听
// 事件的入队列，是用Request对象来存储的，包含name和namespace属性
