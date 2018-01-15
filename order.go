package btcmarkets

import (
	"errors"
	"math"
)

/*
ORDERS [HTTP ENDPOINTS]

Endpoints:
POST /order/create			(Rate Limited: 10x / 10sec)
POST /order/cancel			(Rate Limited: 25x / 10sec)
POST /order/history			(Rate Limited: 10x / 10sec)
POST /order/open			(Rate Limited: 25x / 10sec)
POST /order/trade/history	(Rate Limited: 10x / 10sec)
POST /order/detail			(Rate Limited: 25x / 10sec)
*/

// OrderCreateRequest represents the JSON data structure sent to
// the POST /order/create endpoint.
type OrderCreateRequest struct {
	Currency        Currency    `json:"currency"`
	Instrument      Instrument  `json:"instrument"`
	Price           AmountWhole `json:"price"`
	Volume          AmountWhole `json:"volume"`
	OrderSide       OrderSide   `json:"orderSide"`
	OrderType       OrderType   `json:"ordertype"`
	ClientRequestID string      `json:"clientRequestId"`
}

// OrderCreateResponse represents the JSON data structure returned from
// the POST /order/create endpoint.
type OrderCreateResponse struct {
	Success         bool    `json:"success"`
	ErrorCode       int     `json:"errorCode"`
	ErrorMessage    string  `json:"errorMessage"`
	ID              OrderID `json:"id"`
	ClientRequestID string  `json:"clientRequestId"`
}

// OrderCreate implements the POST /order/create endpoint.
func (c *Client) OrderCreate(
	currency Currency,
	instrument Instrument,
	price AmountWhole,
	volume AmountWhole,
	side OrderSide,
	ordertype OrderType,
	requestID string,
) (*OrderCreateResponse, error) {
	rec := &OrderCreateRequest{
		Currency:        currency,
		Instrument:      instrument,
		Price:           price,
		Volume:          volume,
		ClientRequestID: requestID,
		OrderSide:       side,
		OrderType:       ordertype,
	}

	// Market orders should ignore prices, but add protection in case API changes.
	// A market bid is set to the lowest value, and ask to the highest
	if rec.OrderType == Market && rec.OrderSide == Bid {
		rec.Price = 1
	}
	if rec.OrderType == Market && rec.OrderSide == Ask {
		rec.Price = 99999900000000
	}

	// An AUD currency amount is not allowed to have more than two decimal places
	if rec.Currency == CurrencyAUD && math.Mod(float64(rec.Price), 1000000) != 0 {
		// If the third degree decimal onwards has a value, then return 0 (error)
		return nil, errors.New("AUD currency only allows two decimal places")
	}

	ocr := &OrderCreateResponse{}

	err := c.Post("/order/create", rec, ocr, rateLimit10)
	if err != nil {
		return nil, err
	}

	return ocr, nil
}

// OrderCancelResponse represents the JSON data structure returned from
// the POST /order/cancel endpoint.
type OrderCancelResponse struct {
	Success      bool              `json:"success"`
	ErrorCode    int               `json:"errorCode"`
	ErrorMessage string            `json:"errorMessage"`
	Responses    []OrderCancelData `json:"responses"`
}

// OrderCancelData represents the JSON data structure of a cancel order.
type OrderCancelData struct {
	Success      bool    `json:"success"`
	ErrorCode    int     `json:"errorCode"`
	ErrorMessage string  `json:"errorMessage"`
	ID           OrderID `json:"id"`
}

// OrderCancel implements the POST /order/cancel endpoint.
func (c *Client) OrderCancel(orderIDs ...OrderID) (*OrderCancelResponse, error) {
	ros := &OrdersSpecificRequest{
		Orders: orderIDs,
	}

	ocr := &OrderCancelResponse{}

	err := c.Post("/order/cancel", ros, ocr, rateLimit10)
	if err != nil {
		return nil, err
	}

	return ocr, nil
}

// OrderHistoryRequest represents the JSON data structure sent to
// the POST /order/history endpoint.
type OrderHistoryRequest struct {
	Currency   Currency   `json:"currency"`
	Instrument Instrument `json:"instrument"`
	Limit      int        `json:"limit"`
	Since      OrderID    `json:"since"`
}

