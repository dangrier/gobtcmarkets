package btcmarkets

import (
	"bytes"
	"fmt"
)

const APILocation = "https://api.btcmarkets.net"

type AmountDecimal float64

// ToAmountWhole converts from AmountDecimal to AmountWhole
// by multiplication by 100000000 (used by API)
func (amount AmountDecimal) ToAmountWhole() AmountWhole {
	return AmountWhole(amount * 100000000)
}

type AmountWhole int64

// ToAmountDecimal converts from AmountWhole to AmountDecimal
// by division by 100000000 (used by API)
func (amount AmountWhole) ToAmountDecimal() AmountDecimal {
	return AmountDecimal(amount / 100000000)
}

type Currency string

const (
	CurrencyAUD        Currency = "AUD"
	CurrencyBcash      Currency = "BCH"
	CurrencyBitcoin    Currency = "BTC"
	CurrencyEthereum   Currency = "ETH"
	CurrencyEthClassic Currency = "ETC"
	CurrencyLitecoin   Currency = "LTC"
	CurrencyRipple     Currency = "XRP"
)

type Instrument string

const (
	InstrumentBcash      Instrument = "BCH"
	InstrumentBitcoin    Instrument = "BTC"
	InstrumentEthereum   Instrument = "ETH"
	InstrumentEthClassic Instrument = "ETC"
	InstrumentLitecoin   Instrument = "LTC"
	InstrumentRipple     Instrument = "XRP"
)

type OrderSide string

const (
	Ask OrderSide = "Ask"
	Bid OrderSide = "Bid"
)

type OrderType string

const (
	Limit  OrderType = "Limit"
	Market OrderType = "Market"
)

type MarketTickData struct {
	Bid        AmountDecimal `json:"bestBid"`
	Ask        AmountDecimal `json:"bestAsk"`
	Last       AmountDecimal `json:"lastPrice"`
	Currency   Currency      `json:"currency"`
	Instrument Instrument    `json:"instrument"`
	Timestamp  int64         `json:"timestamp"`
	Volume     float64       `json:"volume24h"`
}

type MarketOrderbookData struct {
	Bids       [][]float64 `json:"bids"`
	Asks       [][]float64 `json:"asks"`
	Currency   Currency    `json:"currency"`
	Instrument Instrument  `json:"instrument"`
	Timestamp  int64       `json:"timestamp"`
}

type MarketTradesData []TradeData

func (td *MarketTradesData) Describe() string {
	var buff bytes.Buffer
	for _, el := range *td {
		buff.WriteString(fmt.Sprintf("%s\n", el.String()))
	}
	return buff.String()
}

type TradeData struct {
	TradeID   int64         `json:"tid"`
	Amount    float64       `json:"amount"`
	Price     AmountDecimal `json:"price"`
	Timestamp int64         `json:"date"`
}

func (td *TradeData) String() string {
	return fmt.Sprintf("Trade %d: %f - %f at %f", td.TradeID, td.Amount*float64(td.Price), td.Amount, td.Price)
}

type AccountBalanceData struct {
	Currency Currency    `json:"currency"`
	Balance  AmountWhole `json:"balance"`
	Pending  float64     `json:"pendingFunds"`
}

func (abd *AccountBalanceData) String() string {
	return fmt.Sprintf("%s: %f", abd.Currency, abd.Balance.ToAmountDecimal())
}

type AccountTradingFeeData struct {
	TradingFee   AmountWhole `json:"tradingFeeRate"`
	Volume30Days AmountWhole `json:"volume30Day"`
}

type OrderStatus string

const (
	OrderStatusNew                = "New"
	OrderStatusPlaced             = "Placed"
	OrderStatusFailed             = "Failed"
	OrderStatusError              = "Error"
	OrderStatusCancelled          = "Cancelled"
	OrderStatusPartiallyCancelled = "Partially Cancelled"
	OrderStatusFullyMatched       = "Fully Matched"
	OrderStatusPartiallyMatched   = "Partially Matched"
)

type OrderCreateData struct {
	Success bool   `json:"success"`
	Error   string `json:"errorMessage"`
	ID      int64  `json:"id"`
}
