// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	types "github.com/openfaas/faas-provider/types"
	appsv1 "k8s.io/api/apps/v1"
)

// EnvProcessName 函数进程名称对应的环境变量名
const EnvProcessName = "fprocess"

// AsFunctionStatus 将Kubernetes Deployment对象转换为OpenFaaS FunctionStatus
// 解析Deployment和容器规格，生成简化的函数状态摘要
func AsFunctionStatus(item appsv1.Deployment) *types.FunctionStatus {
	var replicas uint64
	if item.Spec.Replicas != nil {
		replicas = uint64(*item.Spec.Replicas)
	}

	// 获取函数主容器
	functionContainer := item.Spec.Template.Spec.Containers[0]

	labels := item.Spec.Template.Labels
	// 构建函数状态对象
	function := types.FunctionStatus{
		Name:              item.Name,
		Replicas:          replicas,
		Image:             functionContainer.Image,
		AvailableReplicas: uint64(item.Status.AvailableReplicas),
		InvocationCount:   0,
		Labels:            &labels,
		Annotations:       &item.Spec.Template.Annotations,
		Namespace:         item.Namespace,
		Secrets:           ReadFunctionSecretsSpec(item),
		CreatedAt:         item.CreationTimestamp.Time,
	}

	// 解析资源请求与限制
	req := &types.FunctionResources{Memory: functionContainer.Resources.Requests.Memory().String(), CPU: functionContainer.Resources.Requests.Cpu().String()}
	lim := &types.FunctionResources{Memory: functionContainer.Resources.Limits.Memory().String(), CPU: functionContainer.Resources.Limits.Cpu().String()}

	// 赋值非零的资源配置
	if req.CPU != "0" || req.Memory != "0" {
		function.Requests = req
	}
	if lim.CPU != "0" || lim.Memory != "0" {
		function.Limits = lim
	}

	// 提取函数进程环境变量
	for _, v := range functionContainer.Env {
		if EnvProcessName == v.Name {
			function.EnvProcess = v.Value
		}
	}

	return &function
}
