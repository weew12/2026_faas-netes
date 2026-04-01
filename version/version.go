// License: OpenFaaS Community Edition (CE) EULA
// Copyright (c) 2017,2019-2024 OpenFaaS Author(s)

// package version 提供版本管理功能，用于获取组件的版本号、Git提交信息
package version

var (
	// Version 组件正式发布版本号
	Version string

	// GitCommit 最后一次Git提交的SHA哈希值
	GitCommit string

	// DevVersion 开发环境默认版本标识
	DevVersion = "dev"
)

// BuildVersion 构建并返回当前组件版本
// 未设置正式版本时返回开发版本标识
func BuildVersion() string {
	if len(Version) == 0 {
		return DevVersion
	}
	return Version
}

// GetReleaseInfo 获取组件完整发布信息
// 返回：sha=Git提交哈希，release=组件版本
func GetReleaseInfo() (sha, release string) {
	return GitCommit, BuildVersion()
}
