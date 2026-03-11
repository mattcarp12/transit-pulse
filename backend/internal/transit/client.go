package transit

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

// BART's public GTFS-RT endpoints
const (
	TripUpdatesURL = "http://api.bart.gov/gtfsrt/tripupdate.aspx"
	AlertsURL      = "http://api.bart.gov/gtfsrt/alerts.aspx"
)

// Client handles communication with the BART transit API
type Client struct {
	httpClient *http.Client
}

// NewClient initializes a new transit client with a sensible timeout.
// Never use the default http.Client in production as it has no timeout!
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchTripUpdates pulls the raw binary protobuf from BART and decodes it
// into a readable Go struct provided by the MobilityData bindings.
func (c *Client) FetchTripUpdates(ctx context.Context) (*gtfs.FeedMessage, error) {
	// 1. Create the HTTP request using the provided context (best practice for Lambda)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TripUpdatesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 2. Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trip updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	// 3. Read the raw binary stream into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 4. Initialize an empty FeedMessage and decode the binary data into it
	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(body, feed); err != nil {
		return nil, fmt.Errorf("failed to parse protobuf: %w", err)
	}

	return feed, nil
}