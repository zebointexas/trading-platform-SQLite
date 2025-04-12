package kraken

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"time"

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

// GetPrice gets the price of a trading pair
func (c *Client) GetPrice(pair string) (float64, error) {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		fmt.Println("----------------- 000.32")

		// 检查 API 客户端是否初始化
		if c.api == nil {
			return 0, fmt.Errorf("Kraken API client is not initialized")
		}

		// 调用 Ticker 方法
		ticker, err := c.api.Ticker(pair)
		if err != nil {
			fmt.Printf("Retry %d: failed to fetch ticker for pair %s: %v\n", i+1, pair, err)
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}
		if ticker == nil {
			fmt.Printf("Retry %d: ticker response is nil for pair %s\n", i+1, pair)
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}

		fmt.Println("----------------- 000.33")
		fmt.Printf("Ticker response: %+v\n", ticker)

		// 获取交易对信息（GetPairTickerInfo 只返回 1 个值）
		pairInfo := ticker.GetPairTickerInfo(pair)

		// 检查 pairInfo 是否有效（通过 Close 字段判断）
		if len(pairInfo.Close) == 0 {
			fmt.Printf("Retry %d: ticker info is empty or pair %s not found\n", i+1, pair)
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}

		// 解析价格
		priceStr := pairInfo.Close[0]
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Printf("Retry %d: failed to parse price for pair %s: %v\n", i+1, pair, err)
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}

		return price, nil
	}

	return 0, fmt.Errorf("failed to get price for pair %s after %d retries", pair, maxRetries)
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
