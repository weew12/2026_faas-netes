package k8s

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	v1 "k8s.io/client-go/listers/apps/v1"
)

// FunctionList 用于查询和统计 OpenFaaS 函数部署的工具
type FunctionList struct {
	deployLister      v1.DeploymentLister
	namespace         string
	functionsSelector labels.Selector
}

// NewFunctionList 创建函数列表实例，自动添加函数标签筛选器
func NewFunctionList(namespace string, deployLister v1.DeploymentLister) *FunctionList {

	sel := labels.NewSelector()
	req, _ := labels.NewRequirement("faas_function", selection.Exists, []string{})
	onlyFunctions := sel.Add(*req)

	return &FunctionList{
		deployLister:      deployLister,
		namespace:         namespace,
		functionsSelector: onlyFunctions,
	}
}

// Count 统计当前命名空间中函数部署的数量
func (f *FunctionList) Count() (int, error) {
	list, err := f.deployLister.Deployments(f.namespace).List(f.functionsSelector)
	if err != nil {
		return 0, err
	}

	return len(list), nil
}
