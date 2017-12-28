package btcmarkets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

/*
ACCOUNT [HTTP GET ENDPOINTS]

Endpoints:
/account/balance															(Rate Limited: 25x / 10sec)
/account/:instrument/:currency/tradingfee			(Rate Limited: 10x / 10sec)
*/

// AccountBalance implements the /account/balance endpoint.
func (c *Client) AccountBalance() ([]AccountBalanceData, error) {
	err := c.Limit10()
	if err != nil {
		return []AccountBalanceData{}, errors.New("error conducting rate limiting: " + err.Error())
	}
	ts := time.Now()
	signature := c.messageSignature("/account/balance", ts, "")

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
	err := c.Limit10()
	if err != nil {
		return AccountTradingFeeData{}, errors.New("error conducting rate limiting: " + err.Error())
	}
	ts := time.Now()
	ep := fmt.Sprintf("/account/%s/%s/tradingfee", instrument, currency)
	signature := c.messageSignature(ep, ts, "")

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

type AccountBalanceData struct {
	Currency Currency    `json:"currency"`
	Balance  AmountWhole `json:"balance"`
	Pending  AmountWhole `json:"pendingFunds"`
}

func (abd *AccountBalanceData) String() string {
	return fmt.Sprintf("%s: %f", abd.Currency, abd.Balance.ToAmountDecimal())
}

type AccountTradingFeeData struct {
	TradingFee   AmountWhole `json:"tradingFeeRate"`
	Volume30Days AmountWhole `json:"volume30Day"`
}
