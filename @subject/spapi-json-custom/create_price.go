// Gemini generated
package main

import (
	"encoding/json"
	"fmt"
)

// ValueWithTax corresponds to the innermost JSON object:
// { "value_with_tax": 6.99 }
type ValueWithTax struct {
	Value float64 `json:"value_with_tax"`
}

// PriceSchedule corresponds to the object containing the "schedule" array:
// { "schedule": [ ... ] }
type PriceSchedule struct {
	Schedule []ValueWithTax `json:"schedule"`
}

// Offer corresponds to the top-level JSON object.
// The fields are slices of PriceSchedule to match the JSON structure `[ { ... } ]`.
type Offer struct {
	ListPrice []PriceSchedule `json:"list_price"`
	MaxPrice  []PriceSchedule `json:"maximum_seller_allowed_price"`
	MinPrice  []PriceSchedule `json:"minimum_seller_allowed_price"`
	OurPrice  []PriceSchedule `json:"our_price"`
}

func main() {
	// 1. Create an instance of the Offer struct.
	offer := Offer{
		ListPrice: []PriceSchedule{
			{Schedule: []ValueWithTax{{Value: 6.99}}},
		},
		MaxPrice: []PriceSchedule{
			{Schedule: []ValueWithTax{{Value: 6.99}}},
		},
		MinPrice: []PriceSchedule{
			{Schedule: []ValueWithTax{{Value: 1.99}}},
		},
		OurPrice: []PriceSchedule{
			{Schedule: []ValueWithTax{{Value: 3.99}}},
		},
	}

	// 2. Marshal the struct into JSON.
	// We use MarshalIndent for pretty-printing, similar to the example file.
	jsonBytes, err := json.MarshalIndent(offer, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// 3. Print the resulting JSON string.
	fmt.Println(string(jsonBytes))
}
