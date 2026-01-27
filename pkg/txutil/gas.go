package txutil

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EstimateGasWithBuffer estimates gas and adds a safety buffer
func EstimateGasWithBuffer(ctx context.Context, client *ethclient.Client, msg ethereum.CallMsg, bufferPercent int) (uint64, error) {
	if bufferPercent < 0 || bufferPercent > 100 {
		return 0, fmt.Errorf("buffer percent must be between 0 and 100")
	}

	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	if bufferPercent > 0 {
		buffer := new(big.Int).Mul(big.NewInt(int64(gasLimit)), big.NewInt(int64(bufferPercent)))
		buffer.Div(buffer, big.NewInt(100))
		gasLimit += buffer.Uint64()
	}

	return gasLimit, nil
}

// GetGasPrice returns the current gas price with an optional multiplier
func GetGasPrice(ctx context.Context, client *ethclient.Client, multiplier float64) (*big.Int, error) {
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	if multiplier > 1.0 {
		gasPriceFloat := new(big.Float).SetInt(gasPrice)
		multiplierFloat := big.NewFloat(multiplier)
		gasPriceFloat.Mul(gasPriceFloat, multiplierFloat)
		gasPrice, _ = gasPriceFloat.Int(nil)
	}

	return gasPrice, nil
}

// GetGasTipCap returns the suggested gas tip cap (priority fee) for EIP-1559 transactions
func GetGasTipCap(ctx context.Context, client *ethclient.Client, multiplier float64) (*big.Int, error) {
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas tip cap: %w", err)
	}

	if multiplier > 1.0 {
		gasTipCapFloat := new(big.Float).SetInt(gasTipCap)
		multiplierFloat := big.NewFloat(multiplier)
		gasTipCapFloat.Mul(gasTipCapFloat, multiplierFloat)
		gasTipCap, _ = gasTipCapFloat.Int(nil)
	}

	return gasTipCap, nil
}
