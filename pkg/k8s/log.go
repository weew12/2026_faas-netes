// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	"context"
	"log"
	"strings"

	"github.com/openfaas/faas-provider/logs"
	"k8s.io/client-go/kubernetes"
)

// LogRequestor Kubernetes 日志获取实现，满足 logs.Requestor 接口
type LogRequestor struct {
	client            kubernetes.Interface
	functionNamespace string
}

// NewLogRequestor 创建 Kubernetes 日志请求器实例
func NewLogRequestor(client kubernetes.Interface, functionNamespace string) *LogRequestor {
	return &LogRequestor{
		client:            client,
		functionNamespace: functionNamespace,
	}
}

// Query 实现日志查询接口，获取函数 Pod 日志
// 忽略 Limit 参数，由上层 OpenFaaS Provider 控制行数限制
func (l LogRequestor) Query(ctx context.Context, r logs.Request) (<-chan logs.Message, error) {
	ns := l.functionNamespace

	if len(r.Namespace) > 0 && strings.ToLower(r.Namespace) != "kube-system" {
		ns = r.Namespace
	}

	logStream, err := GetLogs(ctx, l.client, r.Name, ns, int64(r.Tail), r.Since, r.Follow)
	if err != nil {
		log.Printf("LogRequestor: get logs failed: %s\n", err)
		return nil, err
	}

	msgStream := make(chan logs.Message, LogBufferSize)
	go func() {
		defer close(msgStream)
		// 上下文取消时 logStream 会关闭，确保协程退出
		for msg := range logStream {
			msgStream <- logs.Message{
				Timestamp: msg.Timestamp,
				Text:      msg.Text,
				Name:      msg.FunctionName,
				Instance:  msg.PodName,
				Namespace: msg.Namespace,
			}
		}
	}()

	return msgStream, nil
}
