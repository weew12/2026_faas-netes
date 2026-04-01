// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2020. All rights reserved.

// Package config 提供服务启动配置读取与打印功能
package config

import (
	"log"

	ftypes "github.com/openfaas/faas-provider/types"
)

// ReadConfig 从环境变量读取配置的实现结构体
type ReadConfig struct {
}

// Read 从环境变量中获取并解析配置
func (ReadConfig) Read(hasEnv ftypes.HasEnv) (BootstrapConfig, error) {
	cfg := BootstrapConfig{}

	faasConfig, err := ftypes.ReadConfig{}.Read(hasEnv)
	if err != nil {
		return cfg, err
	}

	cfg.FaaSConfig = *faasConfig

	// 解析环境变量配置
	httpProbe := ftypes.ParseBoolValue(hasEnv.Getenv("http_probe"), false)
	setNonRootUser := ftypes.ParseBoolValue(hasEnv.Getenv("set_nonroot_user"), false)

	// 获取函数默认命名空间
	cfg.DefaultFunctionNamespace = ftypes.ParseString(hasEnv.Getenv("function_namespace"), "openfaas-fn")

	cfg.HTTPProbe = httpProbe
	cfg.SetNonRootUser = setNonRootUser

	return cfg, nil
}

// BootstrapConfig 服务启动配置，包含服务端配置与函数默认参数
type BootstrapConfig struct {
	// HTTPProbe 启用后通过 HTTP /_/health 检查健康，而非使用 /tmp/.lock 文件
	HTTPProbe bool

	// SetNonRootUser 是否使用非 root 用户（UID 12000）部署函数
	SetNonRootUser bool

	// DefaultFunctionNamespace 函数默认部署命名空间，由 environment: function_namespace 配置
	DefaultFunctionNamespace string

	// FaaSConfig OpenFaaS 核心配置
	FaaSConfig ftypes.FaaSConfig
}

// Fprint 使用日志格式化打印配置信息
// verbose 为 true 时打印完整配置，false 时打印精简配置
func (c BootstrapConfig) Fprint(verbose bool) {
	log.Printf("HTTP Read Timeout: %s\n", c.FaaSConfig.GetReadTimeout())
	log.Printf("HTTP Write Timeout: %s\n", c.FaaSConfig.WriteTimeout)

	log.Printf("ImagePullPolicy: %s\n", "Always")
	log.Printf("DefaultFunctionNamespace: %s\n", c.DefaultFunctionNamespace)

	if verbose {
		log.Printf("MaxIdleConns: %d\n", c.FaaSConfig.MaxIdleConns)
		log.Printf("MaxIdleConnsPerHost: %d\n", c.FaaSConfig.MaxIdleConnsPerHost)
		log.Printf("HTTPProbe: %v\n", c.HTTPProbe)
		log.Printf("SetNonRootUser: %v\n", c.SetNonRootUser)
	}
}
