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

func generateFilename(url string) string {
	hash := md5.Sum([]byte(url))
	hashStr := hex.EncodeToString(hash[:])
	return hashStr + ".html"
}

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

func main() {
	envFilePath := ".env"
	if _, err := os.Stat(envFilePath); err == nil {
		loadEnvFromFile(envFilePath)
	}

	logFilePath := os.Getenv("LOG_FILE")
	if logFilePath == "" {
		logFilePath = "log.txt"
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(io.MultiWriter(os.Stdout, file))

	csvFilePath := os.Getenv("CSV_FILE")
	if csvFilePath == "" {
		csvFilePath = "urls.csv"
	}

	file, err = os.Open(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	urls := make([]string, len(records))

	for i, record := range records {
		if len(record) > 0 {
			urls[i] = record[0]
		}
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "downloads"
	}

	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for _, url := range urls {
		err := downloadAndSave(url, outputDir)
		if err != nil {
			log.Printf("Error downloading %s: %v", url, err)
		}
	}
}
