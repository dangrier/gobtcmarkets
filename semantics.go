package btcmarkets

const APILocation = "https://api.btcmarkets.net"

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
	Bid        float64    `json:"bestBid"`
	Ask        float64    `json:"bestAsk"`
	Last       float64    `json:"lastPrice"`
	Currency   Currency   `json:"currency"`
	Instrument Instrument `json:"instrument"`
	Timestamp  int64      `json:"timestamp"`
	Volume     float64    `json:"volume24h"`
}

type MarketOrderbookData struct {
	Bids       [][]float64 `json:"bids"`
	Asks       [][]float64 `json:"asks"`
	Currency   Currency    `json:"currency"`
	Instrument Instrument  `json:"instrument"`
	Timestamp  int64       `json:"timestamp"`
}

type MarketTradesData []TradeData

type TradeData struct {
	TradeID   int64   `json:"tid"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"date"`
}

type AccountBalanceData struct {
	Currency Currency `json:"currency"`
	Balance  float64  `json:"balance"`
	Pending  float64  `json:"pendingFunds"`
}

type AccountTradingFeeData struct {
	TradingFee   float64 `json:"tradingFeeRate"`
	Volume30Days float64 `json:"volume30Day"`
}
