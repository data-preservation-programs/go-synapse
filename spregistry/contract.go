package spregistry

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const SPRegistryABIJSON = `[
	{
		"type": "function",
		"name": "REGISTRATION_FEE",
		"inputs": [],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProvider",
		"inputs": [{"name": "providerId", "type": "uint256"}],
		"outputs": [
			{
				"name": "",
				"type": "tuple",
				"components": [
					{"name": "providerId", "type": "uint256"},
					{
						"name": "info",
						"type": "tuple",
						"components": [
							{"name": "serviceProvider", "type": "address"},
							{"name": "payee", "type": "address"},
							{"name": "name", "type": "string"},
							{"name": "description", "type": "string"},
							{"name": "isActive", "type": "bool"}
						]
					}
				]
			}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProviderByAddress",
		"inputs": [{"name": "addr", "type": "address"}],
		"outputs": [
			{
				"name": "",
				"type": "tuple",
				"components": [
					{"name": "providerId", "type": "uint256"},
					{
						"name": "info",
						"type": "tuple",
						"components": [
							{"name": "serviceProvider", "type": "address"},
							{"name": "payee", "type": "address"},
							{"name": "name", "type": "string"},
							{"name": "description", "type": "string"},
							{"name": "isActive", "type": "bool"}
						]
					}
				]
			}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProviderIdByAddress",
		"inputs": [{"name": "addr", "type": "address"}],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProviderWithProduct",
		"inputs": [
			{"name": "providerId", "type": "uint256"},
			{"name": "productType", "type": "uint8"}
		],
		"outputs": [
			{
				"name": "",
				"type": "tuple",
				"components": [
					{"name": "providerId", "type": "uint256"},
					{
						"name": "providerInfo",
						"type": "tuple",
						"components": [
							{"name": "serviceProvider", "type": "address"},
							{"name": "payee", "type": "address"},
							{"name": "name", "type": "string"},
							{"name": "description", "type": "string"},
							{"name": "isActive", "type": "bool"}
						]
					},
					{
						"name": "product",
						"type": "tuple",
						"components": [
							{"name": "productType", "type": "uint8"},
							{"name": "capabilityKeys", "type": "string[]"},
							{"name": "isActive", "type": "bool"}
						]
					},
					{"name": "productCapabilityValues", "type": "bytes[]"}
				]
			}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getAllActiveProviders",
		"inputs": [
			{"name": "offset", "type": "uint256"},
			{"name": "limit", "type": "uint256"}
		],
		"outputs": [
			{"name": "providerIds", "type": "uint256[]"},
			{"name": "hasMore", "type": "bool"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProvidersByProductType",
		"inputs": [
			{"name": "productType", "type": "uint8"},
			{"name": "onlyActive", "type": "bool"},
			{"name": "offset", "type": "uint256"},
			{"name": "limit", "type": "uint256"}
		],
		"outputs": [
			{
				"name": "providers",
				"type": "tuple[]",
				"components": [
					{"name": "providerId", "type": "uint256"},
					{
						"name": "providerInfo",
						"type": "tuple",
						"components": [
							{"name": "serviceProvider", "type": "address"},
							{"name": "payee", "type": "address"},
							{"name": "name", "type": "string"},
							{"name": "description", "type": "string"},
							{"name": "isActive", "type": "bool"}
						]
					},
					{
						"name": "product",
						"type": "tuple",
						"components": [
							{"name": "productType", "type": "uint8"},
							{"name": "capabilityKeys", "type": "string[]"},
							{"name": "isActive", "type": "bool"}
						]
					},
					{"name": "productCapabilityValues", "type": "bytes[]"}
				]
			},
			{"name": "providerIds", "type": "uint256[]"},
			{"name": "hasMore", "type": "bool"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "isProviderActive",
		"inputs": [{"name": "providerId", "type": "uint256"}],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "isRegisteredProvider",
		"inputs": [{"name": "addr", "type": "address"}],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getProviderCount",
		"inputs": [],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "activeProviderCount",
		"inputs": [],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "providerHasProduct",
		"inputs": [
			{"name": "providerId", "type": "uint256"},
			{"name": "productType", "type": "uint8"}
		],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "registerProvider",
		"inputs": [
			{"name": "payee", "type": "address"},
			{"name": "name", "type": "string"},
			{"name": "description", "type": "string"},
			{"name": "productType", "type": "uint8"},
			{"name": "capabilityKeys", "type": "string[]"},
			{"name": "capabilityValues", "type": "bytes[]"}
		],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "payable"
	},
	{
		"type": "function",
		"name": "updateProviderInfo",
		"inputs": [
			{"name": "name", "type": "string"},
			{"name": "description", "type": "string"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "removeProvider",
		"inputs": [],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "addProduct",
		"inputs": [
			{"name": "productType", "type": "uint8"},
			{"name": "capabilityKeys", "type": "string[]"},
			{"name": "capabilityValues", "type": "bytes[]"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "updateProduct",
		"inputs": [
			{"name": "productType", "type": "uint8"},
			{"name": "capabilityKeys", "type": "string[]"},
			{"name": "capabilityValues", "type": "bytes[]"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "removeProduct",
		"inputs": [{"name": "productType", "type": "uint8"}],
		"outputs": [],
		"stateMutability": "nonpayable"
	}
]`

