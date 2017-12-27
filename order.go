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

	var roc RequestOrderCreate
	roc.Currency = currency
	roc.Instrument = instrument
	roc.Price = price
	roc.Volume = volume
	roc.ClientRequestID = "abc-cdf-1000"
	roc.OrderSide = side
	roc.OrderType = ordertype

	// Market orders should ignore prices, but add protection in case API changes.
	// A market bid is set to the lowest value, and ask to the highest
	if ordertype == Market && side == Bid {
		roc.Price = 1
	}
	if ordertype == Market && side == Ask {
		roc.Price = 99999900000000
	}

	// An AUD currency amount is not allowed to have more than two decimal places
	if currency == CurrencyAUD && math.Mod(float64(price), 1000000) != 0 {
		// If the third degree decimal onwards has a value, then return 0 (error)
		return OrderCreateData{}, errors.New("AUD currency only allows two decimal places")
	}

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jRoc, err := json.Marshal(roc)
	if err != nil {
		fmt.Println("error creating order object: " + err.Error())
		return OrderCreateData{}, errors.New("couldn't create object: " + err.Error())
	}
	sRoc := string(jRoc)
	reader := strings.NewReader(sRoc)

	ts := time.Now()
	signature := c.messageSignature("/order/create", ts, sRoc)

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
func (c *Client) OrderCancel(orderid int64) error {
	err := c.Limit10()
	if err != nil {
		return errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestOrderCancel
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
