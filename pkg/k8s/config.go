// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Package k8s 提供 Kubernetes 部署相关配置定义
package k8s

// ProbeConfig 存储部署的存活检查与就绪检查配置
type ProbeConfig struct {
	// 容器启动后首次检查等待时间
	InitialDelaySeconds int32
	// 检查超时时间
	TimeoutSeconds int32
	// 检查执行周期
	PeriodSeconds int32
}

// DeploymentConfig 存储全局部署配置选项
type DeploymentConfig struct {
	// 运行时 HTTP 服务端口
	RuntimeHTTPPort int32
	// 是否使用 HTTP 健康检查
	HTTPProbe bool
	// 就绪检查配置
	ReadinessProbe *ProbeConfig
	// 存活检查配置
	LivenessProbe *ProbeConfig
	// SetNonRootUser 是否覆盖函数用户为非 root 身份
	// 启用后所有函数将强制使用 UID 12000
	SetNonRootUser bool
}
