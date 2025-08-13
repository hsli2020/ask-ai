# 用golang编写一个程序，用来监控一组github repos是否有新的版本发布，或者新的commits提交

我来为你编写一个GitHub仓库监控程序。这个程序可以监控指定仓库的新版本发布和新提交。

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Config 配置结构体
type Config struct {
	GitHubToken string   `json:"github_token"`
	Repos       []string `json:"repos"`
	CheckInterval int    `json:"check_interval"` // 检查间隔（秒）
}

// RepoStatus 仓库状态
type RepoStatus struct {
	Name          string
	LastRelease   string
	LastCommitSHA string
	LastCommitMsg string
	UpdatedAt     time.Time
}

// Monitor 监控器结构体
type Monitor struct {
	client *github.Client
	config Config
	status map[string]*RepoStatus
}

// NewMonitor 创建新的监控器
func NewMonitor(config Config) *Monitor {
	var client *github.Client
	
	if config.GitHubToken != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GitHubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}
	
	return &Monitor{
		client: client,
		config: config,
		status: make(map[string]*RepoStatus),
	}
}

// parseRepoName 解析仓库名称
func (m *Monitor) parseRepoName(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format: %s, should be 'owner/name'", repo)
	}
	return parts[0], parts[1], nil
}

// checkReleases 检查最新发布版本
func (m *Monitor) checkReleases(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	releases, _, err := m.client.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	}
	
	if len(releases) > 0 {
		return releases[0], nil
	}
	
	return nil, nil
}

// checkCommits 检查最新提交
func (m *Monitor) checkCommits(ctx context.Context, owner, repo string) (*github.RepositoryCommit, error) {
	commits, _, err := m.client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}
	
	if len(commits) > 0 {
		return commits[0], nil
	}
	
	return nil, nil
}

// checkRepo 检查单个仓库
func (m *Monitor) checkRepo(repo string) error {
	ctx := context.Background()
	owner, name, err := m.parseRepoName(repo)
	if err != nil {
		return err
	}
	
	// 获取最新发布
	latestRelease, err := m.checkReleases(ctx, owner, name)
	if err != nil {
		log.Printf("Error checking releases for %s: %v", repo, err)
	}
	
	// 获取最新提交
	latestCommit, err := m.checkCommits(ctx, owner, name)
	if err != nil {
		log.Printf("Error checking commits for %s: %v", repo, err)
	}
	
	// 获取当前状态
	currentStatus, exists := m.status[repo]
	
	// 检查是否有新版本发布
	if latestRelease != nil {
		if !exists || currentStatus.LastRelease != latestRelease.GetTagName() {
			log.Printf("🚨 NEW RELEASE DETECTED: %s - %s (Published: %s)", 
				repo, latestRelease.GetTagName(), latestRelease.GetPublishedAt().Format("2006-01-02 15:04:05"))
			if latestRelease.GetBody() != "" {
				log.Printf("   Release Notes: %s", latestRelease.GetBody())
			}
		}
	}
	
	// 检查是否有新提交
	if latestCommit != nil {
		if !exists || currentStatus.LastCommitSHA != latestCommit.GetSHA() {
			log.Printf("📝 NEW COMMIT DETECTED: %s - %s (%.8s)", 
				repo, latestCommit.GetCommit().GetMessage(), latestCommit.GetSHA())
		}
	}
	
	// 更新状态
	m.status[repo] = &RepoStatus{
		Name:          repo,
		LastRelease:   func() string { if latestRelease != nil { return latestRelease.GetTagName() }; return "" }(),
		LastCommitSHA: func() string { if latestCommit != nil { return latestCommit.GetSHA() }; return "" }(),
		LastCommitMsg: func() string { if latestCommit != nil { return latestCommit.GetCommit().GetMessage() }; return "" }(),
		UpdatedAt:     time.Now(),
	}
	
	return nil
}

// Run 开始监控
func (m *Monitor) Run() {
	log.Println("🚀 Starting GitHub repository monitor...")
	log.Printf("Monitoring %d repositories every %d seconds\n", len(m.config.Repos), m.config.CheckInterval)
	
	for {
		log.Println("🔍 Checking repositories...")
		
		for _, repo := range m.config.Repos {
			if err := m.checkRepo(repo); err != nil {
				log.Printf("❌ Error checking repo %s: %v", repo, err)
			}
		}
		
		log.Printf("✅ Check completed. Next check in %d seconds...\n", m.config.CheckInterval)
		time.Sleep(time.Duration(m.config.CheckInterval) * time.Second)
	}
}

