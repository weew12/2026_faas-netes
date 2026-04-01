// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	"context"

	vv1 "github.com/openfaas/faas-netes/pkg/apis/openfaas/v1"
	openfaasv1 "github.com/openfaas/faas-netes/pkg/client/clientset/versioned/typed/openfaas/v1"

	v1 "github.com/openfaas/faas-netes/pkg/client/listers/openfaas/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// FunctionFactory 处理Kubernetes相关操作，将函数转换为Deployment和Service
type FunctionFactory struct {
	Client kubernetes.Interface
	Config DeploymentConfig
}

// NewFunctionFactory 创建FunctionFactory实例
func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig, faasclient openfaasv1.OpenfaasV1Interface) FunctionFactory {
	return FunctionFactory{
		Client: clientset,
		Config: config,
	}
}

// Lister OpenFaaS资源列表器
type Lister struct {
	f openfaasv1.OpenfaasV1Interface
}

// Profiles 获取指定命名空间的Profile列表器
func (l *Lister) Profiles(namespace string) v1.ProfileNamespaceLister {
	return &NamespaceLister{f: l.f, ns: namespace}
}

// NamespaceLister 命名空间级别的Profile资源列表器
type NamespaceLister struct {
	f  openfaasv1.OpenfaasV1Interface
	ns string
}

// Get 根据名称获取指定的Profile资源
func (l *NamespaceLister) Get(name string) (ret *vv1.Profile, err error) {
	value, err := l.f.Profiles(l.ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return value, nil
}

// List 根据标签选择器查询Profile资源列表
func (l *NamespaceLister) List(selector labels.Selector) (ret []*vv1.Profile, err error) {
	list, err := l.f.Profiles(l.ns).List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})

	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		ret = append(ret, &item)
	}

	return ret, nil
}
