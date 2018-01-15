package btcmarkets

import (
	"errors"
)

/*
FUNDTRANSFER [HTTP ENDPOINTS]

Endpoints:
POST /fundtransfer/withdrawCrypto	(Rate Limited: 10x / 10sec)
POST /fundtransfer/withdrawEFT		(Rate Limited: 10x / 10sec)
GET /fundtransfer/history 			**preview only, not available in API yet**
*/

// FundTransferWithdrawCryptoRequest represents the required information for a crypto tranfer.
type FundTransferWithdrawCryptoRequest struct {
	Amount   AmountWhole `json:"amount"`
	Address  string      `json:"address"`
	Currency Currency    `json:"currency"`
}

// FundTransferWithdrawCryptoResponse represents the JSON data structure
// returned from the POST /fundtransfer/withdrawCrypto endpoint.
type FundTransferWithdrawCryptoResponse struct {
	Success        bool        `json:"success"`
	ErrorCode      string      `json:"errorCode"`
	ErrorMessage   string      `json:"errorMessage"`
	Status         string      `json:"status"`
	FundTransferID int64       `json:"fundTransferId"`
	Description    string      `json:"description"`
	Created        int64       `json:"creationTime"`
	Currency       Currency    `json:"currency"`
	Amount         AmountWhole `json:"amount"`
	Fee            AmountWhole `json:"fee"`
}

// WithdrawCrypto implements the POST /fundtransfer/withdrawCrypto API endpoint
func (c *Client) WithdrawCrypto(amount AmountWhole, currency Currency, address string) (*FundTransferWithdrawCryptoResponse, error) {
	ftwcReq := &FundTransferWithdrawCryptoRequest{
		Amount:   amount,
		Address:  address,
		Currency: currency,
	}

	ftwcRes := &FundTransferWithdrawCryptoResponse{}

	err := c.Post("/fundtransfer/withdrawCrypto", ftwcReq, ftwcRes, rateLimit10)
	if err != nil {
		return nil, err
	}

	return ftwcRes, nil
}

// FundTransferWithdrawEFTRequest represents the required information for an external EFT tranfer
type FundTransferWithdrawEFTRequest struct {
	AccountName   string      `json:"accountName"`
	AccountNumber string      `json:"accountNumber"`
	BankName      string      `json:"bankName"`
	BSB           string      `json:"bsbNumber"`
	Amount        AmountWhole `json:"amount"`
	Currency      Currency    `json:"currency"`
}

// FundTransferWithdrawEFTResponse represents the JSON data structure returned from
// the POST /fundtransfer/withdrawEFT endpoint. it sharres the same data structure
// as FundTransferWithdrawCryptoResponse.
type FundTransferWithdrawEFTResponse struct {
	FundTransferWithdrawCryptoResponse
}

// WithdrawEFT implements the POST /fundtransfer/withdrawEFT API endpoint
func (c *Client) WithdrawEFT(amount AmountWhole, currency Currency, accountName, accountNumber, bankName, bsb string) (*FundTransferWithdrawEFTResponse, error) {
	if currency != CurrencyAUD {
		return nil, errors.New("Only AUD currency is currently supported by the API/service")
	}

	ftweReq := &FundTransferWithdrawEFTRequest{
		AccountName:   accountName,
		AccountNumber: accountNumber,
		BankName:      bankName,
		BSB:           bsb,
		Amount:        amount,
		Currency:      currency,
	}

	ftweRes := &FundTransferWithdrawEFTResponse{}

	err := c.Post("/fundtransfer/withdrawEFT", ftweReq, ftweRes, rateLimit10)
	if err != nil {
		return nil, err
	}

	return ftweRes, nil
}
