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

package main

import (
	"crypto/tls"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	aloystechv1 "aloys.tech/api/v1"
	"aloys.tech/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	// scheme 它提供了 Kinds 与对应的 Go Type 的映射，即给定了 Go Type，就能够知道它的 GKV(Group Kind Verision)，这也是 Kubernetes 所有资源的注册模式
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// 每个版本的资源生成的过程中， 都会包含 groupversion_info.go、zz_generated.deepcopy.go 文件，它们的作用是什么呢? 这与 Scheme 模块的原理有关，即 Scheme 通过这 2 个文件实现了 CRD 的注册及资源的拷贝
func init() {
	// Scheme 绑定内置资源
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Scheme 绑定自建 CRD
	utilruntime.Must(aloystechv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancelation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})
	// 在 NewManager 的方法中，实际是根据传入的参数进行 Manager 对象的 Scheme、Cache、Client 等模块的初始化构建
	// Client 实现对 CRD 的“增、删、改、查”, 其中查询的逻辑是通过本地的 Cache 模块实现的
	// 初始化 Manager 对象构建出来后，通过 Manager 的 Cache 监听 CRD，一旦 CRD 在集群中创建了，Cache 监听 到发生了变化，就会触发 Controller 的协调程序 Reconcile 工作
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		// utilruntime.Must(aloystechv1.AddToScheme(scheme)) 这里的 Scheme 已经绑定了自建 CRD
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "3190ef45.aloys.tech",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	// 初始化失败，退出主程序
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// 初始化controller.AppReconciler
	if err = (&controller.AppReconciler{
		// 将 Manager 的 Client 传给 AppReconciler， (r *AppReconciler) Reconciler方法就可以使用client
		Client: mgr.GetClient(),
		// 将 Manager 的 Scheme 传给 AppReconciler， get/list获取集群信息默认是先查询Scheme
		Scheme: mgr.GetScheme(),
		// 初始化事件方法
		Eventer: mgr.GetEventRecorderFor("app-controller"),
		// 并且调用 SetupWithManager 方法传入 Manager 进行 Controller 的初始化
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	// MGR 的类型是一个 Interface，底层实际上调用的是 controllerManager 的 Start 方法。 Start 方法的主要逻辑就是启动 Cache、Controller，将整个事件流运转起来,
	// 先初始化 Cache(Cluster类型)，再启动 Controller
	// Cache 的核心逻辑是初始化内部所有的 Informer，初始化 Informer 后就创建了 Reflector 和内部 Controller，Reflector 和 Controller 两个组件是一个“生产者—消费者” 模型，Reflector 负责监听 APIServer 上指定的 GVK 资源的变化，然后将变更写入 delta 队列中，Controller 负责消费这些变更的事件，然后更新本地 Indexer，最后计算出是创建、 更新，还是删除事件，推给我们之前注册的 Watch Handler
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