// loadConfig 从文件加载配置
func loadConfig(filename string) (Config, error) {
	var config Config
	
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	
	return config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	config := Config{
		GitHubToken:   "",
		Repos: []string{
			"golang/go",
			"kubernetes/kubernetes",
			"docker/docker-ce",
		},
		CheckInterval: 300, // 5分钟
	}
	
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

func main() {
	// 检查配置文件是否存在
	configFile := "config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Println("Config file not found, creating default config...")
		if err := createDefaultConfig(); err != nil {
			log.Fatalf("Failed to create default config: %v", err)
		}
		log.Println("Default config created. Please edit config.json and add your GitHub token if needed.")
		return
	}
	
	// 加载配置
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// 如果没有设置token，给出警告
	if config.GitHubToken == "" {
		log.Println("⚠️  Warning: No GitHub token provided. You may hit rate limits.")
		log.Println("   To add a token, edit config.json and set 'github_token' field.")
	}
	
	// 创建监控器并运行
	monitor := NewMonitor(config)
	monitor.Run()
}
```

接下来创建一个配置文件示例 `config.json`：

```json
{
  "github_token": "your_github_token_here",
  "repos": [
    "golang/go",
    "kubernetes/kubernetes",
    "docker/docker-ce",
    "moby/moby",
    "gin-gonic/gin"
  ],
  "check_interval": 300
}
```

还需要创建 `go.mod` 文件：

```go
module github-monitor

go 1.21

require (
    github.com/google/go-github/v57 v57.0.0
    golang.org/x/oauth2 v0.14.0
)
```

为了更好的用户体验，我还提供一个增强版本，支持Webhook通知：

```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// NotificationConfig 通知配置
type NotificationConfig struct {
	WebhookURL string   `json:"webhook_url"`
	Email      EmailConfig `json:"email"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `json:"enabled"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// EnhancedConfig 增强配置
type EnhancedConfig struct {
	GitHubToken       string              `json:"github_token"`
	Repos             []string            `json:"repos"`
	CheckInterval     int                 `json:"check_interval"`
	Notifications     NotificationConfig  `json:"notifications"`
}

// EnhancedMonitor 增强监控器
type EnhancedMonitor struct {
	client *github.Client
	config EnhancedConfig
	status map[string]*RepoStatus
}

// sendWebhook 发送Webhook通知
func (m *EnhancedMonitor) sendWebhook(message string) error {
	if m.config.Notifications.WebhookURL == "" {
		return nil
	}
	
	payload := map[string]string{
		"text": message,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(m.config.Notifications.WebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

// sendEmail 发送邮件通知
func (m *EnhancedMonitor) sendEmail(subject, body string) error {
	if !m.config.Notifications.Email.Enabled {
		return nil
	}
	
	emailConfig := m.config.Notifications.Email
	
	// 创建邮件内容
	message := fmt.Sprintf(
		"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s",
		emailConfig.To, subject, body)
	
	// 连接SMTP服务器
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.SMTPHost)
	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	
	err := smtp.SendMail(addr, auth, emailConfig.From, []string{emailConfig.To}, []byte(message))
	if err != nil {
		return err
	}
	
	return nil
}

// Enhanced version of checkRepo with notifications
func (m *EnhancedMonitor) checkRepo(repo string) error {
	ctx := context.Background()
	owner, name, err := m.parseRepoName(repo)
	if err != nil {
		return err
	}
	
	// 获取最新发布
	latestRelease, err := m.checkReleases(ctx, owner, name)
	if err != nil {
		log.Printf("Error checking releases for %s: %v", repo, err)
	}
	
	// 获取最新提交
	latestCommit, err := m.checkCommits(ctx, owner, name)
	if err != nil {
		log.Printf("Error checking commits for %s: %v", repo, err)
	}
	
	// 获取当前状态
	currentStatus, exists := m.status[repo]
	
	var notifications []string
	
	// 检查是否有新版本发布
	if latestRelease != nil {
		if !exists || currentStatus.LastRelease != latestRelease.GetTagName() {
			msg := fmt.Sprintf("🚨 NEW RELEASE: %s - %s", repo, latestRelease.GetTagName())
			log.Println(msg)
			notifications = append(notifications, msg)
			
			// 发送通知
			m.sendWebhook(msg)
			m.sendEmail("GitHub Release Alert", msg)
		}
	}
	
	// 检查是否有新提交
	if latestCommit != nil {
		if !exists || currentStatus.LastCommitSHA != latestCommit.GetSHA() {
			msg := fmt.Sprintf("📝 NEW COMMIT: %s - %.8s", repo, latestCommit.GetSHA())
			log.Println(msg)
			if len(notifications) == 0 { // 只在没有其他通知时发送提交通知
				notifications = append(notifications, msg)
				m.sendWebhook(msg)
				m.sendEmail("GitHub Commit Alert", msg)
			}
		}
	}
	
	// 更新状态
	m.status[repo] = &RepoStatus{
		Name:          repo,
		LastRelease:   func() string { if latestRelease != nil { return latestRelease.GetTagName() }; return "" }(),
		LastCommitSHA: func() string { if latestCommit != nil { return latestCommit.GetSHA() }; return "" }(),
		LastCommitMsg: func() string { if latestCommit != nil { return latestCommit.GetCommit().GetMessage() }; return "" }(),
		UpdatedAt:     time.Now(),
	}
	
	return nil
}

// parseRepoName 解析仓库名称（复制自原Monitor）
func (m *EnhancedMonitor) parseRepoName(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format: %s, should be 'owner/name'", repo)
	}
	return parts[0], parts[1], nil
}

// checkReleases 检查最新发布版本（复制自原Monitor）
func (m *EnhancedMonitor) checkReleases(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	releases, _, err := m.client.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	}
	
	if len(releases) > 0 {
		return releases[0], nil
	}
	
	return nil, nil
}

// checkCommits 检查最新提交（复制自原Monitor）
func (m *EnhancedMonitor) checkCommits(ctx context.Context, owner, repo string) (*github.RepositoryCommit, error) {
	commits, _, err := m.client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}
	
	if len(commits) > 0 {
		return commits[0], nil
	}
	
	return nil, nil
}