type Contract struct {
	address common.Address
	abi     abi.ABI
	client  *ethclient.Client

	nonceMu     sync.Mutex
	nonce       uint64
	nonceLoaded bool
}

func NewContract(address common.Address, client *ethclient.Client) (*Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(SPRegistryABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SP registry ABI: %w", err)
	}

	return &Contract{
		address: address,
		abi:     parsedABI,
		client:  client,
	}, nil
}

func (c *Contract) Address() common.Address {
	return c.address
}

func (c *Contract) RegistrationFee(ctx context.Context) (*big.Int, error) {
	data, err := c.abi.Pack("REGISTRATION_FEE")
	if err != nil {
		return nil, fmt.Errorf("failed to pack REGISTRATION_FEE call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("REGISTRATION_FEE call failed: %w", err)
	}

	values, err := c.abi.Unpack("REGISTRATION_FEE", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack REGISTRATION_FEE result: %w", err)
	}

	fee, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for REGISTRATION_FEE: %T", values[0])
	}
	return fee, nil
}

type GetProviderResult struct {
	ProviderID *big.Int
	Info       RawProviderInfo
}

func (c *Contract) GetProvider(ctx context.Context, providerID *big.Int) (*GetProviderResult, error) {
	data, err := c.abi.Pack("getProvider", providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getProvider call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getProvider call failed: %w", err)
	}

	var res struct {
		ProviderID *big.Int `abi:"providerId"`
		Info       struct {
			ServiceProvider common.Address `abi:"serviceProvider"`
			Payee           common.Address `abi:"payee"`
			Name            string         `abi:"name"`
			Description     string         `abi:"description"`
			IsActive        bool           `abi:"isActive"`
		} `abi:"info"`
	}

	err = c.abi.UnpackIntoInterface(&res, "getProvider", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getProvider result: %w", err)
	}

	return &GetProviderResult{
		ProviderID: res.ProviderID,
		Info: RawProviderInfo{
			ServiceProvider: res.Info.ServiceProvider,
			Payee:           res.Info.Payee,
			Name:            res.Info.Name,
			Description:     res.Info.Description,
			IsActive:        res.Info.IsActive,
		},
	}, nil
}

func (c *Contract) GetProviderByAddress(ctx context.Context, addr common.Address) (*GetProviderResult, error) {
	data, err := c.abi.Pack("getProviderByAddress", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getProviderByAddress call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getProviderByAddress call failed: %w", err)
	}

	var res struct {
		ProviderID *big.Int `abi:"providerId"`
		Info       struct {
			ServiceProvider common.Address `abi:"serviceProvider"`
			Payee           common.Address `abi:"payee"`
			Name            string         `abi:"name"`
			Description     string         `abi:"description"`
			IsActive        bool           `abi:"isActive"`
		} `abi:"info"`
	}

	err = c.abi.UnpackIntoInterface(&res, "getProviderByAddress", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getProviderByAddress result: %w", err)
	}

	return &GetProviderResult{
		ProviderID: res.ProviderID,
		Info: RawProviderInfo{
			ServiceProvider: res.Info.ServiceProvider,
			Payee:           res.Info.Payee,
			Name:            res.Info.Name,
			Description:     res.Info.Description,
			IsActive:        res.Info.IsActive,
		},
	}, nil
}

func (c *Contract) GetProviderIDByAddress(ctx context.Context, addr common.Address) (*big.Int, error) {
	data, err := c.abi.Pack("getProviderIdByAddress", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getProviderIdByAddress call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getProviderIdByAddress call failed: %w", err)
	}

	values, err := c.abi.Unpack("getProviderIdByAddress", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getProviderIdByAddress result: %w", err)
	}

	id, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for provider ID: %T", values[0])
	}
	return id, nil
}

