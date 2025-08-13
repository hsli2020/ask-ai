package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Request represents the top-level JSON structure.
type Request struct {
	Header   Header    `json:"header"`
	Messages []Message `json:"messages"`
}

// Header represents the header part of the JSON.
type Header struct {
	SellerID    string `json:"sellerId"`
	Version     string `json:"version"`
	IssueLocale string `json:"issueLocale"`
}

// Message represents a single message in the messages array.
type Message struct {
	MessageID     int    `json:"messageId"`
	SKU           string `json:"sku"`
	OperationType string `json:"operationType"`
}

func main() {
	// Open the CSV file
	file, err := os.Open("skus.csv")
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip the header row
	if _, err := reader.Read(); err != nil {
		fmt.Println("Error reading header from CSV file:", err)
		return
	}
	var messages []Message
	messageID := 1

	// Read the CSV records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading CSV record:", err)
			continue
		}

		if len(record) > 0 {
			sku := record[0]
			messages = append(messages, Message{
				MessageID:     messageID,
				SKU:           sku,
				OperationType: "DELETE",
			})
			messageID++
		}
	}

	// Create the final request structure
	requestData := Request{
		Header: Header{
			SellerID:    "A123456789",
			Version:     "2.0",
			IssueLocale: "en_US",
		},
		Messages: messages,
	}

	// Marshal the data to JSON
	jsonData, err := json.MarshalIndent(requestData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Write the JSON to a file
	outputFilename := AddTimestampToFilename("output.json")
	err = os.WriteFile(outputFilename, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON file:", err)
		return
	}

	fmt.Println("Successfully generated output.json")
}
