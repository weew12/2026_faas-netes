package k8s

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/informers/internalinterfaces"

	"k8s.io/client-go/tools/cache"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// podInformerResync Pod 监听器缓存同步周期
	podInformerResync = 5 * time.Second

	// defaultLogSince 默认日志回溯时间
	defaultLogSince = 5 * time.Minute

	// LogBufferSize 日志缓冲区大小
	LogBufferSize = 500 * 2
)

// Log 日志结构体，用于输出格式化日志
type Log struct {
	// Text 日志内容
	Text string `json:"text"`

	// Namespace 命名空间
	Namespace string `json:"namespace"`

	// PodName Pod 实例名称
	PodName string `json:"podName"`

	// FunctionName 函数名称
	FunctionName string `json:"FunctionName"`

	// Timestamp 日志时间戳
	Timestamp time.Time `json:"timestamp"`
}

// GetLogs 获取指定函数的日志通道
func GetLogs(ctx context.Context, client kubernetes.Interface, functionName, namespace string, tail int64, since *time.Time, follow bool) (<-chan Log, error) {
	added, err := startFunctionPodInformer(ctx, client, functionName, namespace)
	if err != nil {
		return nil, err
	}

	logs := make(chan Log, LogBufferSize)

	go func() {
		var watching uint
		defer close(logs)

		finished := make(chan error)

		for {
			select {
			case <-ctx.Done():
				return
			case <-finished:
				watching--
				if watching == 0 && !follow {
					return
				}
			case p := <-added:
				watching++
				go func() {
					finished <- podLogs(ctx, client.CoreV1().Pods(namespace), p, functionName, namespace, tail, since, follow, logs)
				}()
			}
		}
	}()

	return logs, nil
}

// podLogs 从指定 Pod 中流式读取日志
func podLogs(ctx context.Context, i v1.PodInterface, pod, container, namespace string, tail int64, since *time.Time, follow bool, dst chan<- Log) error {
	log.Printf("Logger: starting log stream for %s\n", pod)
	defer log.Printf("Logger: stopping log stream for %s\n", pod)

	opts := &corev1.PodLogOptions{
		Follow:     follow,
		Timestamps: true,
		Container:  container,
	}

	if tail > 0 {
		opts.TailLines = &tail
	}

	if opts.TailLines == nil || since != nil {
		opts.SinceSeconds = parseSince(since)
	}

	stream, err := i.GetLogs(pod, opts).Stream(context.TODO())
	if err != nil {
		return err
	}
	defer stream.Close()

	done := make(chan error)
	go func() {
		reader := bufio.NewReader(stream)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				done <- err
				return
			}
			msg, ts := extractTimestampAndMsg(string(bytes.Trim(line, "\x00")))
			dst <- Log{Timestamp: ts, Text: msg, PodName: pod, FunctionName: container}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != io.EOF {
			return err
		}
		return nil
	}
}

// extractTimestampAndMsg 从日志行中解析时间戳和内容
func extractTimestampAndMsg(logText string) (string, time.Time) {
	parts := strings.SplitN(logText, " ", 2)
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		log.Printf("error: invalid timestamp '%s'\n", parts[0])
		return "", time.Time{}
	}

	if len(parts) == 2 {
		return parts[1], ts
	}

	return "", ts
}

// parseSince 解析日志回溯时间，默认 5 分钟
func parseSince(r *time.Time) *int64 {
	var since int64
	if r == nil || r.IsZero() {
		since = int64(defaultLogSince.Seconds())
		return &since
	}
	since = int64(time.Since(*r).Seconds())
	return &since
}

// startFunctionPodInformer 启动函数 Pod 监听器，监听新增/删除事件
func startFunctionPodInformer(ctx context.Context, client kubernetes.Interface, functionName, namespace string) (<-chan string, error) {
	functionSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{"faas_function": functionName},
	}
	selector, err := metav1.LabelSelectorAsSelector(functionSelector)
	if err != nil {
		err = errors.Wrap(err, "unable to build function selector")
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	log.Printf("PodInformer: starting informer for %s in: %s\n", selector.String(), namespace)
	factory := informers.NewFilteredSharedInformerFactory(
		client,
		podInformerResync,
		namespace,
		withLabels(selector.String()),
	)

	podInformer := factory.Core().V1().Pods()
	podsResp, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	pods := podsResp.Items
	if len(pods) == 0 {
		err = errors.New("no matching instances found")
		log.Printf("PodInformer: %s", err)
		return nil, err
	}

	added := make(chan string, len(pods))
	podInformer.Informer().AddEventHandler(&podLoggerEventHandler{
		added: added,
	})

	go podInformer.Informer().Run(ctx.Done())
	go func() {
		<-ctx.Done()
		close(added)
	}()

	return added, nil
}

// withLabels 设置标签选择器过滤 Pod
func withLabels(selector string) internalinterfaces.TweakListOptionsFunc {
	return func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector
	}
}

// podLoggerEventHandler Pod 事件处理器
type podLoggerEventHandler struct {
	cache.ResourceEventHandler
	added   chan<- string
	deleted chan<- string
}

// OnAdd Pod 新增时发送名称到通道
func (h *podLoggerEventHandler) OnAdd(obj interface{}, isInInitialList bool) {
	pod := obj.(*corev1.Pod)
	log.Printf("PodInformer: adding instance: %s", pod.Name)
	h.added <- pod.Name
}

// OnUpdate 空实现，日志无需处理更新事件
func (h *podLoggerEventHandler) OnUpdate(oldObj, newObj interface{}) {
}

// OnDelete 空实现，日志流会自动关闭
func (h *podLoggerEventHandler) OnDelete(obj interface{}) {
}
