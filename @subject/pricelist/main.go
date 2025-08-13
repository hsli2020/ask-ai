package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// ServerConfig 结构用于定义一台服务器的配置
type ServerConfig struct {
	Host       string // 地址和端口, e.g., "ftp.example.com:21" or "sftp.example.com:22"
	Protocol   string // "ftp" or "sftp"
	Username   string
	Password   string
	RemotePath string // 远程服务器上的文件完整路径, e.g., "/remote/data/file.zip"
	LocalPath  string
}

func main() {
	// 从 "config.json" 加载配置
	servers, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}
	// ======================================================

	for name, server := range servers {
		log.Printf("Connecting to: %s (%s)", name, server.Host)
		localPath, err := downloadFile(server)
		if err != nil {
			log.Printf("Failed to download %s from %s: %v", server.RemotePath, server.Host, err)
			continue // next server
		}

		log.Printf("File downloaded: %s", localPath)

		// 检查是否是zip文件并解压
		if strings.HasSuffix(strings.ToLower(localPath), ".zip") {
			log.Printf("Unzip file: %s", localPath)
			err := unzip(localPath, ".")
			if err != nil {
				log.Printf("Failed to unzip file %s: %v", localPath, err)
			} else {
				log.Printf("File unzipped: %s", localPath)
			}
		}
	}

	log.Println("End")
}

func LoadConfig(filename string) (map[string]ServerConfig, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file: %w", err)
	}

	var config map[string]ServerConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal JSON: %w", err)
	}

	return config, nil
}

// downloadFile 根据协议选择下载方式
func downloadFile(config ServerConfig) (string, error) {
	localFilename := filepath.Base(config.RemotePath)
	switch strings.ToLower(config.Protocol) {
	case "ftp":
		return localFilename, downloadFTP(config, localFilename)
	case "sftp":
		return localFilename, downloadSFTP(config, localFilename)
	default:
		return "", fmt.Errorf("Unknown protocol: %s", config.Protocol)
	}
}

// downloadFTP 处理FTP下载
func downloadFTP(config ServerConfig, localPath string) error {
	c, err := ftp.Dial(config.Host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("FTP connecttion failed: %w", err)
	}
	defer c.Quit()

	err = c.Login(config.Username, config.Password)
	if err != nil {
		return fmt.Errorf("FTP login failed: %w", err)
	}

	r, err := c.Retr(config.RemotePath)
	if err != nil {
		return fmt.Errorf("Failed to get remote file: %w", err)
	}
	defer r.Close()

	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("Failed to create local file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	if err != nil {
		return fmt.Errorf("Failed to copy file: %w", err)
	}

	return nil
}

// downloadSFTP 处理SFTP下载
func downloadSFTP(config ServerConfig, localPath string) error {
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		// 注意: 在生产环境中，您应该使用更安全的主机密钥验证策略
		// ssh.FixedHostKey(hostKey) or your own implementation
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := config.Host
	if !strings.Contains(addr, ":") {
		addr = addr + ":22" // 默认SFTP端口
	}

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("Unable to create SFTP client: %w", err)
	}
	defer client.Close()

	srcFile, err := client.Open(config.RemotePath)
	if err != nil {
		return fmt.Errorf("Unable to open remote file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("Unable to create local file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("Unable to copy file: %w", err)
	}

	return nil
}

// unzip 解压zip文件到指定目录
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// 检查路径穿越漏洞
		//if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
		//	return fmt.Errorf("不安全的zip路径: %s", fpath)
		//}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		// 在循环的最后关闭文件
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// FileCreatedToday 检查文件是否在今天创建
// 如果文件不存在或发生错误，则返回 false
func FileCreatedToday(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false // 文件不存在或发生其他错误
	}

	modTime := fileInfo.ModTime()
	now := time.Now()

	return modTime.Year() == now.Year() &&
		modTime.Month() == now.Month() &&
		modTime.Day() == now.Day()
}

// FileAge 计算文件自创建以来的时间
func FileAge(filePath string) (time.Duration, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("无法获取文件信息: %v", err)
	}

	// 获取文件的创建时间（在某些系统上可能返回修改时间）
	createTime := fileInfo.ModTime()

	// 计算从创建时间到现在的持续时间
	duration := time.Since(createTime)

	return duration, nil
}
