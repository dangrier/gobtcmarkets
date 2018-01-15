package btcmarkets

import (
	"fmt"
)

/*
ACCOUNT [HTTP ENDPOINTS]

Endpoints:
GET /account/balance							(Rate Limited: 25x / 10sec)
GET /account/:instrument/:currency/tradingfee	(Rate Limited: 10x / 10sec)
*/

// AccountBalanceResponse represents the JSON data structure sent to
// the GET /account/balance endpoint.
type AccountBalanceResponse []AccountBalanceItem

// AccountBalance implements the GET /account/balance endpoint.
func (c *Client) AccountBalance() (*AccountBalanceResponse, error) {
	abr := &AccountBalanceResponse{}

	err := c.Get("/account/balance", abr, rateLimit10)
	if err != nil {
		return nil, err
	}

	return abr, nil
}

// AccountTradingFeeResponse represents the JSON data structure sent to
// the GET /account/:instrument/:currency/tradingfee endpoint.
type AccountTradingFeeResponse struct {
	Success      bool        `json:"success"`
	ErrorCode    int         `json:"errorCode"`
	ErrorMessage string      `json:"errorMessage"`
	TradingFee   AmountWhole `json:"tradingFeeRate"`
	Volume30Days AmountWhole `json:"volume30Day"`
}

// AccountTradingFee implements the /account/:instrument/:currency/tradingfee endpoint.
func (c *Client) AccountTradingFee(instrument Instrument, currency Currency) (*AccountTradingFeeResponse, error) {
	atfd := &AccountTradingFeeResponse{}

	err := c.Get(fmt.Sprintf("/account/%s/%s/tradingfee", instrument, currency), atfd, rateLimit10)
	if err != nil {
		return nil, err
	}

	return atfd, nil
}

// GetBalance returns the balance of the provided currency.
func (a AccountBalanceResponse) GetBalance(currency Currency) AmountWhole {
	for _, b := range a {
		if b.Currency == currency {
			return b.Balance
		}
	}

	return AmountWhole(0)
}

// AccountBalanceItem is the data structure that represents currency account balance.
type AccountBalanceItem struct {
	Currency Currency    `json:"currency"`
	Balance  AmountWhole `json:"balance"`
	Pending  AmountWhole `json:"pendingFunds"`
}

func (ab *AccountBalanceItem) String() string {
	return fmt.Sprintf("%s: %f", ab.Currency, ab.Balance.ToAmountDecimal())
}
