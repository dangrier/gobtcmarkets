package btcmarkets

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// BaseURL is the protocol, and domain of the API to connect to.
	BaseURL = "https://api.btcmarkets.net"
)

var (
	b64        = base64.StdEncoding
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
}

// Client is the main struct type representing an interface with the API as a
// particular client user. It is stateless and obtains state/object information
// from API calls. The client is limited to 25 calls per 10 seconds on certain
// API endpoints, and 10 calls per 10 seconds on others. It is concurrency safe.
// (documented at https://github.com/BTCMarkets/API/wiki/faq)
type Client struct {
	apikey string
	secret []byte
	rate10 *rateLimit
}

// NewClient constructs a new Client for communicating with the BTC Markets API.
// Both your API key and secret are required. The secret should be provided as
// displayed in your BTC Markets account, as a base-64 encoded string.
func NewClient(key, secret string) (*Client, error) {
	if key == "" || secret == "" {
		return nil, errors.New("No key or secret provided")
	}

	binSecret, err := b64.DecodeString(secret)
	if err != nil {
		return nil, errors.New("Failed to decode secret. Secret should be a base-64 encoded string")
	}

	// Rate limiter for 10x/10sec (1/sec with 10x burst)
	rate10 := new(rateLimit)
	rate10.Start(time.Second, rateLimit10)

	return &Client{
		apikey: key,
		secret: binSecret,
		rate10: rate10,
	}, nil
}

// doRequest does all the heavy lifting of sending the HTTP request. It handles the
// attaching of certain authentication headers that are required with every request.
// It also handles request rate limiting.
func (c *Client) doRequest(req *http.Request, body string, v interface{}, rateLimit RateLimitValue) error {
	if rateLimit == rateLimit10 {
		err := c.Limit10()
		if err != nil {
			return fmt.Errorf("Error conducting rate limiting (%s)", err.Error())
		}
	}

	timeMillis := time.Now().UnixNano() / int64(time.Millisecond)
	timestamp := strconv.FormatInt(timeMillis, 10)

	h := hmac.New(sha512.New, []byte(c.secret))
	h.Write([]byte(fmt.Sprintf("%s\n%s\n%s", req.URL.RequestURI(), timestamp, body)))
	signature := b64.EncodeToString(h.Sum(nil))

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("signature", signature)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to execute request (%s)", err.Error())
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return fmt.Errorf("Failed to decode response (%s)", err.Error())
	}

	return nil
}

// Get handles a GET request to any BTC Markest API endpoint.
func (c *Client) Get(path string, v interface{}, rateLimit RateLimitValue) error {
	req, _, err := NewRequest("GET", BaseURL+path, nil)
	if err != nil {
		return err
	}

	err = c.doRequest(req, "", v, rateLimit)
	if err != nil {
		return err
	}

	return nil
}

// Post handles a POST request to any BTC Markest API endpoint.
func (c *Client) Post(path string, data interface{}, v interface{}, rateLimit RateLimitValue) error {
	req, body, err := NewRequest("POST", BaseURL+path, data)
	if err != nil {
		return err
	}

	err = c.doRequest(req, body, v, rateLimit)
	if err != nil {
		return err
	}

	return nil
}

// NewRequest returns a new HTTP request.
func NewRequest(method, path string, data interface{}) (req *http.Request, bodyString string, err error) {
	if method == "POST" && data != nil {
		// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
		// characters in the request body (results in an auth error)
		json, err := json.Marshal(data)
		if err != nil {
			return nil, "", fmt.Errorf("Failed to marshall POST request data (%s)", err.Error())
		}

		bodyString = string(json)
		reader := strings.NewReader(bodyString)

		req, err = http.NewRequest(method, path, reader)
	} else {
		req, err = http.NewRequest(method, path, nil)
	}

	if err != nil {
		return nil, "", fmt.Errorf("Failed to instantiate a new request (%s)", err.Error())
	}

	return
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
