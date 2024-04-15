/*
Copyright 2024.

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

	aloystechv1 "aloys.tech/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aloys.tech.aloys.tech,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aloys.tech.aloys.tech,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aloys.tech.aloys.tech,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
// Reconcile 的 含义，用户自定义了 CRD 结构，而在 Kubernetes 集群中，想要实现这样的 CRD 结构定义， Reconcile 需要协调逻辑
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	// Requeue 告诉 Controller 是否需要重新将对象加入队列（从新调用，不是直接重试），默认为 False
	// RequeueAfter 大于 0 表示 Controller 需要在设置的时间间隔后，将对象重新加入队列 注意，当设置了RequeueAfter，就表示Requeue为True，即无须RequeueAfter与 Requeue=True 被同时设置
	// ctrl.Result{Requeue: true, RequeueAfter: 1}
	// TODO(user): your logic here

	// logger := log.FromContext(ctx)
	// app := &aloystechv1.App{}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// CRD 的 Controller 初始化的核心代码是 SetupWithManager 方法，借助这个方法，就可以完成 CRD 在 Manager 对象中的安装，最后通过 Manager 对象的 start 方法来完成 CRD Controller 的运行
// 在 Controller 初始化的过程中，借助了 Options 参数对象中设计的 Reconciler 对象，并将 其传递给了 Controller 对象的 do 字段。所以当我们调用 SetupWithManager 方法的时候， 不仅完成了 Controller 的初始化，还完成了 Controller 监听资源的注册与发现过程，同时 将 CRD 的必要实现方法(Reconcile 方法)进行了再现
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// NewControllerManagedBy 初始化 Builder 对象 mgr 字段。
	return ctrl.NewControllerManagedBy(mgr).
		// Builder 关联 CRD API 定义的 Scheme 信息，从而得知 CRD 的 Controller 需要监听的 CRD 类型、版本等信息
		// Controller需要监听资源在这里配置 Owns().
		For(&aloystechv1.App{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		// WithOptions(controller.Options{ 可以传入Controller初始化参数
		// 	MaxConcurrentReconciles: 0, // Reconciles 最大并发数
		// 	CacheSyncTimeout:        0, // 是指设置等待同步缓存的时间限制。默认2分钟
		// 	RecoverPanic:            nil, reconcile异常时是否自动恢复
		// 	NeedLeaderElection:      nil, // 控制器是否需要使用leader选举。默认为true，
		// 	Reconciler:              nil,  //定义了 Reconcile(
		// 	RateLimiter:             nil, // 用于限制请求排队的频率。默认为MaxOfRateLimiter，它具有整体和每个项目的速率限制。整体是一个令牌桶，每项是指数级的。
		// 	LogConstructor:          nil, //用于记录日志的日志对象。
		// }).
		// Builder 初始化最重要的两个步骤是doController 和 doWatch
		// doController 是完成 Controller 对象的构建，从而实现基于 Scheme 和 Controller 对象的 CRD 的监听流程
		// predicate.Predicate 是 Controller.Watch 的参数，是用于过滤事件的 过滤器，过滤器可以复用或者组合
		// Owns监听Object，并将Object对应的Owner加入队列。例如，例子中监听Pod对象，根据 Pod 的 Owner 将 Pod 所属的 ReplicaSet 资源加入队列
		// Owns(&corev1.Pod{}).
		// Watches监听指定资源，使用指定方法对事件进行处理。建议使用For()和Owns()，而不是直接使用 Watches() 方法
		// Watches().
		// 设置事件的过滤器，选择部分create/update/delete/generic事件触发同步,只监听实现的方法
		// WithEventFilter(predicate.Predicate()).
		// Named设置Controller的名称，Controller的名称会出现在监控、日志等信息中。在默认情况下，Controller 使用小写字母命名。
		// WithEventFilter(predicate.Funcs{
		//         CreateFunc: func(_ event.CreateEvent) bool {
		//            return false
		//         },
		//      }).
		// Watches(source.Source, handler.EventHandler, ...WatchesOption)
		// For(client.Object,...ForOption)
		// Owns(client.Object,...OwnsOption)
		// 其中For和Owns是等同与Watches。For的第二个参数默认为EnqueueRequestForObject。Owns的第二个参数默认为EnqueueRequestForOwner
		//
		// 方法参数说明
		//
		// Source：第一个参数，kubernetes对象类型
		//
		// EventHandler：第二个参数，从DeltaFIFO中取出来的数据，在进入工作队列前进行的操作。EnqueueRequestForObject表示不做任何处理，直接进入工作队列。EnqueueRequestForOwner需要和For方法配合使用，Owns中的对象中ownerReference引用的对象类型需要和For中定义的对象类型相同，且ownerReference中的controller为true。
		//
		// Predicate：第三个参数，从工作队列取出来的数据，在进行reconcile处理前进行的操作。通过builder的WithEventFilter可以给所有的对象添加Predicate。
		//
		// EventHandler和Predicate方法说明
		//
		// Create：kubernetes对象新增时调用
		//
		// Update：kubernetes对象更新时调用
		//
		// Delete：kubernetes对象删除时调用
		//
		// Generic：未知的操作。非kubernetes集群的变更事件。在operator中自行使用
		Complete(r)
}

// Complete--blder.Build(r)--blder.doController(r)-- blder.ctrl, err = newController(controllerName, blder.mgr, ctrlOptions)这个的返回值 复制给了blder.ctr
//  newController(controllerName--newController = controller.New-- New(name string, mgr manager.Manager--NewUnmanaged(name, mgr, options) 初始化Controller --&controller.Controller{  Do: options.Reconciler, 这个do字段就像（Reconciler reconcile.Reconciler） 这是一个接口类型 type Reconciler interface
// (r *AppReconciler) Reconcile 我们自己的crd实现了这个接口
