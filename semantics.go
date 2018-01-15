package btcmarkets

import "math"

// AmountDecimal is a float type which represents the API numbers returned which
// can have decimal places.
//
// The AmountDecimal is 1/100000000 of an AmountWhole.
type AmountDecimal float64

// ToAmountWhole converts from AmountDecimal to AmountWhole
// by multiplication by 100000000 (used by API)
func (amount AmountDecimal) ToAmountWhole() AmountWhole {
	return AmountWhole(amount * 100000000)
}

func (amount AmountDecimal) TrimCurrency() AmountDecimal {
	return amount - AmountDecimal(math.Mod(float64(amount), 1))
}

// AmountWhole is an integer type which represents the API numbers returned
// which can have decimal places.
//
// The AmountWhole is 100000000x an AmountDecimal.
type AmountWhole int64

// ToAmountDecimal converts from AmountWhole to AmountDecimal
// by division by 100000000 (used by API)
func (amount AmountWhole) ToAmountDecimal() AmountDecimal {
	return AmountDecimal(amount) / AmountDecimal(100000000)
}

// Currency represents the name of a real-world or crypto currency
type Currency string

// Enumerated currencies.
const (
	CurrencyAUD        Currency = "AUD"
	CurrencyBcash      Currency = "BCH"
	CurrencyBitcoin    Currency = "BTC"
	CurrencyEthereum   Currency = "ETH"
	CurrencyEthClassic Currency = "ETC"
	CurrencyLitecoin   Currency = "LTC"
	CurrencyRipple     Currency = "XRP"
)

// Instrument represents the name of a crypto currency
type Instrument string

// Enumerated instruments.
const (
	InstrumentBcash      Instrument = "BCH"
	InstrumentBitcoin    Instrument = "BTC"
	InstrumentEthereum   Instrument = "ETH"
	InstrumentEthClassic Instrument = "ETC"
	InstrumentLitecoin   Instrument = "LTC"
	InstrumentRipple     Instrument = "XRP"
)
