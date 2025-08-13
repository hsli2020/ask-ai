
package feeds_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client for the Selling Partner API for Feeds
type Client struct {
	Endpoint   string
	HTTPClient *http.Client
}

// NewClient creates a new instance of the Client
func NewClient(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
	}
}


// GetFeeds returns feed details for the feeds that match the filters that you specify.
func (c *Client) GetFeeds(ctx context.Context, params GetFeedsParams) (*GetFeedsResponse, error) {
	req, err := c.newGetFeedsRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var getFeedsResponse GetFeedsResponse
	if err := json.NewDecoder(resp.Body).Decode(&getFeedsResponse); err != nil {
		return nil, err
	}

	return &getFeedsResponse, nil
}

func (c *Client) newGetFeedsRequest(ctx context.Context, params GetFeedsParams) (*http.Request, error) {
	u, err := url.Parse(c.Endpoint + "/feeds/2021-06-30/feeds")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if len(params.FeedTypes) > 0 {
		q.Set("feedTypes", strings.Join(params.FeedTypes, ","))
	}
	if len(params.MarketplaceIDs) > 0 {
		q.Set("marketplaceIds", strings.Join(params.MarketplaceIDs, ","))
	}
	if params.PageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", params.PageSize))
	}
	if len(params.ProcessingStatuses) > 0 {
		q.Set("processingStatuses", strings.Join(params.ProcessingStatuses, ","))
	}
	if !params.CreatedSince.IsZero() {
		q.Set("createdSince", params.CreatedSince.Format(time.RFC3339))
	}
	if !params.CreatedUntil.IsZero() {
		q.Set("createdUntil", params.CreatedUntil.Format(time.RFC3339))
	}
	if params.NextToken != "" {
		q.Set("nextToken", params.NextToken)
	}
	u.RawQuery = q.Encode()

	return http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
}


// CreateFeed creates a feed.
func (c *Client) CreateFeed(ctx context.Context, body CreateFeedSpecification) (*CreateFeedResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"/feeds/2021-06-30/feeds", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, c.handleErrorResponse(resp)
	}

	var createFeedResponse CreateFeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&createFeedResponse); err != nil {
		return nil, err
	}

	return &createFeedResponse, nil
}


// GetFeed returns feed details for the feed that you specify.
func (c *Client) GetFeed(ctx context.Context, feedID string) (*Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint+"/feeds/2021-06-30/feeds/"+feedID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var feed Feed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	return &feed, nil
}


// CancelFeed cancels the feed that you specify.
func (c *Client) CancelFeed(ctx context.Context, feedID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.Endpoint+"/feeds/2021-06-30/feeds/"+feedID, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}


// CreateFeedDocument creates a feed document for the feed type that you specify.
func (c *Client) CreateFeedDocument(ctx context.Context, body CreateFeedDocumentSpecification) (*CreateFeedDocumentResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"/feeds/2021-06-30/documents", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var createFeedDocumentResponse CreateFeedDocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&createFeedDocumentResponse); err != nil {
		return nil, err
	}

	return &createFeedDocumentResponse, nil
}

// GetFeedDocument returns the information required for retrieving a feed document's contents.
func (c *Client) GetFeedDocument(ctx context.Context, feedDocumentID string) (*FeedDocument, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint+"/feeds/2021-06-30/documents/"+feedDocumentID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var feedDocument FeedDocument
	if err := json.NewDecoder(resp.Body).Decode(&feedDocument); err != nil {
		return nil, err
	}

	return &feedDocument, nil
}
