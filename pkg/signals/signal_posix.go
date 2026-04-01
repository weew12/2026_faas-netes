//go:build !windows
// +build !windows

// Package signals 定义程序关闭相关的信号常量
package signals

import (
	"os"
	"syscall"
)

// shutdownSignals 系统关闭信号集合，包含中断信号和终止信号
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
