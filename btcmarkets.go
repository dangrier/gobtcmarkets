package btcmarkets

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
}

// NewClient constructs a new Client from provided key "k" and secret "s".
// "k" is the API key from BTC Markets.
// "s" is the private key.
// Both are expected in their base-64 encoded (default) state, as obtained from
// BTC Markets directly. Where strings are used, base-64 encoding is implied -
// otherwise []byte will represent the actual bytes.
func NewClient(k, s string) (*Client, error) {
	b64 := base64.StdEncoding

	binSecret, err := b64.DecodeString(s)
	if err != nil {
		return nil, errors.New("Key format error")
	}

	return &Client{
		apikey: k,
		secret: binSecret,
	}, nil
}

func (c *Client) authenticateRequestMessage(endpoint string, timestamp int64, requestBody string) string {
	constructedMessage := fmt.Sprintf("%s\n%d\n%s", endpoint, timestamp, requestBody)
	b64 := base64.StdEncoding

	mac := hmac.New(sha512.New, c.secret)
	mac.Write([]byte(constructedMessage))
	signature := mac.Sum(nil)

	return b64.EncodeToString(signature)
}

// String is present to implement the Stringer interface for the Client type.
func (c *Client) String() string {
	return fmt.Sprintf("{%s}", c.apikey)
}

/*
ORDERS [HTTP POST ENDPOINTS]

Possible status values:
- New
- Placed
- Failed
- Error
- Cancelled
- Partially Cancelled
- Fully Matched
- Partially Matched

Endpoints:
/order/create
/order/cancel
/order/history
/order/open
/order/trade/history
/order/detail
*/

// OrderCreate implements the /order/create endpoint.
func (c *Client) OrderCreate(currency Currency, instrument Instrument, price float64, volume float64) {

}

/*
MARKET [HTTP GET ENDPOINTS]

Endpoints:
/market/BTC/AUD/tick
/market/BTC/AUD/orderbook
/market/BTC/AUD/trades

An optional GET parameter "since" can be used against the trades endpoint
*/

// MarketTick implements the /market/[instrument]/[currency]/tick endpoint.
func (c *Client) MarketTick(instrument Instrument, currency Currency) (MarketTickData, error) {
	res, err := http.Get(fmt.Sprintf("%s/market/%s/%s/tick", APILocation, instrument, currency))
	if err != nil {
		return MarketTickData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketTickData{}, err
	}

	var mtd MarketTickData
	err = json.Unmarshal(body, &mtd)
	if err != nil {
		return MarketTickData{}, err
	}

	return mtd, nil
}

/*
ACCOUNT [HTTP GET ENDPOINTS]

Endpoints:
/account/balance
/account/:instrument/:currency/tradingfee
*/

/*
FUNDTRANSFER [MIXED ENDPOINTS]

Endpoints:
/fundtransfer/withdrawCrypto
/fundtransfer/withdrawEFT
/fundtransfer/history

*/
