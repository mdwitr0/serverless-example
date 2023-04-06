package etherscan

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

type Transaction struct {
	Hash  string `json:"hash"`
	Value string `json:"value"`
}

type Response struct {
	JsonRPC string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		Transactions []Transaction `json:"transactions"`
		TimeStamp    string        `json:"timeStamp"`
	} `json:"result"`
}

type EtherscanClient struct {
	apiKey string
	client *resty.Client
}

func NewEtherscanClient(apiKey string) *EtherscanClient {
	client := resty.New().SetBaseURL("https://api.etherscan.io/api").SetQueryParam("apikey", apiKey)

	return &EtherscanClient{
		apiKey: apiKey,
		client: client,
	}
}

func (c *EtherscanClient) GetTransactions(blockNumber int64) (transactions []Transaction, err error) {
	queryParams := map[string]string{
		"module":  "proxy",
		"action":  "eth_getBlockByNumber",
		"tag":     fmt.Sprintf("0x%x", blockNumber),
		"boolean": "true",
	}

	var response Response
	_, err = c.client.R().SetQueryParams(queryParams).SetResult(&response).Get("")
	if err != nil {
		return
	}

	transactions = response.Result.Transactions
	return
}

func (c *EtherscanClient) GetLatestBlockNumber() (blockNumber int64, err error) {
	queryParams := map[string]string{
		"module": "proxy",
		"action": "eth_blockNumber",
	}

	type latestBlockResponse struct {
		JsonRPC string `json:"jsonrpc"`
		Id      int    `json:"id"`
		Result  string `json:"result"`
	}

	var response latestBlockResponse
	_, err = c.client.R().SetQueryParams(queryParams).SetResult(&response).Get("")
	if err != nil {
		return
	}

	blockNumber, err = strconv.ParseInt(response.Result, 0, 64)
	return
}
