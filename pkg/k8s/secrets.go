// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

package k8s

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	types "github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// secretsMountPath 密钥挂载路径
	secretsMountPath = "/var/openfaas/secrets"
	// secretLabel 用于标识 OpenFaaS 托管密钥的标签
	secretLabel = "app.kubernetes.io/managed-by"
	// secretLabelValue 标签值
	secretLabelValue = "openfaas"
	// secretsProjectVolumeNameTmpl 投影卷名称模板
	secretsProjectVolumeNameTmpl = "%s-projected-secrets"
)

// SecretsClient 定义 Kubernetes 密钥的标准化 CRUD 接口
// 确保密钥格式和标签符合 OpenFaaS 使用规范
type SecretsClient interface {
	// List 返回可用的函数密钥名称列表（不返回敏感值）
	List(namespace string) (names []string, err error)
	// Create 创建新的函数密钥
	Create(secret types.Secret) error
	// Replace 更新函数密钥的值
	Replace(secret types.Secret) error
	// Delete 删除函数密钥
	Delete(name string, namespace string) error
	// GetSecrets 根据名称批量查询密钥详情
	GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error)
}

// SecretInterfacer 暴露获取 SecretInterface 的方法
// 用于在指定命名空间下操作密钥
type SecretInterfacer interface {
	Secrets(namespace string) typedV1.SecretInterface
}

// secretClient SecretsClient 的 Kubernetes 实现
type secretClient struct {
	kube SecretInterfacer
}

// NewSecretsClient 创建 SecretsClient 实例
func NewSecretsClient(kube kubernetes.Interface) SecretsClient {
	return &secretClient{
		kube: kube.CoreV1(),
	}
}

// List 列出指定命名空间下由 OpenFaaS 管理的所有密钥名称
func (c secretClient) List(namespace string) (names []string, err error) {
	res, err := c.kube.Secrets(namespace).List(context.TODO(), c.selector())
	if err != nil {
		log.Printf("failed to list secrets in %s: %v\n", namespace, err)
		return nil, err
	}

	names = make([]string, len(res.Items))
	for idx, item := range res.Items {
		names[idx] = item.Name
	}
	return names, nil
}

// Create 创建带正确标签和结构的函数密钥
func (c secretClient) Create(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	req := &apiv1.Secret{
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Labels: map[string]string{
				secretLabel: secretLabelValue,
			},
		},
	}

	req.Data = c.getValidSecretData(secret)

	_, err = c.kube.Secrets(secret.Namespace).Create(context.TODO(), req, metav1.CreateOptions{})
	if err != nil {
		log.Printf("failed to create secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	log.Printf("created secret %s.%s\n", secret.Name, secret.Namespace)

	return nil
}

// Replace 覆盖更新已存在的函数密钥
func (c secretClient) Replace(secret types.Secret) error {
	err := c.validateSecret(secret)
	if err != nil {
		return err
	}

	kube := c.kube.Secrets(secret.Namespace)
	found, err := kube.Get(context.TODO(), secret.Name, metav1.GetOptions{})
	if err != nil {
		log.Printf("can not retrieve secret for update %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	found.Data = c.getValidSecretData(secret)

	_, err = kube.Update(context.TODO(), found, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("can not update secret %s.%s: %v\n", secret.Name, secret.Namespace, err)
		return err
	}

	return nil
}

// Delete 删除指定的函数密钥
func (c secretClient) Delete(namespace string, name string) error {
	err := c.kube.Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("can not delete %s.%s: %v\n", name, namespace, err)
	}
	return err
}

// GetSecrets 批量获取密钥对象
func (c secretClient) GetSecrets(namespace string, secretNames []string) (map[string]*apiv1.Secret, error) {
	kube := c.kube.Secrets(namespace)
	opts := metav1.GetOptions{}

	secrets := map[string]*apiv1.Secret{}
	for _, secretName := range secretNames {
		secret, err := kube.Get(context.TODO(), secretName, opts)
		if err != nil {
			return nil, err
		}
		secrets[secretName] = secret
	}

	return secrets, nil
}

// selector 生成标签选择器，只筛选 OpenFaaS 管理的密钥
func (c secretClient) selector() metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", secretLabel, secretLabelValue),
	}
}

