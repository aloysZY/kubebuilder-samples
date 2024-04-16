package controller

import (
	"k8s.io/apimachinery/pkg/types"
)

func GetNamespacedName(name, suffix, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name + suffix,
		Namespace: namespace,
	}
}
