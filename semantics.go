package btcmarkets

const APILocation = "https://api.btcmarkets.net"

type Currency string

const (
	AUD Currency = "AUD"
)

type Instrument string

const (
	Bcash      Instrument = "BCH"
	Bitcoin    Instrument = "BTC"
	Ethereum   Instrument = "ETH"
	EthClassic Instrument = "ETC"
	Litecoin   Instrument = "LTC"
	Ripple     Instrument = "XRP"
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
