package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"serverless-example/configs"
	"serverless-example/etherscan"
	"sort"
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

func Handler(ctx context.Context) {
	envConfig, err := configs.LoadEnvConfig()
	if err != nil {
		panic(err)
	}

	etherscanClient := etherscan.NewEtherscanClient(envConfig.EtherscanApiKey)

	latestBlockNumber, err := etherscanClient.GetLatestBlockNumber()
	if err != nil {
		fmt.Println("Error while getting the latest block number:", err)
		return
	}

	transactions, err := etherscanClient.GetTransactions(latestBlockNumber)
	if err != nil {
		fmt.Println("Error while getting transactions:", err)
		return
	}

	report := generateReport(transactions)
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Error while generating JSON report:", err)
		return
	}

	fmt.Println(string(reportJSON))
}

func generateReport(transactions []etherscan.Transaction) Report {
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

func normalizeTransactions(transactions []etherscan.Transaction) []NormalizedTransaction {
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
