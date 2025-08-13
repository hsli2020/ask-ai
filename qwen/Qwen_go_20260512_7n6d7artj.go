package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SendRequest 发送动态 HTTP 请求
// ctx: 上下文（控制超时/取消/Trace）
// client: HTTP 客户端（建议复用）
// method: HTTP 方法（GET, POST, PUT, DELETE, PATCH 等）
// apiURL: 目标地址（可包含已有参数）
// queryParams: 动态查询参数
// payload: 请求体，支持 io.Reader / []byte / string / 任意可 JSON 序列化的结构体或 nil
func SendRequest(ctx context.Context, client *http.Client, method, apiURL string, queryParams map[string]string, payload any) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	// 1. 解析 URL 并合并 Query 参数
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v) // 同名参数会覆盖，如需追加多值改用 q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	// 2. 智能转换 Payload 为 io.Reader
	var bodyReader io.Reader
	var autoJSON bool
	if payload != nil {
		switch p := payload.(type) {
		case io.Reader:
			bodyReader = p
		case []byte:
			bodyReader = bytes.NewReader(p)
		case string:
			bodyReader = strings.NewReader(p)
		default:
			// 默认尝试 JSON 序列化
			jsonData, err := json.Marshal(p)
			if err != nil {
				return nil, fmt.Errorf("marshal payload failed: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
			autoJSON = true
		}
	}

	// 3. 构建请求
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 4. 设置默认请求头
	req.Header.Set("Accept", "application/json")
	if autoJSON {
		req.Header.Set("Content-Type", "application/json")
	}
	// 注意：若传入的是 []byte/string，默认不强制设置 Content-Type，
	// 调用方可通过 req.Header.Set 自行覆盖

	// 5. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}