type GetProviderWithProductResult struct {
	ProviderID               *big.Int
	ProviderInfo             RawProviderInfo
	Product                  RawProduct
	ProductCapabilityValues  [][]byte
}

func (c *Contract) GetProviderWithProduct(ctx context.Context, providerID *big.Int, productType uint8) (*GetProviderWithProductResult, error) {
	data, err := c.abi.Pack("getProviderWithProduct", providerID, productType)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getProviderWithProduct call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getProviderWithProduct call failed: %w", err)
	}

	var res struct {
		ProviderID   *big.Int `abi:"providerId"`
		ProviderInfo struct {
			ServiceProvider common.Address `abi:"serviceProvider"`
			Payee           common.Address `abi:"payee"`
			Name            string         `abi:"name"`
			Description     string         `abi:"description"`
			IsActive        bool           `abi:"isActive"`
		} `abi:"providerInfo"`
		Product struct {
			ProductType    uint8    `abi:"productType"`
			CapabilityKeys []string `abi:"capabilityKeys"`
			IsActive       bool     `abi:"isActive"`
		} `abi:"product"`
		ProductCapabilityValues [][]byte `abi:"productCapabilityValues"`
	}

	err = c.abi.UnpackIntoInterface(&res, "getProviderWithProduct", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getProviderWithProduct result: %w", err)
	}

	return &GetProviderWithProductResult{
		ProviderID: res.ProviderID,
		ProviderInfo: RawProviderInfo{
			ServiceProvider: res.ProviderInfo.ServiceProvider,
			Payee:           res.ProviderInfo.Payee,
			Name:            res.ProviderInfo.Name,
			Description:     res.ProviderInfo.Description,
			IsActive:        res.ProviderInfo.IsActive,
		},
		Product: RawProduct{
			ProductType:    res.Product.ProductType,
			CapabilityKeys: res.Product.CapabilityKeys,
			IsActive:       res.Product.IsActive,
		},
		ProductCapabilityValues: res.ProductCapabilityValues,
	}, nil
}

func (c *Contract) GetAllActiveProviders(ctx context.Context, offset, limit *big.Int) ([]*big.Int, bool, error) {
	data, err := c.abi.Pack("getAllActiveProviders", offset, limit)
	if err != nil {
		return nil, false, fmt.Errorf("failed to pack getAllActiveProviders call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, false, fmt.Errorf("getAllActiveProviders call failed: %w", err)
	}

	values, err := c.abi.Unpack("getAllActiveProviders", result)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unpack getAllActiveProviders result: %w", err)
	}

	providerIDs, ok := values[0].([]*big.Int)
	if !ok {
		return nil, false, fmt.Errorf("unexpected type for provider IDs: %T", values[0])
	}
	hasMore, ok := values[1].(bool)
	if !ok {
		return nil, false, fmt.Errorf("unexpected type for hasMore: %T", values[1])
	}

	return providerIDs, hasMore, nil
}

func (c *Contract) IsProviderActive(ctx context.Context, providerID *big.Int) (bool, error) {
	data, err := c.abi.Pack("isProviderActive", providerID)
	if err != nil {
		return false, fmt.Errorf("failed to pack isProviderActive call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("isProviderActive call failed: %w", err)
	}

	values, err := c.abi.Unpack("isProviderActive", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack isProviderActive result: %w", err)
	}

	active, ok := values[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for isProviderActive: %T", values[0])
	}
	return active, nil
}

func (c *Contract) IsRegisteredProvider(ctx context.Context, addr common.Address) (bool, error) {
	data, err := c.abi.Pack("isRegisteredProvider", addr)
	if err != nil {
		return false, fmt.Errorf("failed to pack isRegisteredProvider call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("isRegisteredProvider call failed: %w", err)
	}

	values, err := c.abi.Unpack("isRegisteredProvider", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack isRegisteredProvider result: %w", err)
	}

	registered, ok := values[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for isRegisteredProvider: %T", values[0])
	}
	return registered, nil
}

func (c *Contract) GetProviderCount(ctx context.Context) (*big.Int, error) {
	data, err := c.abi.Pack("getProviderCount")
	if err != nil {
		return nil, fmt.Errorf("failed to pack getProviderCount call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getProviderCount call failed: %w", err)
	}

	values, err := c.abi.Unpack("getProviderCount", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getProviderCount result: %w", err)
	}

	count, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for getProviderCount: %T", values[0])
	}
	return count, nil
}

func (c *Contract) ActiveProviderCount(ctx context.Context) (*big.Int, error) {
	data, err := c.abi.Pack("activeProviderCount")
	if err != nil {
		return nil, fmt.Errorf("failed to pack activeProviderCount call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("activeProviderCount call failed: %w", err)
	}

	values, err := c.abi.Unpack("activeProviderCount", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack activeProviderCount result: %w", err)
	}

	count, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for activeProviderCount: %T", values[0])
	}
	return count, nil
}

func (c *Contract) ProviderHasProduct(ctx context.Context, providerID *big.Int, productType uint8) (bool, error) {
	data, err := c.abi.Pack("providerHasProduct", providerID, productType)
	if err != nil {
		return false, fmt.Errorf("failed to pack providerHasProduct call: %w", err)
	}

	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c.address,
		Data: data,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("providerHasProduct call failed: %w", err)
	}

	values, err := c.abi.Unpack("providerHasProduct", result)
	if err != nil {
		return false, fmt.Errorf("failed to unpack providerHasProduct result: %w", err)
	}

	hasProduct, ok := values[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type for providerHasProduct: %T", values[0])
	}
	return hasProduct, nil
}

