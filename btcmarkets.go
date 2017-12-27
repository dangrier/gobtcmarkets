package btcmarkets

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var netHTTPClient = &http.Client{
	Timeout: time.Second * 10,
}

// Client is the main struct type representing an interface with the API as a
// particular client user. It is stateless and obtains state/object information
// from API calls. The client is limited to 25 calls per 10 seconds on certain
// API endpoints, and 10 calls per 10 seconds on others
// (documented at https://github.com/BTCMarkets/API/wiki/faq)
type Client struct {
	apikey string
	secret []byte
	rate10 *rateLimit
}

// NewClient constructs a new Client from provided key "k" and secret "s".
// "k" is the API key from BTC Markets.
// "s" is the private key.
// Both are expected in their base-64 encoded (default) state, as obtained from
// BTC Markets directly. Where strings are used, base-64 encoding is implied -
// otherwise []byte will represent the actual bytes.
func NewClient(k, s string) (*Client, error) {
	if k == "" || s == "" {
		return nil, errors.New("No key information supplied")
	}

	b64 := base64.StdEncoding

	binSecret, err := b64.DecodeString(s)
	if err != nil {
		return nil, errors.New("Key format error")
	}

	// Rate limiter for 10x/10sec (1/sec with 10x burst)
	rate10 := new(rateLimit)
	rate10.Start(time.Second, 10)

	return &Client{
		apikey: k,
		secret: binSecret,
		rate10: rate10,
	}, nil
}

func (c *Client) messageSignature(endpoint string, timestamp time.Time, requestBody string) string {
	// Convert time.Time into int64 milliseconds timestamp
	ts := timestamp.UnixNano() / int64(time.Millisecond)

	var constructedMessage string

	constructedMessage = fmt.Sprintf("%s\n%d\n%s", endpoint, ts, requestBody)

	b64 := base64.StdEncoding

	mac := hmac.New(sha512.New, c.secret)
	mac.Write([]byte(constructedMessage))
	signature := mac.Sum(nil)

	return b64.EncodeToString(signature)
}

// String is present to implement the Stringer interface for the Client type.
func (c *Client) String() string {
	return fmt.Sprintf("Client:{%s}", c.apikey)
}

// Limit10 performs rate limiting of 10x / 10secs
func (c *Client) Limit10() error {
	return c.rate10.Limit()
}

// Limit25 performs rate limiting of 25x / 10secs
/*func (c *Client) Limit25() error {
	return c.rate25.Limit()
}
*/
