package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// 加载配置文件，生成config对象
	// config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err.Error())
	}
	// 配置 API路径和请求的资源组/资源版本信息 config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 配置数据的编解码器
	config.NegotiatedSerializer = scheme.Codecs
	// 要查询的数据所在的资源组
	config.APIPath = "/api"
	// 实例化 RESTClient 对象
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err.Error())
	}
	// 预设返回值存放对象
	result := &corev1.PodList{}
	// Get 方法设置 HTTP 请求方法 ;Namespace 方法设置操作的命名空间
	// Resource 方法设置操作的资源类型 ;VersionedParams 方法设置请求的查询参数
	// Do 方法发起请求并用 Into 方法将 APIServer 返回的结果写入 Result 变量中
	// Limit: 100 分页
	err = restClient.Get().
		Namespace("default").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 2}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result)
	if err != nil {
		panic(err)
	}

	// 打印 Pod 信息
	for _, d := range result.Items {
		fmt.Printf(
			"NAMESPACE:%v \t NAME: %v \t STATUS: %v\n",
			d.Namespace,
			d.Name,
			d.Status.Phase,
		)
	}
}
