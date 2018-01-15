# Go BTC Markets

This package is a Go (Golang) effort to use the BTC Markets API, available at https://api.btcmarkets.net and documented on the [BTCMarkets/API](https://github.com/BTCMarkets/API) project. Please visit the API documentation to find out more about the service API. This is simply an API client.

This package is not affiliated with or created by BTC Markets Pty Ltd, but is intended to assist people in using their API with Go.

## Getting Started

You will need [Go](https://golang.org/) installed to use this package.

Then run a go get to grab the latest source for this package.

```bash
go get -u github.com/dangrier/gobtcmarkets
```

You will also need a public and private key pair generated from your account at [BTC Markets](https://btcmarkets.net).

### Installing / Usage

To use this package, include it in your imports.

```go
import "github.com/dangrier/gobtcmarkets"
```

The package itself is exported on the name `btcmarkets`.

#### Example - Basic Dumb Bot

```go
// Filename: basicbot.go
//
// Basic Bot finds out how many bitcoins it can buy with your available balance,
// then buys them, then waits to hit a profit threshold, sells the bitcoin, and
// moves the profit to your chosen account.
//
// Please be aware! If you run this - make sure you understand what it is doing first!
// There is also a chance that profit may never be realised, which is why this is a
// basic (dumb) bot.
//
// It may seem like some sections of this code will send a large amount of requests
// to the API, however the rate limiting will prevent this - all you have to worry
// about is making the API calls!
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dangrier/gobtcmarkets"
)

func main() {
	// Create a new btcmarkets.Client object
	cl, err := btcmarkets.NewClient("YOUR PUBLIC KEY HERE", "YOUR PRIVATE/SECRET KEY HERE")
	if err != nil {
		log.Fatal(err)
	}

	// Check your starting account balances
	bal, err := cl.AccountBalance()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the AUD balance from list of balances
	//
	// The ToAmountDecimal() method converts between API number forms Whole and Decimal.
	// If the method is not used to convert, the number will be out by an order of 10^8!
	aud := bal.GetBalance(btcmarkets.CurrencyAUD).ToAmountDecimal()

	fmt.Printf("%d is $%.2f\n", bal.GetBalance(btcmarkets.CurrencyAUD), aud)

	// Get the current trading fee
	fee, err := cl.AccountTradingFee(btcmarkets.InstrumentBitcoin, btcmarkets.CurrencyAUD)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fee is %.4fx\n", fee.TradingFee.ToAmountDecimal())

	// The usable amount of cash is the total amount minus the fee component
	var usable btcmarkets.AmountDecimal
	usable = (aud - (fee.TradingFee.ToAmountDecimal() * aud)).TrimCurrency()
	// Triming off the trailing decimal places after 2 places is required by the API

	fmt.Printf("Usable amount is $%.2f\n", usable)

	// Get the current market rate for AUD/BTC
	rate, err := cl.MarketTick(btcmarkets.InstrumentBitcoin, btcmarkets.CurrencyAUD)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Last $/BTC was $%.2f\n", rate.Last)

	// Determine how many coins to buy!
	coins := usable / rate.Last

	fmt.Printf("$%.2f is worth %.10fBTC\n", usable, coins)

	// Place an order for bitcoin
	order, err := cl.OrderCreate(btcmarkets.CurrencyAUD, btcmarkets.InstrumentBitcoin, rate.Last.ToAmountWhole(), coins.ToAmountWhole(), btcmarkets.Bid, btcmarkets.Market)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the order to be matched
	fmt.Printf("Order %d placed, awaiting fully matched state", order.ID)
	breakout := false
	for {
		m, _ := cl.OrderHistory(btcmarkets.CurrencyAUD, btcmarkets.InstrumentBitcoin, 10, 0)

		if !m.Success {
			log.Fatal("Something went wrong checking the order status! Please check manually!")
		}

		for _, o := range m.Orders {
			if o.OrderID == order.ID {
				if o.Status == btcmarkets.OrderStatusCancelled ||
					o.Status == btcmarkets.OrderStatusError ||
					o.Status == btcmarkets.OrderStatusFailed {
					log.Fatal(o.Error)
				}
				if o.Status == btcmarkets.OrderStatusFullyMatched {
					breakout = true
					fmt.Print("...MATCHED!\n\n")
					break
				} else {
					//fmt.Print(".")
					fmt.Printf("Status: %s\n", o.Status)
				}
			}
		}

		if breakout {
			break
		}
	}

	// YOU NOW HAVE BITCOIN!

	// Set the profit threshold for selling here
	profitThreshold := btcmarkets.AmountDecimal(0.01)
	fmt.Printf("Starting to check for profits with a threshold of %.2fx:\n", profitThreshold)
	var currentValue btcmarkets.MarketTickData
	for {
		// Check value of a bitcoin
		currentValue, err = cl.MarketTick(btcmarkets.InstrumentBitcoin, btcmarkets.CurrencyAUD)
		if err != nil {
			log.Fatal(err)
		}

		profit := (currentValue.Last * coins) - (currentValue.Last * coins * fee.TradingFee.ToAmountDecimal()) - usable

		if profit >= profitThreshold {
			break
		}

		fmt.Printf("%s - Profit: $%.2f (target: $%.2f)\n", time.Now().Format("2006-01-02 15:04:05"), profit, (profitThreshold * usable))
	}

	// YOU NOW HAVE ENOUGH PROFIT!

	// Sell the bitcoin
	sellOrder, err := cl.OrderCreate(btcmarkets.CurrencyAUD, btcmarkets.InstrumentBitcoin, rate.Last.ToAmountWhole(), coins.ToAmountWhole(), btcmarkets.Ask, btcmarkets.Market, "ABC123")
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the order to be matched
	fmt.Printf("\n\nSell order %d placed, awaiting fully matched state", sellOrder.ID)
	breakout = false
	for {
		m, _ := cl.OrderHistory(btcmarkets.CurrencyAUD, btcmarkets.InstrumentBitcoin, 10, 0)

		if !m.Success {
			log.Fatal("Something went wrong checking the order status! Please check manually!")
		}

		for _, o := range m.Orders {
			if o.OrderID == sellOrder.ID {
				if o.Status == btcmarkets.OrderStatusCancelled ||
					o.Status == btcmarkets.OrderStatusError ||
					o.Status == btcmarkets.OrderStatusFailed {
					log.Fatal(o.Error)
				}
				if o.Status == btcmarkets.OrderStatusFullyMatched {
					breakout = true
					fmt.Print("...MATCHED!\n\n")
					break
				} else {
					fmt.Print(".")
				}
			}
		}

		if breakout {
			break
		}
	}

	fmt.Printf("DONE! Made $%.2f profit!", profitThreshold*usable)
}
```

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/dangrier/gobtcmarkets/tags).

Versions prefixed with major version 0 (0.X.X) should be considered initial development versions and should not be relied upon for non-breaking changes.

## Authors

* **Dan Grier** - *Initial work* - [dangrier](https://github.com/dangrier)
* **Nick Law** - *Core contributor* - [nicklaw5](https://github.com/nicklaw5)

If you like to show your support for the work that went into creating this library feel free to send Dan a beer at [Bitcoin: 1CynPcMe1ZnHV3r2Zoi7snLMmj1RWSDsXy](https://blockchain.info/payment_request?address=1CynPcMe1ZnHV3r2Zoi7snLMmj1RWSDsXy)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
