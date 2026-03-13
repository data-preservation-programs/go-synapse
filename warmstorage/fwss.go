package warmstorage

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const fwssABIJSON = `[
	{
		"type": "function",
		"name": "getServicePrice",
		"inputs": [],
		"outputs": [
			{
				"name": "pricing",
				"type": "tuple",
				"components": [
					{"name": "pricePerTiBPerMonthNoCDN", "type": "uint256"},
					{"name": "pricePerTiBCdnEgress", "type": "uint256"},
					{"name": "pricePerTiBCacheMissEgress", "type": "uint256"},
					{"name": "tokenAddress", "type": "address"},
					{"name": "epochsPerMonth", "type": "uint256"},
					{"name": "minimumPricePerMonth", "type": "uint256"}
				]
			}
		],
		"stateMutability": "view"
	}
]`

type FWSSContract struct {
	address common.Address
	abi     abi.ABI
	client  *ethclient.Client
}

func NewFWSSContract(address common.Address, client *ethclient.Client) (*FWSSContract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(fwssABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse FWSS ABI: %w", err)
	}

	return &FWSSContract{
		address: address,
		abi:     parsedABI,
		client:  client,
	}, nil
}

func (c *FWSSContract) GetServicePrice(ctx context.Context) (*ServicePrice, error) {
	data, err := c.abi.Pack("getServicePrice")
	if err != nil {
		return nil, fmt.Errorf("failed to pack getServicePrice call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getServicePrice: %w", err)
	}

	values, err := c.abi.Unpack("getServicePrice", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getServicePrice result: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("empty result from getServicePrice")
	}

	pricing, ok := values[0].(struct {
		PricePerTiBPerMonthNoCDN   *big.Int       `abi:"pricePerTiBPerMonthNoCDN"`
		PricePerTiBCdnEgress       *big.Int       `abi:"pricePerTiBCdnEgress"`
		PricePerTiBCacheMissEgress *big.Int       `abi:"pricePerTiBCacheMissEgress"`
		TokenAddress               common.Address `abi:"tokenAddress"`
		EpochsPerMonth             *big.Int       `abi:"epochsPerMonth"`
		MinimumPricePerMonth       *big.Int       `abi:"minimumPricePerMonth"`
	})
	if !ok {
		return nil, fmt.Errorf("unexpected type for getServicePrice result: %T", values[0])
	}

	return &ServicePrice{
		PricePerTiBPerMonthNoCDN:   pricing.PricePerTiBPerMonthNoCDN,
		PricePerTiBCDNEgress:       pricing.PricePerTiBCdnEgress,
		PricePerTiBCacheMissEgress: pricing.PricePerTiBCacheMissEgress,
		TokenAddress:               pricing.TokenAddress,
		EpochsPerMonth:             pricing.EpochsPerMonth,
		MinimumPricePerMonth:       pricing.MinimumPricePerMonth,
	}, nil
}
