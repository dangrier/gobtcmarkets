package btcmarkets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*
FUNDTRANSFER [MIXED ENDPOINTS]

Endpoints:
/fundtransfer/withdrawCrypto		(Rate Limited: 10x / 10sec)
/fundtransfer/withdrawEFT				(Rate Limited: 10x / 10sec)
/fundtransfer/history **preview only, not available in API yet**

*/

// WithdrawCrypto takes cash from your BTC Markets account and sends it to the specified crypto address
func (c *Client) WithdrawCrypto(amount AmountWhole, currency Currency, address string) error {
	err := c.Limit10()
	if err != nil {
		return errors.New("error conducting rate limiting: " + err.Error())
	}

	var reqObject RequestFundTransferWithdrawCrypto
	reqObject.Address = address
	reqObject.Amount = amount
	reqObject.Currency = currency

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jreqObject, err := json.Marshal(reqObject)
	if err != nil {
		return errors.New("couldn't create object: " + err.Error())
	}
	sreqObject := string(jreqObject)
	reader := strings.NewReader(sreqObject)

	ts := time.Now()
	signature := c.messageSignature("/fundtransfer/withdrawCrypto", ts, sreqObject)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/fundtransfer/withdrawCrypto", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return errors.New("couldn't create request: " + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		return errors.New("couldn't receive response: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("couldn't read response body: " + err.Error())
	}

	var withdrawResult WithdrawalData
	err = json.Unmarshal(body, &withdrawResult)
	if err != nil {
		return errors.New("couldn't unmarshal response: " + err.Error())
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !withdrawResult.Success {
		return errors.New("request error: " + withdrawResult.Error)
	}

	return nil
}

// WithdrawEFT takes cash from your BTC Markets account and sends it to the specified crypto address
func (c *Client) WithdrawEFT(amount AmountWhole, currency Currency, accountname string, accountnumber string, bankname string, bsb string) error {
	err := c.Limit10()
	if err != nil {
		return errors.New("error conducting rate limiting: " + err.Error())
	}

	if currency != CurrencyAUD {
		return errors.New("only AUD currency is currently supported by the API/service")
	}

	var reqObject RequestFundTransferWithdrawEFT
	reqObject.Amount = amount
	reqObject.Currency = currency
	reqObject.AccountName = accountname
	reqObject.AccountNumber = accountnumber
	reqObject.BankName = bankname
	reqObject.BSB = bsb

	// Have to use json.Marshal instead of Encoder, as API is sensitive to \n
	// characters in the request body (results in an auth error)
	jreqObject, err := json.Marshal(reqObject)
	if err != nil {
		return errors.New("couldn't create object: " + err.Error())
	}
	sreqObject := string(jreqObject)
	reader := strings.NewReader(sreqObject)

	ts := time.Now()
	signature := c.messageSignature("/fundtransfer/withdrawCrypto", ts, sreqObject)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/fundtransfer/withdrawCrypto", APILocation), reader)
	if err != nil {
		fmt.Println("error creating order request: " + err.Error())
		return errors.New("couldn't create request: " + err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("timestamp", fmt.Sprintf("%d", ts.UnixNano()/int64(time.Millisecond)))
	req.Header.Set("signature", signature)

	res, err := netHTTPClient.Do(req)
	if err != nil {
		return errors.New("couldn't receive response: " + err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("couldn't read response body: " + err.Error())
	}

	var withdrawResult WithdrawalData
	err = json.Unmarshal(body, &withdrawResult)
	if err != nil {
		return errors.New("couldn't unmarshal response: " + err.Error())
	}

	// The API should return 0 on an error - but this is not explicit.
	// Force this just in case
	if !withdrawResult.Success {
		return errors.New("request error: " + withdrawResult.Error)
	}

	return nil
}

// RequestFundTransferWithdrawCrypto represents the required information for a crypto tranfer
type RequestFundTransferWithdrawCrypto struct {
	Amount   AmountWhole `json:"amount"`
	Currency Currency    `json:"currency"`
	Address  string      `json:"address"`
}

// RequestFundTransferWithdrawEFT represents the required information for an external EFT tranfer
type RequestFundTransferWithdrawEFT struct {
	Amount        AmountWhole `json:"amount"`
	Currency      Currency    `json:"currency"`
	AccountName   string      `json:"accountName"`
	AccountNumber string      `json:"accountNumber"`
	BankName      string      `json:"bankName"`
	BSB           string      `json:"bsbNumber"`
}

// WithdrawalData represents the basic returned status of the withdrawal request - as the transfer
// would not be completed immediately. In future, the /fundtransfer/history API endpoint may assist.
type WithdrawalData struct {
	Success bool   `json:"success"`
	Error   string `json:"errorMessage"`
}
