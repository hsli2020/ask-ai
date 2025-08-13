package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Feed struct {
	Header   Header    `json:"header"`
	Messages []Message `json:"messages"`
}

type Header struct {
	SellerID    string `json:"sellerId"`
	Version     string `json:"version"`
	IssueLocale string `json:"issueLocale"`
}

type Message struct {
	MessageID     int     `json:"messageId"`
	Sku           string  `json:"sku"`
	OperationType string  `json:"operationType"`
	ProductType   string  `json:"productType"`
	Patches       []Patch `json:"patches"`
}

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type FulfillmentAvailability struct {
	FulfillmentChannelCode string `json:"fulfillment_channel_code"`
	Quantity               int    `json:"quantity"`
}

type PurchasableOffer struct {
	Currency string     `json:"currency"`
	OurPrice []OurPrice `json:"our_price"`
}

type OurPrice struct {
	Schedule []Schedule `json:"schedule"`
}

type Schedule struct {
	ValueWithTax float64 `json:"value_with_tax"`
}

func main() {
	file, err := os.Open("data.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header row
	_, err = reader.Read()
	if err != nil {
		fmt.Println("Error reading header:", err)
		return
	}

	var messages []Message
	messageID := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading record:", err)
			continue
		}

		sku := record[0]
		price, _ := strconv.ParseFloat(record[1], 64)
		currency := record[2]
		quantity, _ := strconv.Atoi(record[3])

		messages = append(messages, Message{
			MessageID:     messageID,
			Sku:           sku,
			OperationType: "PATCH",
			ProductType:   "PRODUCT",
			Patches: []Patch{
				{
					Op:   "replace",
					Path: "/attributes/fulfillment_availability",
					Value: []FulfillmentAvailability{
						{
							FulfillmentChannelCode: "DEFAULT",
							Quantity:               quantity,
						},
					},
				},
				{
					Op:   "replace",
					Path: "/attributes/purchasable_offer",
					Value: []PurchasableOffer{
						{
							Currency: currency,
							OurPrice: []OurPrice{
								{
									Schedule: []Schedule{
										{
											ValueWithTax: price,
										},
									},
								},
							},
						},
					},
				},
			},
		})
		messageID++
	}

	feed := Feed{
		Header: Header{
			SellerID:    "AXXXXXXXXXXXX",
			Version:     "2.0",
			IssueLocale: "en_US",
		},
		Messages: messages,
	}

	jsonBytes, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	fmt.Println(string(jsonBytes))
}
