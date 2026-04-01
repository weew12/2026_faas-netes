package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	controller "github.com/openfaas/faas-netes/pkg/apis/iam"
)

// SchemeGroupVersion 用于注册API对象的组与版本标识
var SchemeGroupVersion = schema.GroupVersion{Group: controller.GroupName, Version: "v1"}

// Resource 将资源名称转换为带组限定的GroupResource对象
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// localSchemeBuilder 与 AddToScheme 用于K8s资源类型注册
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// 仅在此注册手动编写的函数，生成代码的注册逻辑在自动生成文件中
	// 分离设计可避免因缺少生成文件导致编译失败
	localSchemeBuilder.Register(addKnownTypes)
}

// addKnownTypes 将IAM API的已知资源类型注册到Scheme中
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Policy{},
		&PolicyList{},
		&Role{},
		&RoleList{},
		&JwtIssuer{},
		&JwtIssuerList{},
	)
	// 向Scheme中添加当前API组与版本
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
