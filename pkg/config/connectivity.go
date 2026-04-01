package config

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/openfaas/faas-netes/version"
)

// ConnectivityCheck 检查控制器是否能通过 HTTPS 访问公网
// 用于商业使用授权验证
func ConnectivityCheck() error {
	req, err := http.NewRequest(http.MethodGet, "https://checkip.amazonaws.com", nil)
	if err != nil {
		return err
	}

	// 设置请求头 User-Agent，携带当前组件版本信息
	req.Header.Set("User-Agent", fmt.Sprintf("openfaas-ce/%s faas-netes", version.BuildVersion()))

	// 发送 HTTP 请求
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	// 校验响应状态码是否为 200 OK
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)

		return fmt.Errorf("unexpected status code checking connectivity: %d, body: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
