package btcmarkets

import (
	"bytes"
	"fmt"
)

/*
MARKET [HTTP ENDPOINTS]

Endpoints:
GET /market/BTC/AUD/tick
GET /market/BTC/AUD/orderbook
GET /market/BTC/AUD/trades
*/

// MarketTickResponse represents the JSON data structure returned from
// the GET /market/:instrument/:currency/tick endpoint.
type MarketTickResponse struct {
	Bid        AmountDecimal `json:"bestBid"`
	Ask        AmountDecimal `json:"bestAsk"`
	Last       AmountDecimal `json:"lastPrice"`
	Currency   Currency      `json:"currency"`
	Instrument Instrument    `json:"instrument"`
	Timestamp  int64         `json:"timestamp"`
	Volume     float64       `json:"volume24h"`
}

// MarketTick implements the GET /market/:instrument/:currency/tick endpoint.
func (c *Client) MarketTick(instrument Instrument, currency Currency) (*MarketTickResponse, error) {
	mtr := &MarketTickResponse{}

	err := c.Get(fmt.Sprintf("/market/%s/%s/tick", instrument, currency), mtr, rateLimit10)
	if err != nil {
		return nil, err
	}

	return mtr, nil
}

// MarketOrderbookResponse represents the JSON data structure returned from
// the GET /market/:instrument/:currency/orderbook endpoint.
type MarketOrderbookResponse struct {
	Bids       [][]float64 `json:"bids"`
	Asks       [][]float64 `json:"asks"`
	Currency   Currency    `json:"currency"`
	Instrument Instrument  `json:"instrument"`
	Timestamp  int64       `json:"timestamp"`
}

// MarketOrderbook implements the GET /market/:instrument/:currency/orderbook endpoint.
func (c *Client) MarketOrderbook(instrument Instrument, currency Currency) (*MarketOrderbookResponse, error) {
	mor := &MarketOrderbookResponse{}

	err := c.Get(fmt.Sprintf("/market/%s/%s/orderbook", instrument, currency), mor, rateLimit10)
	if err != nil {
		return nil, err
	}

	return mor, nil
}

// MarketTradesResponse represents the JSON data structure returned from
// the GET /market/:instrument/:currency/trades endpoint.
type MarketTradesResponse []MarketTradeDataItem

// MarketTrades implements the GET /market/:instrument/:currency/trades endpoint.
//
// "since" is an optional parameter which, when greater than 0 will only get MarketTrades
// which occurred since the supplied trade ID.
func (c *Client) MarketTrades(instrument Instrument, currency Currency, since OrderID) (*MarketTradesResponse, error) {
	var sinceURI string
	if since > 0 {
		sinceURI = fmt.Sprintf("?since=%d", since)
	}

	mtr := &MarketTradesResponse{}

	err := c.Get(fmt.Sprintf("/market/%s/%s/trades%s", instrument, currency, sinceURI), mtr, rateLimit10)
	if err != nil {
		// TODO: fix this error
		return nil, err
	}

	return mtr, nil
}

// Describe is a helper method for displaying trades in human-readable format.
func (mtr *MarketTradesResponse) Describe() string {
	var buff bytes.Buffer
	for _, el := range *mtr {
		buff.WriteString(fmt.Sprintf("%s\n", el.String()))
	}

	return buff.String()
}

// TradeID is a unique trade id.
type TradeID int64

// MarketTradeDataItem is the data structure that represents a single trade
type MarketTradeDataItem struct {
	TradeID   TradeID       `json:"tid"`
	Amount    AmountDecimal `json:"amount"`
	Price     AmountDecimal `json:"price"`
	Timestamp int64         `json:"date"`
}

// String is a helper method for displaying a trade in human-readable format.
func (mtdi *MarketTradeDataItem) String() string {
	return fmt.Sprintf("Trade %d: %f - %f at %f", mtdi.TradeID, mtdi.Amount*mtdi.Price, mtdi.Amount, mtdi.Price)
}
