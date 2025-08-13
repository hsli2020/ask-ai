package ordersv0

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

const (
	EndpointNorthAmerica = "https://sellingpartnerapi-na.amazon.com"
	EndpointEurope       = "https://sellingpartnerapi-eu.amazon.com"
	EndpointFarEast      = "https://sellingpartnerapi-fe.amazon.com"
)

// Client represents a client for the Selling Partner API for Orders.
type Client struct {
	Endpoint   string
	HTTPClient *http.Client
}

// NewClient creates a new client for the Selling Partner API for Orders.
func NewClient(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
	}
}

// GetOrders returns orders created or updated during the specified time period.
func (c *Client) GetOrders(ctx context.Context, params url.Values) (*GetOrdersResponse, error) {
	var getOrdersResponse GetOrdersResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders?%s", c.Endpoint, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrdersResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrdersResponse, nil
}

// GetOrder returns the order that you specify.
func (c *Client) GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error) {
	var getOrderResponse GetOrderResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders/%s", c.Endpoint, orderID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrderResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrderResponse, nil
}

// GetOrderBuyerInfo returns buyer information for the order that you specify.
func (c *Client) GetOrderBuyerInfo(ctx context.Context, orderID string) (*GetOrderBuyerInfoResponse, error) {
	var getOrderBuyerInfoResponse GetOrderBuyerInfoResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders/%s/buyerInfo", c.Endpoint, orderID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrderBuyerInfoResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrderBuyerInfoResponse, nil
}

// GetOrderAddress returns the shipping address for the order that you specify.
func (c *Client) GetOrderAddress(ctx context.Context, orderID string) (*GetOrderAddressResponse, error) {
	var getOrderAddressResponse GetOrderAddressResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders/%s/address", c.Endpoint, orderID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrderAddressResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrderAddressResponse, nil
}

// GetOrderItems returns detailed order item information for the order that you specify.
func (c *Client) GetOrderItems(ctx context.Context, orderID string, params url.Values) (*GetOrderItemsResponse, error) {
	var getOrderItemsResponse GetOrderItemsResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders/%s/orderItems?%s", c.Endpoint, orderID, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrderItemsResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrderItemsResponse, nil
}

// GetOrderItemsBuyerInfo returns buyer information for the order items in the order that you specify.
func (c *Client) GetOrderItemsBuyerInfo(ctx context.Context, orderID string, params url.Values) (*GetOrderItemsBuyerInfoResponse, error) {
	var getOrderItemsBuyerInfoResponse GetOrderItemsBuyerInfoResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/orders/v0/orders/%s/orderItems/buyerInfo?%s", c.Endpoint, orderID, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(body, &getOrderItemsBuyerInfoResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &getOrderItemsBuyerInfoResponse, nil
}

// UpdateShipmentStatus updates the shipment status for an order that you specify.
func (c *Client) UpdateShipmentStatus(ctx context.Context, orderID string, payload UpdateShipmentStatusRequest) error {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/orders/v0/orders/%s/shipment", c.Endpoint, orderID), bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