func (c *Contract) RegisterProvider(opts *bind.TransactOpts, payee common.Address, name, description string, productType uint8, capabilityKeys []string, capabilityValues [][]byte) (*types.Transaction, error) {
	data, err := c.abi.Pack("registerProvider", payee, name, description, productType, capabilityKeys, capabilityValues)
	if err != nil {
		return nil, fmt.Errorf("failed to pack registerProvider call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) UpdateProviderInfo(opts *bind.TransactOpts, name, description string) (*types.Transaction, error) {
	data, err := c.abi.Pack("updateProviderInfo", name, description)
	if err != nil {
		return nil, fmt.Errorf("failed to pack updateProviderInfo call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) RemoveProvider(opts *bind.TransactOpts) (*types.Transaction, error) {
	data, err := c.abi.Pack("removeProvider")
	if err != nil {
		return nil, fmt.Errorf("failed to pack removeProvider call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) AddProduct(opts *bind.TransactOpts, productType uint8, capabilityKeys []string, capabilityValues [][]byte) (*types.Transaction, error) {
	data, err := c.abi.Pack("addProduct", productType, capabilityKeys, capabilityValues)
	if err != nil {
		return nil, fmt.Errorf("failed to pack addProduct call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) UpdateProduct(opts *bind.TransactOpts, productType uint8, capabilityKeys []string, capabilityValues [][]byte) (*types.Transaction, error) {
	data, err := c.abi.Pack("updateProduct", productType, capabilityKeys, capabilityValues)
	if err != nil {
		return nil, fmt.Errorf("failed to pack updateProduct call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) RemoveProduct(opts *bind.TransactOpts, productType uint8) (*types.Transaction, error) {
	data, err := c.abi.Pack("removeProduct", productType)
	if err != nil {
		return nil, fmt.Errorf("failed to pack removeProduct call: %w", err)
	}

	return c.transact(opts, data)
}

func (c *Contract) transact(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	nonce, err := c.getNextNonce(opts.Context, opts.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	chainID, err := c.client.ChainID(opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	gasTipCap, err := c.client.SuggestGasTipCap(opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas tip cap: %w", err)
	}

	header, err := c.client.HeaderByNumber(opts.Context, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %w", err)
	}

	baseFee := header.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(0)
	}
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	value := opts.Value
	if value == nil {
		value = big.NewInt(0)
	}

	msg := ethereum.CallMsg{
		From:      opts.From,
		To:        &c.address,
		Value:     value,
		Data:      data,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	}

	gasLimit, err := c.client.EstimateGas(opts.Context, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &c.address,
		Value:     value,
		Data:      data,
	})

	signedTx, err := opts.Signer(opts.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = c.client.SendTransaction(opts.Context, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

func (c *Contract) getNextNonce(ctx context.Context, from common.Address) (uint64, error) {
	c.nonceMu.Lock()
	defer c.nonceMu.Unlock()

	if !c.nonceLoaded {
		pendingNonce, err := c.client.PendingNonceAt(ctx, from)
		if err != nil {
			return 0, err
		}
		c.nonce = pendingNonce
		c.nonceLoaded = true
	}

	nonce := c.nonce
	c.nonce++
	return nonce, nil
}
