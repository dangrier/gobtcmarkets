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
	if k == "" || s == "" {
		return nil, errors.New("No key information supplied")
	}

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

func (c *Client) authenticateRequestMessage(endpoint string, timestamp time.Time, requestBody string) string {
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
	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/tick", APILocation, instrument, currency))
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

// MarketOrderbook implements the /market/[instrument]/[currency]/orderbook endpoint.
func (c *Client) MarketOrderbook(instrument Instrument, currency Currency) (MarketOrderbookData, error) {
	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/orderbook", APILocation, instrument, currency))
	if err != nil {
		return MarketOrderbookData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketOrderbookData{}, err
	}

	var mod MarketOrderbookData
	err = json.Unmarshal(body, &mod)
	if err != nil {
		return MarketOrderbookData{}, err
	}

	return mod, nil
}

// MarketTrades implements the /market/[instrument]/[currency]/trades endpoint.
//
// "since" is an optional parameter which, when greater than 0 will only get MarketTrades
// which occurred since the supplied trade ID.
func (c *Client) MarketTrades(instrument Instrument, currency Currency, since int64) (MarketTradesData, error) {
	var uriSince string
	if since > 0 {
		uriSince = fmt.Sprintf("?since=%d", since)
	}

	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/trades%s", APILocation, instrument, currency, uriSince))
	if err != nil {
		return MarketTradesData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketTradesData{}, err
	}

	var mtd MarketTradesData
	err = json.Unmarshal(body, &mtd)
	if err != nil {
		return MarketTradesData{}, err
	}

	return mtd, nil
}

/*
ACCOUNT [HTTP GET ENDPOINTS]

Endpoints:
/account/balance
/account/:instrument/:currency/tradingfee
*/

// AccountBalance implements the /account/balance endpoint.
func (c *Client) AccountBalance() ([]AccountBalanceData, error) {
	ts := time.Now()
	signature := c.authenticateRequestMessage("/account/balance", ts, "")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/account/balance", APILocation), nil)
	if err != nil {
		fmt.Println("error creating account balance request: " + err.Error())
		return []AccountBalanceData{}, errors.New("error creating account balance request: " + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving account balance response: " + err.Error())
		return []AccountBalanceData{}, errors.New("error receiving account balance response: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []AccountBalanceData{}, errors.New("error reading account balance response body: " + err.Error())
	}

	var accbal []AccountBalanceData
	err = json.Unmarshal(body, &accbal)
	if err != nil {
		return []AccountBalanceData{}, errors.New("error unmarshalling account balance body: " + err.Error())
	}

	return accbal, nil
}

// AccountTradingFee implements the /account/:instrument/:currency/tradingfee endpoint.
func (c *Client) AccountTradingFee(instrument Instrument, currency Currency) (AccountTradingFeeData, error) {
	ts := time.Now()
	ep := fmt.Sprintf("/account/%s/%s/tradingfee", instrument, currency)
	signature := c.authenticateRequestMessage(ep, ts, "")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", APILocation, ep), nil)
	if err != nil {
		fmt.Println("error creating account trading fee request: " + err.Error())
		return AccountTradingFeeData{}, errors.New("error creating account trading fee request: " + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving account trading fee response: " + err.Error())
		return AccountTradingFeeData{}, errors.New("error receiving account trading fee response: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return AccountTradingFeeData{}, errors.New("error reading account trading fee response body: " + err.Error())
	}

	var trafee AccountTradingFeeData
	err = json.Unmarshal(body, &trafee)
	if err != nil {
		return AccountTradingFeeData{}, errors.New("error unmarshalling account trading fee body: " + err.Error())
	}

	return trafee, nil
}

/*
FUNDTRANSFER [MIXED ENDPOINTS]

Endpoints:
/fundtransfer/withdrawCrypto
/fundtransfer/withdrawEFT
/fundtransfer/history

*/

// FundtransferHistory implements the /fundtransfer/history endpoint.
func (c *Client) FundtransferHistory() {
	// Not implemented in API - preview only!
}