// validateSecret 校验密钥的名称和命名空间非空
func (c secretClient) validateSecret(secret types.Secret) error {
	if strings.TrimSpace(secret.Namespace) == "" {
		return errors.New("namespace may not be empty")
	}

	if strings.TrimSpace(secret.Name) == "" {
		return errors.New("name may not be empty")
	}

	return nil
}

// getValidSecretData 统一处理密钥值，优先使用 RawValue
func (c secretClient) getValidSecretData(secret types.Secret) map[string][]byte {

	if len(secret.RawValue) > 0 {
		return map[string][]byte{
			secret.Name: secret.RawValue,
		}
	}

	return map[string][]byte{
		secret.Name: []byte(secret.Value),
	}

}

// ConfigureSecrets 为 Deployment 配置密钥挂载和镜像拉取密钥
// 区分 Docker 密钥和普通文件密钥
func (f *FunctionFactory) ConfigureSecrets(request types.FunctionDeployment, deployment *appsv1.Deployment, existingSecrets map[string]*apiv1.Secret) error {
	secretVolumeProjections := []apiv1.VolumeProjection{}

	for _, secretName := range request.Secrets {
		deployedSecret, ok := existingSecrets[secretName]
		if !ok {
			return fmt.Errorf("Required secret '%s' was not found in the cluster", secretName)
		}

		switch deployedSecret.Type {

		case apiv1.SecretTypeDockercfg,
			apiv1.SecretTypeDockerConfigJson:

			deployment.Spec.Template.Spec.ImagePullSecrets = append(
				deployment.Spec.Template.Spec.ImagePullSecrets,
				apiv1.LocalObjectReference{
					Name: secretName,
				},
			)
		default:

			projectedPaths := []apiv1.KeyToPath{}
			for secretKey := range deployedSecret.Data {
				projectedPaths = append(projectedPaths, apiv1.KeyToPath{Key: secretKey, Path: secretKey})
			}

			projection := &apiv1.SecretProjection{Items: projectedPaths}
			projection.Name = secretName
			secretProjection := apiv1.VolumeProjection{
				Secret: projection,
			}
			secretVolumeProjections = append(secretVolumeProjections, secretProjection)
		}
	}

	volumeName := fmt.Sprintf(secretsProjectVolumeNameTmpl, request.Service)
	projectedSecrets := apiv1.Volume{
		Name: volumeName,
		VolumeSource: apiv1.VolumeSource{
			Projected: &apiv1.ProjectedVolumeSource{
				Sources: secretVolumeProjections,
			},
		},
	}

	existingVolumes := removeVolume(volumeName, deployment.Spec.Template.Spec.Volumes)
	deployment.Spec.Template.Spec.Volumes = existingVolumes
	if len(secretVolumeProjections) > 0 {
		deployment.Spec.Template.Spec.Volumes = append(existingVolumes, projectedSecrets)
	}

	updatedContainers := []apiv1.Container{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		mount := apiv1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: secretsMountPath,
		}

		container.VolumeMounts = removeVolumeMount(volumeName, container.VolumeMounts)
		if len(secretVolumeProjections) > 0 {
			container.VolumeMounts = append(container.VolumeMounts, mount)
		}

		updatedContainers = append(updatedContainers, container)
	}

	deployment.Spec.Template.Spec.Containers = updatedContainers

	return nil
}

// ReadFunctionSecretsSpec 从 Deployment 中解析出函数使用的密钥名称
// 是 ConfigureSecrets 的逆操作
func ReadFunctionSecretsSpec(item appsv1.Deployment) []string {
	secrets := []string{}

	for _, s := range item.Spec.Template.Spec.ImagePullSecrets {
		secrets = append(secrets, s.Name)
	}

	volumeName := fmt.Sprintf(secretsProjectVolumeNameTmpl, item.Name)
	var sourceSecrets []apiv1.VolumeProjection
	for _, v := range item.Spec.Template.Spec.Volumes {
		if v.Name == volumeName {
			sourceSecrets = v.Projected.Sources
			break
		}
	}

	for _, s := range sourceSecrets {
		if s.Secret == nil {
			continue
		}
		secrets = append(secrets, s.Secret.Name)
	}

	sort.Strings(secrets)
	return secrets
}