// NewEnhancedMonitor 创建增强监控器
func NewEnhancedMonitor(config EnhancedConfig) *EnhancedMonitor {
	var client *github.Client
	
	if config.GitHubToken != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GitHubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}
	
	return &EnhancedMonitor{
		client: client,
		config: config,
		status: make(map[string]*RepoStatus),
	}
}

// Run 增强版运行方法
func (m *EnhancedMonitor) Run() {
	log.Println("🚀 Starting Enhanced GitHub repository monitor...")
	log.Printf("Monitoring %d repositories every %d seconds\n", len(m.config.Repos), m.config.CheckInterval)
	
	for {
		log.Println("🔍 Checking repositories...")
		
		for _, repo := range m.config.Repos {
			if err := m.checkRepo(repo); err != nil {
				log.Printf("❌ Error checking repo %s: %v", repo, err)
			}
		}
		
		log.Printf("✅ Check completed. Next check in %d seconds...\n", m.config.CheckInterval)
		time.Sleep(time.Duration(m.config.CheckInterval) * time.Second)
	}
}

// loadEnhancedConfig 加载增强配置
func loadEnhancedConfig(filename string) (EnhancedConfig, error) {
	var config EnhancedConfig
	
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	
	return config, nil
}

// createEnhancedDefaultConfig 创建增强版默认配置
func createEnhancedDefaultConfig() error {
	config := EnhancedConfig{
		GitHubToken:   "",
		Repos: []string{
			"golang/go",
			"kubernetes/kubernetes",
			"docker/docker-ce",
		},
		CheckInterval: 300,
		Notifications: NotificationConfig{
			WebhookURL: "",
			Email: EmailConfig{
				Enabled:  false,
				SMTPHost: "smtp.gmail.com",
				SMTPPort: 587,
				Username: "",
				Password: "",
				From:     "",
				To:       "",
			},
		},
	}
	
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}
```

## 使用说明

1. **安装依赖**：
```bash
go mod init github-monitor
go get github.com/google/go-github/v57
go get golang.org/x/oauth2
```

2. **配置文件**：
   - 运行程序会自动生成 `config.json`
   - 编辑配置文件添加要监控的仓库
   - 可选：添加GitHub token以避免API限制

3. **运行程序**：
```bash
go run main.go
```

## 功能特性

- ✅ 监控指定仓库的新版本发布
- ✅ 监控指定仓库的新提交
- ✅ 支持自定义检查间隔
- ✅ 支持GitHub API token认证
- ✅ 详细的日志输出
- ✅ 自动创建默认配置文件
- ✅ Webhook和邮件通知支持（增强版）

## 配置说明

- `github_token`: GitHub个人访问令牌（可选，但建议使用）
- `repos`: 要监控的仓库列表，格式为 `owner/repo`
- `check_interval`: 检查间隔（秒）
- `notifications`: 通知配置（增强版）

这个程序可以持续运行，定期检查指定的GitHub仓库，当检测到新版本或新提交时会输出日志并可选择发送通知。
