package btcmarkets

type RequestOrderCreate struct {
	Currency        Currency    `json:"currency"`
	Instrument      Instrument  `json:"instrument"`
	Price           AmountWhole `json:"price"`
	Volume          AmountWhole `json:"volume"`
	OrderSide       OrderSide   `json:"orderSide"`
	OrderType       OrderType   `json:"ordertype"`
	ClientRequestID string      `json:"clientRequestId"`
}

type RequestOrderCancel struct {
	Orders []int64 `json:"orderIds"`
}
