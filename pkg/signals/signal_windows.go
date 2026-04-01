// Package signals 提供进程信号处理相关定义
package signals

import (
	"os"
)

// shutdownSignals 定义进程关闭信号
var shutdownSignals = []os.Signal{os.Interrupt}
