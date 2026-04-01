package signals

import (
	"os"
	"os/signal"
)

// onlyOneSignalHandler 确保信号处理器仅被初始化一次
var onlyOneSignalHandler = make(chan struct{})

// SetupSignalHandler 注册监听系统退出信号（SIGTERM/SIGINT）
// 返回一个停止通道，收到信号时通道会被关闭
// 若捕获到第二次退出信号，程序将直接以退出码 1 终止运行
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // 调用两次会触发 panic，强制单例

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // 第二次收到信号，直接退出
	}()

	return stop
}
