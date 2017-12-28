package btcmarkets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

/*
ORDERS [HTTP POST ENDPOINTS]

Endpoints:
/order/create							(Rate Limited: 10x / 10sec)
/order/cancel							(Rate Limited: 25x / 10sec)
/order/history						(Rate Limited: 10x / 10sec)
/order/open								(Rate Limited: 25x / 10sec)
/order/trade/history			(Rate Limited: 10x / 10sec)
/order/detail							(Rate Limited: 25x / 10sec)
*/

// OrderCreate implements the /order/create endpoint.
func (c *Client) OrderCreate(currency Currency, instrument Instrument, price AmountWhole, volume AmountWhole, side OrderSide, ordertype OrderType) (OrderCreateData, error) {
	err := c.Limit10()
	if err != nil {
		return OrderCreateData{}, errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderCreate
	reqObject.Currency = currency
	reqObject.Instrument = instrument
	reqObject.Price = price
	reqObject.Volume = volume
	reqObject.ClientRequestID = "abc-cdf-1000"
	reqObject.OrderSide = side
	reqObject.OrderType = ordertype

	// Market orders should ignore prices, but add protection in case API changes.
	// A market bid is set to the lowest value, and ask to the highest
	if ordertype == Market && side == Bid {
		reqObject.Price = 1
	}
	if ordertype == Market && side == Ask {
		reqObject.Price = 99999900000000
	}

	// An AUD currency amount is not allowed to have more than two decimal places
	if currency == CurrencyAUD && math.Mod(float64(price), 1000000) != 0 {
		// If the third degree decimal onwards has a value, then return 0 (error)
		return OrderCreateData{}, errors.New("AUD currency only allows two decimal places")
	}

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jreqObject, err := json.Marshal(reqObject)
	if err != nil {
		fmt.Println("error creating order object: " + err.Error())
		return OrderCreateData{}, errors.New("couldn't create object: " + err.Error())
	}
	sreqObject := string(jreqObject)
	reader := strings.NewReader(sreqObject)

	ts := time.Now()
	signature := c.messageSignature("/order/create", ts, sreqObject)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/order/create", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return OrderCreateData{}, errors.New("couldn't create request: " + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving order response: " + err.Error())
		return OrderCreateData{}, errors.New("couldn't receive order response: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return OrderCreateData{}, errors.New("couldn't read response body: " + err.Error())
	}

	var orderResult OrderCreateData
	err = json.Unmarshal(body, &orderResult)
	if err != nil {
		return OrderCreateData{}, errors.New("couldn't unmarshal response: " + err.Error())
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !orderResult.Success {
		return OrderCreateData{}, errors.New("request error: " + orderResult.Error)
	}

	return orderResult, nil
}

// OrderCancel implements the /order/cancel endpoint.
//
// Takes the orderid of the order to be cancelled (ie. the same ID created in
// the OrderCreate result object's ID field)
//
// TODO: Update inconsistent error returns
func (c *Client) OrderCancel(orderid OrderID) error {
	err := c.Limit10()
	if err != nil {
		return errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderSpecific
	reqObject.Orders = append(reqObject.Orders, orderid)

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jsonObj, err := json.Marshal(reqObject)
	if err != nil {
		return err
	}
	sObj := string(jsonObj)
	reader := strings.NewReader(sObj)

	ts := time.Now()
	signature := c.messageSignature("/order/cancel", ts, sObj)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/order/cancel", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving order response: " + err.Error())
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var orderResult OrderCreateData
	err = json.Unmarshal(body, &orderResult)
	if err != nil {
		return err
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !orderResult.Success {
		return err
	}

	return nil
}

// OrderHistory implements the /order/history API endpoint
func (c *Client) OrderHistory(currency Currency, instrument Instrument, limit int, since OrderID) (OrderHistoryData, error) {
	err := c.Limit10()
	if err != nil {
		return OrderHistoryData{}, errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderHistory
	reqObject.Currency = currency
	reqObject.Instrument = instrument
	reqObject.Limit = limit
	reqObject.Since = since

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jsonObj, err := json.Marshal(reqObject)
	if err != nil {
		return OrderHistoryData{}, err
	}
	sObj := string(jsonObj)
	reader := strings.NewReader(sObj)

	ts := time.Now()
	signature := c.messageSignature("/order/history", ts, sObj)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/order/history", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return OrderHistoryData{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving order response: " + err.Error())
		return OrderHistoryData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return OrderHistoryData{}, err
	}

	fmt.Printf("\n\n%s\n\n", body)

	var orderResult OrderHistoryData
	err = json.Unmarshal(body, &orderResult)
	if err != nil {
		return OrderHistoryData{}, err
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !orderResult.Success {
		return OrderHistoryData{}, err
	}

	return orderResult, nil
}

// OrderOpen implements the /order/open API endpoint
func (c *Client) OrderOpen(currency Currency, instrument Instrument, limit int, since OrderID) (OrderHistoryData, error) {
	err := c.Limit10()
	if err != nil {
		return OrderHistoryData{}, errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderHistory
	reqObject.Currency = currency
	reqObject.Instrument = instrument
	reqObject.Limit = limit
	reqObject.Since = since

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jsonObj, err := json.Marshal(reqObject)
	if err != nil {
		return OrderHistoryData{}, err
	}
	sObj := string(jsonObj)
	reader := strings.NewReader(sObj)

	ts := time.Now()
	signature := c.messageSignature("/order/open", ts, sObj)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/order/open", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return OrderHistoryData{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving order response: " + err.Error())
		return OrderHistoryData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return OrderHistoryData{}, err
	}

	var orderResult OrderHistoryData
	err = json.Unmarshal(body, &orderResult)
	if err != nil {
		return OrderHistoryData{}, err
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !orderResult.Success {
		return OrderHistoryData{}, err
	}

	return orderResult, nil
}

// OrderDetail implements the /order/detail API endpoint
//
// This endpoint doesn't appear to give the details of an order as expected,
// instead it appears to indicate whether an order exists.
func (c *Client) OrderDetail(orderid OrderID) error {
	err := c.Limit10()
	if err != nil {
		return errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderSpecific
	reqObject.Orders = append(reqObject.Orders, orderid)

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jsonObj, err := json.Marshal(reqObject)
	if err != nil {
		return err
	}
	sObj := string(jsonObj)
	reader := strings.NewReader(sObj)

	ts := time.Now()
	signature := c.messageSignature("/order/cancel", ts, sObj)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/order/cancel", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		fmt.Println("error receiving order response: " + err.Error())
		return err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

type RequestOrderCreate struct {
	Currency        Currency    `json:"currency"`
	Instrument      Instrument  `json:"instrument"`
	Price           AmountWhole `json:"price"`
	Volume          AmountWhole `json:"volume"`
	OrderSide       OrderSide   `json:"orderSide"`
	OrderType       OrderType   `json:"ordertype"`
	ClientRequestID string      `json:"clientRequestId"`
}

type RequestOrderSpecific struct {
	Orders []OrderID `json:"orderIds"`
}

type RequestOrderHistory struct {
	Currency   Currency   `json:"currency"`
	Instrument Instrument `json:"instrument"`
	Limit      int        `json:"limit"`
	Since      OrderID    `json:"since"`
}

// OrderSide is used when creating an order to state whether the order is an
// ask (sell), or a bid (buy).
type OrderSide string

const (
	Ask OrderSide = "Ask"
	Bid OrderSide = "Bid"
)

// OrderType is used when creating an order to state whether the order is
// limited by the provided parameters, or whether to place the order at the
// market value.
type OrderType string

const (
	Limit  OrderType = "Limit"
	Market OrderType = "Market"
)

// OrderStatus is a string which describes the status of the order after being
// made.
type OrderStatus string

const (
	// OrderStatusNew is an order which is created but has not yet been placed
	OrderStatusNew = "New"

	// OrderStatusPlaced is a placed order which is unfilled
	OrderStatusPlaced = "Placed"

	// OrderStatusFailed is an order which has failed
	OrderStatusFailed = "Failed"

	// OrderStatusError is an order which has failed due to an error
	OrderStatusError = "Error"

	// OrderStatusCancelled is an order cancelled by the client
	OrderStatusCancelled = "Cancelled"

	// OrderStatusPartiallyCancelled is an order that has been partially
	// completed, but cancelled before fully matched / completed
	OrderStatusPartiallyCancelled = "Partially Cancelled"

	// OrderStatusFullyMatched is a completed successful order
	OrderStatusFullyMatched = "Fully Matched"

	// OrderStatusPartiallyMatched is a partially completed order for which some
	// of the instrument has been traded, but not enough to complete the order
	OrderStatusPartiallyMatched = "Partially Matched"
)

// OrderCreateData is used to structure the returned data from the OrderCreate
// method. ID is the order ID of the created order.
type OrderCreateData struct {
	Success bool    `json:"success"`
	Error   string  `json:"errorMessage"`
	ID      OrderID `json:"id"`
}

// OrderID is an integer representing the returned ID of a created order
type OrderID int64

type OrderHistoryData struct {
	Success bool                   `json:"success"`
	Error   string                 `json:"errorMessage"`
	Orders  []OrderHistoryDataItem `json:"orders"`
}

type OrderHistoryDataItem struct {
	OrderID    OrderID                `json:"id"`
	Currency   Currency               `json:"currency"`
	Instrument Instrument             `json:"instrument"`
	OrderSide  OrderSide              `json:"orderSide"`
	OrderType  OrderType              `json:"ordertype"`
	Created    int64                  `json:"creationTime"`
	Status     OrderStatus            `json:"status"`
	Error      string                 `json:"errorMessage"`
	Price      AmountWhole            `json:"price"`
	Volume     AmountWhole            `json:"volume"`
	VolumeOpen AmountWhole            `json:"openVolume"`
	Trades     []TradeHistoryDataItem `json:"trades"`
}

type TradeHistoryDataItem struct {
	TradeID     TradeID     `json:"id"`
	Created     int64       `json:"creationTime"`
	Description string      `json:"description"`
	Price       AmountWhole `json:"price"`
	Volume      AmountWhole `json:"volume"`
	Fee         AmountWhole `json:"fee"`
}
