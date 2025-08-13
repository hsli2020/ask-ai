// 你是一个golang专家，帮我编写一个golang程序，完成以下功能：
// 一、从env文件中读入环境变量；
// 二、显示log信息，并将log文件存入文件中；
// 三、从csv文件中读入信息，是一些网址
// 四、对于每个从csv中读入的网址，下载对应的网页内容，并存入文件中，根据网址生成文件名

// 作为一个Golang专家，我来帮你编写这个程序。这个程序需要完成几个任务：
// 读取环境变量、记录日志、从CSV文件读取网址，并下载每个网址的网页内容并保存到文件中。
package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// - main.go
// - env.example
// - urls.csv

// env.example文件是一个示例env文件，用户可以复制它并重命名为.env来配置环境变量。
//
// 例如，env.example内容：
//
//   LOG_FILE=log.txt
//   CSV_FILE=urls.csv
//   OUTPUT_DIR=downloads
//
// 用户需要复制这个文件并重命名为.env，并根据需要修改变量值。
//
// 假设env文件每行都是key=value格式，使用=作为分隔符。
func loadEnvFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '=' // Assuming env file has key=value format

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if len(line) == 2 {
			key := line[0]
			value := line[1]
			os.Setenv(key, value)
		}
	}

	return nil
}

// 下载网页并保存到文件。
func downloadAndSave(url string, dir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to download %s: status code %d", url, resp.StatusCode)
	}

	fileName := generateFilename(url)
	filePath := filepath.Join(dir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Downloaded %s to %s\n", url, filePath)

	return nil
}

// generateFilename函数根据网址生成文件名。
// 一个简单的方法是使用网址的哈希值，每个网址都会有一个唯一的文件名。
func generateFilename(url string) string {
	hash := md5.Sum([]byte(url))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr + ".html"
}

func main() {
	envFilePath := ".env"

	// 检查.env文件是否存在，并读取它。
	if _, err := os.Stat(envFilePath); err == nil {
		loadEnvFromFile(envFilePath)
	}

	// 从环境变量中获取日志文件路径。
	logFilePath := os.Getenv("LOG_FILE")
	if logFilePath == "" {
		logFilePath = "log.txt"
	}

	// 创建日志写入器，日志会同时输出到控制台和指定的日志文件中。
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	// 读取CSV文件中的网址，从环境变量中获取CSV文件路径。
	csvFilePath := os.Getenv("CSV_FILE")
	if csvFilePath == "" {
		csvFilePath = "urls.csv"
	}

	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// 假设CSV文件每行只有一个网址。
	urls := make([]string, len(records))

	for i, record := range records {
		if len(record) > 0 {
			urls[i] = record[0]
		}
	}

	// 下载每个网址的网页内容，并保存到文件中。

	// 从环境变量中获取输出目录。
	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "downloads"
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 遍历所有网址并下载它们。
	for _, url := range urls {
		err := downloadAndSave(url, outputDir)
		if err != nil {
			log.Printf("Error downloading %s: %v", url, err)
		}
	}
}
