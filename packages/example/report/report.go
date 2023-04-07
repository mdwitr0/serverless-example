package report

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"math/big"
	"os"
	"sort"
	"strconv"
)

type NormalizedTransaction struct {
	Hash  string  `json:"hash"`
	Value float64 `json:"value"`
}

type Report struct {
	TotalTransactions int                     `json:"total_transactions"`
	TotalValue        float64                 `json:"total_value"`
	Transactions      []NormalizedTransaction `json:"transactions"`
}

type Request struct {
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(in Request) (*Response, error) {
	envConfig, err := LoadEnvConfig()
	if err != nil {
		panic(err)
	}

	etherscanClient := NewEtherscanClient(envConfig.EtherscanApiKey)

	latestBlockNumber, _ := etherscanClient.GetLatestBlockNumber()

	transactions, _ := etherscanClient.GetTransactions(latestBlockNumber)

	report := generateReport(transactions)
	reportJSON, _ := json.Marshal(report)
	return &Response{
		StatusCode: 200,
		Body:       string(reportJSON),
	}, nil
}

func generateReport(transactions []Transaction) Report {
	humanReadableTransactions := normalizeTransactions(transactions)

	sort.Slice(humanReadableTransactions, func(i, j int) bool {
		return humanReadableTransactions[i].Value > humanReadableTransactions[j].Value
	})

	totalValue := 0.0
	for _, tx := range humanReadableTransactions {
		totalValue += tx.Value
	}

	return Report{
		TotalTransactions: len(transactions),
		TotalValue:        totalValue,
		Transactions:      humanReadableTransactions,
	}
}

func normalizeTransactions(transactions []Transaction) []NormalizedTransaction {
	humanReadableTransactions := make([]NormalizedTransaction, len(transactions))

	for i, tx := range transactions {
		value := new(big.Int)
		value.SetString(tx.Value[2:], 16)
		valueInEther := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(1e18))

		valueFloat, _ := valueInEther.Float64()

		humanReadableTransactions[i] = NormalizedTransaction{
			Hash:  tx.Hash,
			Value: valueFloat,
		}
	}

	return humanReadableTransactions
}

type Transaction struct {
	Hash  string `json:"hash"`
	Value string `json:"value"`
}

type EtherscanResponse struct {
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

	var response EtherscanResponse
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

type EnvConfig struct {
	EtherscanApiKey string
}

func LoadEnvConfig() (config EnvConfig, err error) {
	_ = godotenv.Load(".env")

	etherscanApiKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanApiKey == "" {
		err = fmt.Errorf("ETHERSCAN_API_KEY is not set")
		return
	}

	config = EnvConfig{
		EtherscanApiKey: etherscanApiKey,
	}
	return
}
