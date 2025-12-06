package contracts

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const ERC20ABIJSON = `[
	{
		"type": "function",
		"name": "name",
		"inputs": [],
		"outputs": [{"name": "", "type": "string"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "symbol",
		"inputs": [],
		"outputs": [{"name": "", "type": "string"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "decimals",
		"inputs": [],
		"outputs": [{"name": "", "type": "uint8"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "totalSupply",
		"inputs": [],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "balanceOf",
		"inputs": [{"name": "account", "type": "address"}],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "allowance",
		"inputs": [
			{"name": "owner", "type": "address"},
			{"name": "spender", "type": "address"}
		],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "approve",
		"inputs": [
			{"name": "spender", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "transfer",
		"inputs": [
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "transferFrom",
		"inputs": [
			{"name": "from", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "nonces",
		"inputs": [{"name": "owner", "type": "address"}],
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "DOMAIN_SEPARATOR",
		"inputs": [],
		"outputs": [{"name": "", "type": "bytes32"}],
		"stateMutability": "view"
	}
]`


type ERC20Contract struct {
	address common.Address
	abi     abi.ABI
	client  *ethclient.Client
}


func NewERC20Contract(address common.Address, client *ethclient.Client) (*ERC20Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ERC20ABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	return &ERC20Contract{
		address: address,
		abi:     parsedABI,
		client:  client,
	}, nil
}


func (e *ERC20Contract) Address() common.Address {
	return e.address
}


func (e *ERC20Contract) Name(ctx context.Context) (string, error) {
	data, err := e.abi.Pack("name")
	if err != nil {
		return "", fmt.Errorf("failed to pack name call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("name call failed: %w", err)
	}

	values, err := e.abi.Unpack("name", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack name result: %w", err)
	}

	return values[0].(string), nil
}


func (e *ERC20Contract) Symbol(ctx context.Context) (string, error) {
	data, err := e.abi.Pack("symbol")
	if err != nil {
		return "", fmt.Errorf("failed to pack symbol call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("symbol call failed: %w", err)
	}

	values, err := e.abi.Unpack("symbol", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack symbol result: %w", err)
	}

	return values[0].(string), nil
}


func (e *ERC20Contract) Decimals(ctx context.Context) (uint8, error) {
	data, err := e.abi.Pack("decimals")
	if err != nil {
		return 0, fmt.Errorf("failed to pack decimals call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("decimals call failed: %w", err)
	}

	values, err := e.abi.Unpack("decimals", result)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack decimals result: %w", err)
	}

	return values[0].(uint8), nil
}


func (e *ERC20Contract) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	data, err := e.abi.Pack("balanceOf", account)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("balanceOf call failed: %w", err)
	}

	values, err := e.abi.Unpack("balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack balanceOf result: %w", err)
	}

	return values[0].(*big.Int), nil
}


func (e *ERC20Contract) Allowance(ctx context.Context, owner, spender common.Address) (*big.Int, error) {
	data, err := e.abi.Pack("allowance", owner, spender)
	if err != nil {
		return nil, fmt.Errorf("failed to pack allowance call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("allowance call failed: %w", err)
	}

	values, err := e.abi.Unpack("allowance", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack allowance result: %w", err)
	}

	return values[0].(*big.Int), nil
}


func (e *ERC20Contract) Nonces(ctx context.Context, owner common.Address) (*big.Int, error) {
	data, err := e.abi.Pack("nonces", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to pack nonces call: %w", err)
	}

	result, err := e.client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("nonces call failed: %w", err)
	}

	values, err := e.abi.Unpack("nonces", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack nonces result: %w", err)
	}

	return values[0].(*big.Int), nil
}


func (e *ERC20Contract) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	data, err := e.abi.Pack("approve", spender, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack approve call: %w", err)
	}

	return e.transact(opts, data)
}


func (e *ERC20Contract) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	data, err := e.abi.Pack("transfer", to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack transfer call: %w", err)
	}

	return e.transact(opts, data)
}

func (e *ERC20Contract) transact(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	nonce, err := e.client.PendingNonceAt(opts.Context, opts.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := e.client.SuggestGasPrice(opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	value := opts.Value
	if value == nil {
		value = big.NewInt(0)
	}

	msg := ethereum.CallMsg{
		From:  opts.From,
		To:    &e.address,
		Value: value,
		Data:  data,
	}

	gasLimit, err := e.client.EstimateGas(opts.Context, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	tx := types.NewTransaction(nonce, e.address, value, gasLimit, gasPrice, data)

	signedTx, err := opts.Signer(opts.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = e.client.SendTransaction(opts.Context, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}
