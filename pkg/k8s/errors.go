// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// IsNotFound 判断错误是否为 Kubernetes 资源不存在/已删除错误
func IsNotFound(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}
