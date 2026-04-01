// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	"path/filepath"

	types "github.com/openfaas/faas-provider/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// FunctionProbes 函数健康检查配置，包含存活检查与就绪检查
type FunctionProbes struct {
	Liveness  *corev1.Probe
	Readiness *corev1.Probe
}

// MakeProbes 创建并返回函数的存活检查和就绪检查配置
// 默认健康检查使用 cat /tmp/.lock 命令，每10秒执行一次
func (f *FunctionFactory) MakeProbes(r types.FunctionDeployment) (*FunctionProbes, error) {
	var handler corev1.ProbeHandler

	// 根据配置选择 HTTP 健康检查 或 文件检查
	if f.Config.HTTPProbe {
		handler = corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/_/health",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: int32(f.Config.RuntimeHTTPPort),
				},
			},
		}
	} else {
		path := filepath.Join("/tmp/", ".lock")
		handler = corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"cat", path},
			},
		}
	}

	probes := FunctionProbes{}
	// 配置就绪检查
	probes.Readiness = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: f.Config.ReadinessProbe.InitialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.ReadinessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.ReadinessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	// 配置存活检查
	probes.Liveness = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: f.Config.LivenessProbe.InitialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.LivenessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.LivenessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	return &probes, nil
}
