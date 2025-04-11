package kraken

import (
	"errors"
	"reflect"
	"strconv"

	krakenapi "github.com/beldur/kraken-go-api-client"
)

type Client struct {
	api *krakenapi.KrakenAPI
}

func NewClient(apiKey, apiSecret string) *Client {
	api := krakenapi.New(apiKey, apiSecret)
	return &Client{api: api}
}

// Getter 方法，供外部访问
func (c *Client) NewClientAPI() *krakenapi.KrakenAPI {
	return c.api
}

// GetPrice gets the current price for the specified cryptocurrency pair
func (c *Client) GetPrice(pair string) (float64, error) {
	ticker, err := c.api.Ticker(pair)
	if err != nil {
		return 0, err
	}
	pairInfo := ticker.GetPairTickerInfo(pair)
	if len(pairInfo.Close) == 0 {
		return 0, errors.New("ticker not found for pair: " + pair)
	}
	priceStr := pairInfo.Close[0]
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

// GetBalance gets the account balance
func (c *Client) GetBalance() (map[string]float64, error) {
	balanceResponse, err := c.api.Balance()
	if err != nil {
		return nil, err
	}
	balances := make(map[string]float64)
	v := reflect.ValueOf(*balanceResponse)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Float()
		balances[field.Name] = value
	}
	return balances, nil
}

// PlaceOrder places an order
func (c *Client) PlaceOrder(pair, orderType, orderSide string, volume, price float64) (string, error) {
	volumeStr := strconv.FormatFloat(volume, 'f', -1, 64)
	priceStr := strconv.FormatFloat(price, 'f', -1, 64)

	opts := map[string]string{}
	if orderType == "limit" {
		opts["price"] = priceStr
	}

	response, err := c.api.AddOrder(pair, orderSide, orderType, volumeStr, opts)
	if err != nil {
		return "", err
	}

	if len(response.TransactionIds) > 0 {
		return response.TransactionIds[0], nil
	}

	return "", errors.New("no transaction ID returned")
}

// GetOrderStatus gets the status of an order
func (c *Client) GetOrderStatus(orderID string) (string, error) {
	args := map[string]string{}
	result, err := c.api.QueryOrders(orderID, args)
	if err != nil {
		return "", err
	}
	if orderInfo, ok := (*result)[orderID]; ok {
		return orderInfo.Status, nil
	}
	return "", errors.New("order not found")
}

// GetTradablePairs gets the tradable pairs
func (c *Client) GetTradablePairs() ([]string, error) {
	assetPairs, err := c.api.AssetPairs()
	if err != nil {
		return nil, err
	}
	pairs := []string{}
	v := reflect.ValueOf(*assetPairs)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		pairs = append(pairs, field.Name)
	}
	return pairs, nil
}
