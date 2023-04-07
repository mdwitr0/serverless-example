package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

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
