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
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	crdv1 "kubebuilder/api/v1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	tmpDir    string = "internal/template/" // 定义模板文件的路径
	tmpSuffix string = ".yaml"              // 设置模板文件的后缀，防止乱用
)

var (
	logger logr.Logger
)

//+kubebuilder:rbac:groups=crd.example.com,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.example.com,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.example.com,resources=apps/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

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
	logger = log.FromContext(ctx)
	var appt crdv1.App
	appt = crdv1.App{}
	// 从缓存中获取App的定义
	err := r.Get(ctx, req.NamespacedName, &appt)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // 忽略未找到错误
	}

	// deployment 处理
	// 1 获取deployment，看看是否存在
	var deployment *appv1.Deployment
	deployment = &appv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, deployment)

	// 没有找到deployment，则建立
	if errors.IsNotFound(err) {
		err1 := parseTemp("deployment", appt, deployment) // 对模板解析

		if err1 != nil {
			logger.Error(err1, "deployment解析出错")
			return ctrl.Result{}, err1
		}
		// 验证获取的dep是不是和app关联的
		// SetControllerReference里面验证了deployment的ref是不是appt，如果没有就设置上，如果不是就返回err
		if err_o := controllerutil.SetControllerReference(&appt, deployment, r.Scheme); err_o != nil {
			// 无关联报错
			return ctrl.Result{}, err_o
		}
		err = r.Create(ctx, deployment) // 创建
		if err != nil {
			logger.Error(err, "创建dep出错")
			return ctrl.Result{}, err // 重试
		}
	}

	// 没错，说明dep创建了，更新下
	if err == nil {
		err = r.Update(ctx, deployment) // 更新
		if err != nil {
			logger.Error(err, "更新dep出错")
			return ctrl.Result{}, err // 重试
		}	
	}

	// service 和ingress建立过程和deployment一致，可以抽成方法，这里暂不简化
	var svc *corev1.Service
	svc = &corev1.Service{}
	err = r.Get(ctx, req.NamespacedName, svc)

	err1 := parseTemp("service", appt, svc) // 对模板解析
	if err1 != nil {
		logger.Error(err1, "svc解析出错")
		return ctrl.Result{}, err1
	}

	// 没有找到svc，则建立
	if errors.IsNotFound(err) {
		if erro := controllerutil.SetControllerReference(&appt, svc, r.Scheme); erro != nil {
			// 无关联报错
			return ctrl.Result{Requeue: false}, erro
		}
		err = r.Create(ctx, svc) // 创建
		if err != nil {
			logger.Error(err, "创建svc出错")
			return ctrl.Result{}, err // 重试
		}
	}

	// 没错，说明svc创建了，更新下
	if err == nil {
		if !appt.Spec.EnableService {
			if err := r.Delete(ctx, svc); err != nil {
				return ctrl.Result{}, err
			}
		}else{
			err = r.Update(ctx, svc) // 更新
			if err != nil {
				logger.Error(err, "更新svc出错")
				return ctrl.Result{}, err // 重试
			}
		}
	}

	var ingress *netv1.Ingress
	ingress = &netv1.Ingress{}
	err = r.Get(ctx, req.NamespacedName, ingress)

	err1 = parseTemp("ingress", appt, ingress) // 对模板解析
	if err1 != nil {
		logger.Error(err1, "in解析出错")
		return ctrl.Result{}, err1
	}

	// 没有找到ingress，则建立
	if errors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(&appt, ingress, r.Scheme); err != nil {
			// 无关联报错
			return ctrl.Result{}, err
		}
		err = r.Create(ctx, ingress) // 创建
		if err != nil {
			logger.Error(err, "创建ingress出错")
			return ctrl.Result{}, err // 重试
		}
	}

	// 没错，说明ingress创建了，更新下
	if err == nil {
		if !appt.Spec.EnableIngress {
			if err := r.Delete(ctx, ingress); err != nil {
				return ctrl.Result{}, err
			}
		}else{
			err = r.Update(ctx, ingress) // 更新
			if err != nil {
				logger.Error(err, "更新ingress出错")
				return ctrl.Result{}, err // 重试
			}
		}
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil // result有两个属性，一个是是否重试，一个是间隔多久重试
}

// SetupWithManager sets up the controller with the Manager.
// 将controller注册到manager中
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr). // 创建builder
							For(&crdv1.App{}).         // 将自定义对象放入builder的ForInput属性，ownsInput管理 ownref 的关联子对象 Watches 用来设置其他关注的对象  withFilter是用来设置prdicate withoptions 用来设置controller的一些配置，例如worker数量等
							Owns(&appv1.Deployment{}). // 增加监听对象
							Owns(&netv1.Ingress{}).
							Owns(&corev1.Service{}).
							Complete(r) // 这里面会调用Reconcile，里面对controller进行了初始化，并将controller add到了manager，controller.do 存放了Reconcile方法
}

// complete中包含两个主要过程，do controller 和 do watch
// do watch 中实现了事件的监听和处理  predicate用来过滤事件，controller start来开始事件的监听
// 事件的入队列，是用Request对象来存储的，包含name和namespace属性

// 模板解析，解析结果放到rlt
func parseTemp(mode string, app crdv1.App, rlt interface{}) error {
	tempFile := path.Join(tmpDir, mode+tmpSuffix)
	temp, err := template.ParseFiles(tempFile)
	if err != nil {
		logger.Error(err, "模板解析出错")
		return err
	}

	writer := bytes.NewBuffer(make([]byte, 0)) // bytes.NewBuffer 和 bytes.Buffer啥区别？
	writer = new(bytes.Buffer)
	err = temp.Execute(writer, app)
	if err != nil {
		logger.Error(err, "模板替换出错")
		return err
	}
	fmt.Printf("替换完为%s", writer)
	err = yaml.Unmarshal(writer.Bytes(), rlt)
	if err != nil {
		logger.Error(err, "yaml解析错误")
		return err
	}
	return nil
}