// OrderHistoryResponse represents the JSON data structure returned from
// the POST /order/history endpoint. It shares the same data
// structure as OrderDetailResponse
type OrderHistoryResponse struct {
	OrderDetailResponse
}

// OrderHistory implements the POST /order/history API endpoint
func (c *Client) OrderHistory(currency Currency, instrument Instrument, limit int, since OrderID) (*OrderHistoryResponse, error) {
	ohReq := &OrderHistoryRequest{
		Currency:   currency,
		Instrument: instrument,
		Limit:      limit,
		Since:      since,
	}

	ohRes := &OrderHistoryResponse{}

	err := c.Post("/order/history", ohReq, ohRes, rateLimit10)
	if err != nil {
		return nil, err
	}

	return ohRes, nil
}

// OrderOpenRequest represents the JSON data structure sent to
// the POST /order/open endpoint. It shares the same data
// structure as OrderHistoryRequest
type OrderOpenRequest struct {
	OrderHistoryRequest
}

// OrderOpenResponse represents the JSON data structure returned from
// the POST /order/open endpoint. It shares the same data
// structure as OrderDetailResponse
type OrderOpenResponse struct {
	OrderDetailResponse
}

// OrderOpen implements the POST /order/open API endpoint
func (c *Client) OrderOpen(currency Currency, instrument Instrument, limit int, since OrderID) (*OrderOpenResponse, error) {
	roh := &OrderOpenRequest{}
	roh.Currency = currency
	roh.Instrument = instrument
	roh.Limit = limit
	roh.Since = since

	oor := &OrderOpenResponse{}

	err := c.Post("/order/open", roh, oor, rateLimit10)
	if err != nil {
		return nil, err
	}

	return oor, nil
}

// OrderDetailResponse represents the JSON data structure returned from
// the POST /order/detail endpoint.
type OrderDetailResponse struct {
	Success      bool            `json:"success"`
	ErrorCode    string          `json:"errorCode"`
	ErrorMessage string          `json:"errorMessage"`
	Orders       []OrderDataItem `josn:"orders"`
}

// OrderDetail implements the POST /order/detail API endpoint
func (c *Client) OrderDetail(orderIDs ...OrderID) (*OrderDetailResponse, error) {
	ros := &OrdersSpecificRequest{
		Orders: orderIDs,
	}

	odr := &OrderDetailResponse{}

	err := c.Post("/order/detail", ros, odr, rateLimit10)
	if err != nil {
		return nil, err
	}

	return odr, nil
}

// OrdersSpecificRequest is the data structure for sending multiple orders ids
// for endpoints that support that feature.
type OrdersSpecificRequest struct {
	Orders []OrderID `json:"orderIds"`
}

// OrderSide is used when creating an order to state whether the order is an
// ask (sell), or a bid (buy).
type OrderSide string

// Enumerated order sides
const (
	Ask OrderSide = "Ask"
	Bid OrderSide = "Bid"
)

// OrderType is used when creating an order to state whether the order is
// limited by the provided parameters, or whether to place the order at the
// market value.
type OrderType string

// Enumerated order types
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

// OrderID is an integer representing the returned ID of a created order
type OrderID int64

// OrderDataItem is the data structure that represents a single order
type OrderDataItem struct {
	OrderID      OrderID              `json:"id"`
	Currency     Currency             `json:"currency"`
	Instrument   Instrument           `json:"instrument"`
	OrderSide    OrderSide            `json:"orderSide"`
	OrderType    OrderType            `json:"ordertype"`
	Created      int64                `json:"creationTime"`
	Status       OrderStatus          `json:"status"`
	ErrorMessage string               `json:"errorMessage"`
	Price        AmountWhole          `json:"price"`
	Volume       AmountWhole          `json:"volume"`
	VolumeOpen   AmountWhole          `json:"openVolume"`
	Trades       []OrderTradeDataItem `json:"trades"`
}

// OrderTradeDataItem is the data structure that represents a single trade
type OrderTradeDataItem struct {
	TradeID     TradeID     `json:"id"`
	Created     int64       `json:"creationTime"`
	Description string      `json:"description"`
	Price       AmountWhole `json:"price"`
	Volume      AmountWhole `json:"volume"`
	Fee         AmountWhole `json:"fee"`
}
