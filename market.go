package btcmarkets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

/*
MARKET [HTTP GET ENDPOINTS]

Endpoints:
/market/BTC/AUD/tick
/market/BTC/AUD/orderbook
/market/BTC/AUD/trades

An optional GET parameter "since" can be used against the trades endpoint
*/

// MarketTick implements the /market/[instrument]/[currency]/tick endpoint.
func (c *Client) MarketTick(instrument Instrument, currency Currency) (MarketTickData, error) {
	err := c.Limit10()
	if err != nil {
		return MarketTickData{}, errors.New("error conducting rate limiting: " + err.Error())
	}
	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/tick", APILocation, instrument, currency))
	if err != nil {
		return MarketTickData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketTickData{}, err
	}

	var mtd MarketTickData
	err = json.Unmarshal(body, &mtd)
	if err != nil {
		return MarketTickData{}, err
	}

	return mtd, nil
}

// MarketOrderbook implements the /market/[instrument]/[currency]/orderbook endpoint.
func (c *Client) MarketOrderbook(instrument Instrument, currency Currency) (MarketOrderbookData, error) {
	err := c.Limit10()
	if err != nil {
		return MarketOrderbookData{}, errors.New("error conducting rate limiting: " + err.Error())
	}

	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/orderbook", APILocation, instrument, currency))
	if err != nil {
		return MarketOrderbookData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketOrderbookData{}, err
	}

	var mod MarketOrderbookData
	err = json.Unmarshal(body, &mod)
	if err != nil {
		return MarketOrderbookData{}, err
	}

	return mod, nil
}

// MarketTrades implements the /market/[instrument]/[currency]/trades endpoint.
//
// "since" is an optional parameter which, when greater than 0 will only get MarketTrades
// which occurred since the supplied trade ID.
func (c *Client) MarketTrades(instrument Instrument, currency Currency, since int64) (MarketTradesData, error) {
	err := c.Limit10()
	if err != nil {
		return MarketTradesData{}, errors.New("error conducting rate limiting: " + err.Error())
	}

	var uriSince string
	if since > 0 {
		uriSince = fmt.Sprintf("?since=%d", since)
	}

	res, err := netHTTPClient.Get(fmt.Sprintf("%s/market/%s/%s/trades%s", APILocation, instrument, currency, uriSince))
	if err != nil {
		return MarketTradesData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return MarketTradesData{}, err
	}

	var mtd MarketTradesData
	err = json.Unmarshal(body, &mtd)
	if err != nil {
		return MarketTradesData{}, err
	}

	return mtd, nil
}
