
package feeds_api

import "time"

// Error response returned when the request is unsuccessful.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ErrorList is a list of error responses returned when a request is unsuccessful.
type ErrorList struct {
	Errors []Error `json:"errors"`
}

// CreateFeedResponse is the response schema for the createFeed operation.
type CreateFeedResponse struct {
	FeedID string `json:"feedId"`
}

// Feed is detailed information about the feed.
type Feed struct {
	FeedID              string    `json:"feedId"`
	FeedType            string    `json:"feedType"`
	MarketplaceIDs      []string  `json:"marketplaceIds,omitempty"`
	CreatedTime         time.Time `json:"createdTime"`
	ProcessingStatus    string    `json:"processingStatus"`
	ProcessingStartTime time.Time `json:"processingStartTime,omitempty"`
	ProcessingEndTime   time.Time `json:"processingEndTime,omitempty"`
	ResultFeedDocumentID string    `json:"resultFeedDocumentId,omitempty"`
}

// GetFeedsResponse is the response schema for the getFeeds operation.
type GetFeedsResponse struct {
	Feeds     []Feed `json:"feeds"`
	NextToken string `json:"nextToken,omitempty"`
}

// FeedDocument is information required for the feed document.
type FeedDocument struct {
	FeedDocumentID      string `json:"feedDocumentId"`
	URL                 string `json:"url"`
	CompressionAlgorithm string `json:"compressionAlgorithm,omitempty"`
}

// FeedOptions is additional options to control the feed.
type FeedOptions map[string]string

// CreateFeedSpecification is information required to create the feed.
type CreateFeedSpecification struct {
	FeedType          string      `json:"feedType"`
	MarketplaceIDs    []string    `json:"marketplaceIds"`
	InputFeedDocumentID string      `json:"inputFeedDocumentId"`
	FeedOptions       FeedOptions `json:"feedOptions,omitempty"`
}

// CreateFeedDocumentSpecification specifies the content type for the createFeedDocument operation.
type CreateFeedDocumentSpecification struct {
	ContentType string `json:"contentType"`
}

// CreateFeedDocumentResponse is information required to upload a feed document's contents.
type CreateFeedDocumentResponse struct {
	FeedDocumentID string `json:"feedDocumentId"`
	URL            string `json:"url"`
}

// GetFeedsParams holds parameters for the GetFeeds operation.
type GetFeedsParams struct {
	FeedTypes        []string
	MarketplaceIDs   []string
	PageSize         int
	ProcessingStatuses []string
	CreatedSince     time.Time
	CreatedUntil     time.Time
	NextToken        string
}
