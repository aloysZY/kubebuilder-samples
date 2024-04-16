package utils

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	aloystechv1 "aloys.tech/api/v1"
	appv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parseTemplate(templateName string, app *aloystechv1.App) []byte {
	tmpl, err := template.ParseFiles("internal/template/" + templateName + ".yml")
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func NewDeployment(app *aloystechv1.App) *appv1.Deployment {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recover NewDeployment panic:%v\n", r)
		}
	}()
	d := &appv1.Deployment{}
	err := yaml.Unmarshal(parseTemplate("deployment", app), d)
	if err != nil {
		panic(err)
	}
	return d
}

func NewIngress(app *aloystechv1.App) *netv1.Ingress {
	i := &netv1.Ingress{}
	err := yaml.Unmarshal(parseTemplate("ingress", app), i)
	if err != nil {
		panic(err)
	}
	return i
}

func NewService(app *aloystechv1.App) *corev1.Service {
	s := &corev1.Service{}
	switch strings.ToUpper(app.Spec.Service.Type) {
	case "NODEPORT":
		err := yaml.Unmarshal(parseTemplate("service", app), s)
		if err != nil {
			panic(err)
		}
		return s
	default:
		err := yaml.Unmarshal(parseTemplate("service_nodePort", app), s)
		if err != nil {
			panic(err)
		}
		return s
	}
}

func NewHorizontalPodAutoscaler(app *aloystechv1.App) *autoscalingv2.HorizontalPodAutoscaler {
	h := &autoscalingv2.HorizontalPodAutoscaler{}
	err := yaml.Unmarshal(parseTemplate("hpa", app), h)
	if err != nil {
		return nil
	}
	return h
}
