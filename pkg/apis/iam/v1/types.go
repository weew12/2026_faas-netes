package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyList Policy 资源列表
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Policy `json:"items"`
}

// ConditionMap 条件映射，用于定义策略的匹配条件
type ConditionMap map[string]map[string][]string

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Policy 用于为函数定义访问权限策略
// +kubebuilder:printcolumn:name="Statement",type=string,JSONPath=`.spec.statement`
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PolicySpec `json:"spec"`
}

// PolicySpec Policy 资源的规范定义
type PolicySpec struct {
	Statement []PolicyStatement `json:"statement"`
}

// PolicyStatement 策略语句，定义具体的权限规则
type PolicyStatement struct {
	// SID 策略的唯一标识
	SID string `json:"sid"`

	// Effect 策略生效类型，当前仅支持允许(Allow)
	Effect string `json:"effect"`

	// Action 策略适用的操作集合，例如 Function:Read
	Action []string `json:"action"`

	// Resource 策略适用的资源集合，当前仅支持命名空间
	Resource []string `json:"resource"`

	// +optional
	// Condition 策略适用的条件集合（可选）
	Condition *ConditionMap `json:"condition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoleList Role 资源列表
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Role `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Role 用于为函数定义角色，关联身份与策略
// +kubebuilder:printcolumn:name="Principal",type=string,JSONPath=`.spec.principal`
// +kubebuilder:printcolumn:name="Condition",type=string,JSONPath=`.spec.condition`
// +kubebuilder:printcolumn:name="Policy",type=string,JSONPath=`.spec.policy`
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RoleSpec `json:"spec"`
}

// RoleSpec 将 JWT 中的身份/属性映射到一组策略
type RoleSpec struct {
	// +optional
	// Policy 应用于该角色的命名策略列表（可选）
	Policy []string `json:"policy"`

	// +optional
	// Principal 角色适用的身份主体（可选）
	Principal map[string][]string `json:"principal"`

	// +optional
	// Condition 用于匹配 JWT 声明的条件集合（可选）
	Condition *ConditionMap `json:"condition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JwtIssuerList JwtIssuer 资源列表
type JwtIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JwtIssuer `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JwtIssuer 用于定义函数的 JWT 签发者
// +kubebuilder:printcolumn:name="Issuer",type=string,JSONPath=`.spec.iss`
// +kubebuilder:printcolumn:name="Audience",type=string,JSONPath=`.spec.aud`
// +kubebuilder:printcolumn:name="Expiry",type=string,JSONPath=`.spec.tokenExpiry`
type JwtIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JwtIssuerSpec `json:"spec"`
}

// JwtIssuerSpec JwtIssuer 资源的规范定义
type JwtIssuerSpec struct {
	// Issuer JWT 签发者标识
	Issuer string `json:"iss"`

	// +optional
	// IssuerInternal 内部使用的公钥下载地址（可选），适用于系统签发者
	IssuerInternal string `json:"issInternal,omitempty"`

	// Audience JWT 的目标受众，通常为客户端 ID
	Audience []string `json:"aud"`

	// +optional
	// TokenExpiry Token 过期时间（可选）
	TokenExpiry string `json:"tokenExpiry,omitempty"`
}
