// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

// removeVolume 从 Volume 切片中移除指定名称的卷
// 使用无内存分配过滤算法，高效且无额外开销
func removeVolume(volumeName string, volumes []corev1.Volume) []corev1.Volume {
	if volumes == nil {
		return []corev1.Volume{}
	}

	newVolumes := volumes[:0]
	for _, v := range volumes {
		if v.Name != volumeName {
			newVolumes = append(newVolumes, v)
		}
	}

	return newVolumes
}

// removeVolumeMount 从 VolumeMount 切片中移除指定名称的挂载项
// 使用无内存分配过滤算法，高效且无额外开销
func removeVolumeMount(volumeName string, mounts []corev1.VolumeMount) []corev1.VolumeMount {
	if mounts == nil {
		return []corev1.VolumeMount{}
	}

	newMounts := mounts[:0]
	for _, v := range mounts {
		if v.Name != volumeName {
			newMounts = append(newMounts, v)
		}
	}

	return newMounts
}
