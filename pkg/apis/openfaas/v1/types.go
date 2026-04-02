package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// +genclient
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type == "Ready")].status`,priority=1,description="The function's desired state has been applied by the controller"
// +kubebuilder:printcolumn:name="Healthy",type=string,JSONPath=`.status.conditions[?(@.type == "Healthy")].status`,description="All replicas of the function's desired state are available to serve traffic"
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`,description="The desired number of replicas"
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Unavailable",type=integer,JSONPath=`.status.unavailableReplicas`,priority=1

// Function OpenFaaS 函数资源定义
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FunctionSpec `json:"spec"`

	// +optional
	Status FunctionStatus `json:"status,omitempty"`
}

// FunctionSpec 函数资源规格定义
type FunctionSpec struct {
	// 函数名称
	Name string `json:"name"`

	// 函数容器镜像
	Image string `json:"image"`
	// +optional
	// 函数处理入口（可选）
	Handler string `json:"handler,omitempty"`
	// +optional
	// 注解（可选）
	Annotations *map[string]string `json:"annotations,omitempty"`
	// +optional
	// 标签（可选）
	Labels *map[string]string `json:"labels,omitempty"`
	// +optional
	// 环境变量（可选）
	Environment *map[string]string `json:"environment,omitempty"`
	// +optional
	// 部署约束（可选）
	Constraints []string `json:"constraints,omitempty"`
	// +optional
	// 密钥引用（可选）
	Secrets []string `json:"secrets,omitempty"`
	// +optional
	// 资源限制（可选）
	Limits *FunctionResources `json:"limits,omitempty"`
	// +optional
	// 资源请求（可选）
	Requests *FunctionResources `json:"requests,omitempty"`
	// +optional
	// 根文件系统只读
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`
}

// FunctionResources 函数 CPU / 内存资源配置
type FunctionResources struct {
	// 内存配置
	Memory string `json:"memory,omitempty"`
	// CPU 配置
	CPU string `json:"cpu,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FunctionList 函数资源列表
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Function `json:"items"`
}

// FunctionStatus 函数运行状态
type FunctionStatus struct {
	// Conditions 资源状态观测条件
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +optional
	// 期望副本数
	Replicas int32 `json:"replicas,omitempty"`

	// +optional
	// 可用副本数
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// +optional
	// 不可用副本数
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// 控制器观测到的资源版本
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// 应用到当前函数的 OpenFaaS 配置文件
	// +optional
	Profiles []AppliedProfile `json:"profiles,omitempty"`
}

// AppliedProfile 已应用到函数的配置文件描述
type AppliedProfile struct {
	// 关联的 Profile 对象引用
	ProfileRef ResourceRef `json:"profileRef"`

	// 应用到函数的 Profile 资源版本
	ObservedGeneration int64 `json:"observedGeneration"`
}

// ResourceRef 跨命名空间资源引用
type ResourceRef struct {
	// 资源名称
	Name string `json:"name,omitempty"`
	// 资源命名空间
	Namespace string `json:"namespace,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Profile 用于自定义函数 Pod 模板的配置文件
type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProfileSpec `json:"spec"`
}

// ProfileSpec OpenFaaS 扩展配置，可预定义并通过注解应用到函数
type ProfileSpec struct {
	// Pod 容忍配置
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// 运行时类名
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`

	// Pod 安全上下文
	// +optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// Pod 调度亲和性
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// 拓扑分布约束
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// DNS 策略
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// 自定义 DNS 配置
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty"`

	// 容器资源请求与限制
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// 优先级类名
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// 部署策略
	// +optional
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProfileList Profile 资源列表
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Profile `json:"items"`
}
