package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	controller "github.com/openfaas/faas-netes/pkg/apis/openfaas"
)

// SchemeGroupVersion API组与版本，用于注册OpenFaaS资源对象
var SchemeGroupVersion = schema.GroupVersion{Group: controller.GroupName, Version: "v1"}

// Resource 将资源名称转换为带API组限定的GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder 资源类型注册构建器
	// localSchemeBuilder 本地Scheme构建器实例
	// AddToScheme 将资源类型添加到Scheme
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// 仅在此注册手动编写的函数，自动生成代码的注册逻辑放在生成文件中
	// 分离设计可避免缺失生成文件时导致编译失败
	localSchemeBuilder.Register(addKnownTypes)
}

// addKnownTypes 将OpenFaaS已知资源类型注册到Scheme
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Function{},
		&FunctionList{},
		&Profile{},
		&ProfileList{},
	)
	// 将当前API组版本添加到Scheme
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
