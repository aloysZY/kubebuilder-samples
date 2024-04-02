package main

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 加载kubeconfig文件，生成config对象
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	// dynamic.NewForConfig 函数通过 config 实例化 dynamicClient 对象
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	// 通过 schema.GroupVersionResource 设置要请求对象的资源组、资源版本和资源
	// 设置命名空间和请求参数,得到 unstructured.UnstructuredList 指针类型的 PodList
	gvr := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	unstructObj, err := dynamicClient.Resource(gvr).Namespace("kube-system").List(context.TODO(), metav1.ListOptions{Limit: 10})
	if err != nil {
		panic(err)
	}
	// 通过 runtime.DefaultUnstructuredConverter 函数将 unstructured.UnstructuredList 转为 DeploymentList 类型
	deploymentList := &appsv1.DeploymentList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		unstructObj.UnstructuredContent(),
		deploymentList,
	)
	if err != nil {
		panic(err)
	}
	for _, v := range deploymentList.Items {
		fmt.Printf(
			"KIND: %v \t NAMESPACE: %v \t NAME:%v \n",
			v.Kind,
			v.Namespace,
			v.Name,
		)
	}
}
