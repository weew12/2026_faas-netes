// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	types "github.com/openfaas/faas-provider/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// SecurityContextUserID 启用非 root 用户时使用的 UID
// 取值大于 10000，遵循 kubesec.io 安全建议
const SecurityContextUserID = int64(12000)

// ConfigureContainerUserID 为函数容器设置运行用户 UID
// SetNonRootUser 为 true 时使用 UID 12000，否则使用镜像默认用户（可能为 root）
func (f *FunctionFactory) ConfigureContainerUserID(deployment *appsv1.Deployment) {
	userID := SecurityContextUserID
	var functionUser *int64

	if f.Config.SetNonRootUser {
		functionUser = &userID
	}

	if deployment.Spec.Template.Spec.Containers[0].SecurityContext == nil {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
	}

	deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = functionUser
}

// ConfigureReadOnlyRootFilesystem 配置容器根文件系统只读模式
// 1. 启用时：设置 ReadOnlyRootFilesystem=true，并挂载临时目录 /tmp
// 2. 禁用时：关闭只读模式，并移除 /tmp 挂载
// 适用于创建与更新操作
func (f *FunctionFactory) ConfigureReadOnlyRootFilesystem(request types.FunctionDeployment, deployment *appsv1.Deployment) {
	if deployment.Spec.Template.Spec.Containers[0].SecurityContext != nil {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = &request.ReadOnlyRootFilesystem
	} else {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			ReadOnlyRootFilesystem: &request.ReadOnlyRootFilesystem,
		}
	}

	existingVolumes := removeVolume("temp", deployment.Spec.Template.Spec.Volumes)
	deployment.Spec.Template.Spec.Volumes = existingVolumes

	existingMounts := removeVolumeMount("temp", deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = existingMounts

	if request.ReadOnlyRootFilesystem {
		deployment.Spec.Template.Spec.Volumes = append(
			existingVolumes,
			corev1.Volume{
				Name: "temp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		)

		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
			existingMounts,
			corev1.VolumeMount{
				Name:      "temp",
				MountPath: "/tmp",
				ReadOnly:  false},
		)
	}
}
