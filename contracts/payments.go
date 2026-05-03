package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/data-preservation-programs/go-synapse/pkg/abix"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)


const PaymentsABIJSON = `[
	{
		"type": "function",
		"name": "accounts",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "owner", "type": "address"}
		],
		"outputs": [
			{"name": "funds", "type": "uint256"},
			{"name": "lockupCurrent", "type": "uint256"},
			{"name": "lockupRate", "type": "uint256"},
			{"name": "lockupLastSettledAt", "type": "uint256"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getAccountInfoIfSettled",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "owner", "type": "address"}
		],
		"outputs": [
			{"name": "fundedUntilEpoch", "type": "uint256"},
			{"name": "currentFunds", "type": "uint256"},
			{"name": "availableFunds", "type": "uint256"},
			{"name": "currentLockupRate", "type": "uint256"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "deposit",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "payable"
	},
	{
		"type": "function",
		"name": "withdraw",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "withdrawTo",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "setOperatorApproval",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "operator", "type": "address"},
			{"name": "approved", "type": "bool"},
			{"name": "rateAllowance", "type": "uint256"},
			{"name": "lockupAllowance", "type": "uint256"},
			{"name": "maxLockupPeriod", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	},
	{
		"type": "function",
		"name": "operatorApprovals",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "client", "type": "address"},
			{"name": "operator", "type": "address"}
		],
		"outputs": [
			{"name": "isApproved", "type": "bool"},
			{"name": "rateAllowance", "type": "uint256"},
			{"name": "lockupAllowance", "type": "uint256"},
			{"name": "rateUsed", "type": "uint256"},
			{"name": "lockupUsed", "type": "uint256"},
			{"name": "maxLockupPeriod", "type": "uint256"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getRail",
		"inputs": [
			{"name": "railId", "type": "uint256"}
		],
		"outputs": [
			{
				"name": "",
				"type": "tuple",
				"components": [
					{"name": "token", "type": "address"},
					{"name": "from", "type": "address"},
					{"name": "to", "type": "address"},
					{"name": "operator", "type": "address"},
					{"name": "validator", "type": "address"},
					{"name": "paymentRate", "type": "uint256"},
					{"name": "lockupPeriod", "type": "uint256"},
					{"name": "lockupFixed", "type": "uint256"},
					{"name": "settledUpTo", "type": "uint256"},
					{"name": "endEpoch", "type": "uint256"},
					{"name": "commissionRateBps", "type": "uint256"},
					{"name": "serviceFeeRecipient", "type": "address"}
				]
			}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getRailsForPayerAndToken",
		"inputs": [
			{"name": "payer", "type": "address"},
			{"name": "token", "type": "address"},
			{"name": "offset", "type": "uint256"},
			{"name": "limit", "type": "uint256"}
		],
		"outputs": [
			{
				"name": "results",
				"type": "tuple[]",
				"components": [
					{"name": "railId", "type": "uint256"},
					{"name": "isTerminated", "type": "bool"},
					{"name": "endEpoch", "type": "uint256"}
				]
			},
			{"name": "nextOffset", "type": "uint256"},
			{"name": "total", "type": "uint256"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "getRailsForPayeeAndToken",
		"inputs": [
			{"name": "payee", "type": "address"},
			{"name": "token", "type": "address"},
			{"name": "offset", "type": "uint256"},
			{"name": "limit", "type": "uint256"}
		],
		"outputs": [
			{
				"name": "results",
				"type": "tuple[]",
				"components": [
					{"name": "railId", "type": "uint256"},
					{"name": "isTerminated", "type": "bool"},
					{"name": "endEpoch", "type": "uint256"}
				]
			},
			{"name": "nextOffset", "type": "uint256"},
			{"name": "total", "type": "uint256"}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "settleRail",
		"inputs": [
			{"name": "railId", "type": "uint256"},
			{"name": "untilEpoch", "type": "uint256"}
		],
		"outputs": [
			{"name": "totalSettledAmount", "type": "uint256"},
			{"name": "totalNetPayeeAmount", "type": "uint256"},
			{"name": "totalOperatorCommission", "type": "uint256"},
			{"name": "totalNetworkFee", "type": "uint256"},
			{"name": "finalSettledEpoch", "type": "uint256"},
			{"name": "note", "type": "string"}
		],
		"stateMutability": "payable"
	},
	{
		"type": "function",
		"name": "settleTerminatedRailWithoutValidation",
		"inputs": [
			{"name": "railId", "type": "uint256"}
		],
		"outputs": [
			{"name": "totalSettledAmount", "type": "uint256"},
			{"name": "totalNetPayeeAmount", "type": "uint256"},
			{"name": "totalOperatorCommission", "type": "uint256"},
			{"name": "totalNetworkFee", "type": "uint256"},
			{"name": "finalSettledEpoch", "type": "uint256"},
			{"name": "note", "type": "string"}
		],
		"stateMutability": "nonpayable"
	}
]`


type PaymentsContract struct {
	address common.Address
	abi     abi.ABI
	client  *ethclient.Client
}


type RailViewResult struct {
	Token               common.Address
	From                common.Address
	To                  common.Address
	Operator            common.Address
	Validator           common.Address
	PaymentRate         *big.Int
	LockupPeriod        *big.Int
	LockupFixed         *big.Int
	SettledUpTo         *big.Int
	EndEpoch            *big.Int
	CommissionRateBps   *big.Int
	ServiceFeeRecipient common.Address
}


type RailInfoResult struct {
	RailId       *big.Int
	IsTerminated bool
	EndEpoch     *big.Int
}

// getRailOutput mirrors the Rail struct getRail returns. Tagged for json
// round-trip via abix.UnpackSingleTuple; raw type assertion against the
// anonymous struct go-ethereum builds is fragile across versions.
type getRailOutput struct {
	Token               common.Address `json:"token"`
	From                common.Address `json:"from"`
	To                  common.Address `json:"to"`
	Operator            common.Address `json:"operator"`
	Validator           common.Address `json:"validator"`
	PaymentRate         *big.Int       `json:"paymentRate"`
	LockupPeriod        *big.Int       `json:"lockupPeriod"`
	LockupFixed         *big.Int       `json:"lockupFixed"`
	SettledUpTo         *big.Int       `json:"settledUpTo"`
	EndEpoch            *big.Int       `json:"endEpoch"`
	CommissionRateBps   *big.Int       `json:"commissionRateBps"`
	ServiceFeeRecipient common.Address `json:"serviceFeeRecipient"`
}

// getRailsForPayerAndTokenItem mirrors a single tuple element of the
// results array. Same json-tag pattern as getRailOutput.
type getRailsForPayerAndTokenItem struct {
	RailId       *big.Int `json:"railId"`
	IsTerminated bool     `json:"isTerminated"`
	EndEpoch     *big.Int `json:"endEpoch"`
}


func NewPaymentsContract(address common.Address, client *ethclient.Client) (*PaymentsContract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(PaymentsABIJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse payments ABI: %w", err)
	}

	return &PaymentsContract{
		address: address,
		abi:     parsedABI,
		client:  client,
	}, nil
}


func (p *PaymentsContract) Address() common.Address {
	return p.address
}


func (p *PaymentsContract) Accounts(ctx context.Context, token, owner common.Address) (funds, lockupCurrent, lockupRate, lockupLastSettledAt *big.Int, err error) {
	data, err := p.abi.Pack("accounts", token, owner)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to pack accounts call: %w", err)
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("accounts call failed: %w", err)
	}

	values, err := p.abi.Unpack("accounts", result)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unpack accounts result: %w", err)
	}

	return values[0].(*big.Int), values[1].(*big.Int), values[2].(*big.Int), values[3].(*big.Int), nil
}


func (p *PaymentsContract) GetAccountInfoIfSettled(ctx context.Context, token, owner common.Address) (fundedUntilEpoch, currentFunds, availableFunds, currentLockupRate *big.Int, err error) {
	data, err := p.abi.Pack("getAccountInfoIfSettled", token, owner)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to pack getAccountInfoIfSettled call: %w", err)
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("getAccountInfoIfSettled call failed: %w", err)
	}

	values, err := p.abi.Unpack("getAccountInfoIfSettled", result)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unpack getAccountInfoIfSettled result: %w", err)
	}

	return values[0].(*big.Int), values[1].(*big.Int), values[2].(*big.Int), values[3].(*big.Int), nil
}


func (p *PaymentsContract) GetOperatorApproval(ctx context.Context, token, client, operator common.Address) (isApproved bool, rateAllowance, lockupAllowance, rateUsed, lockupUsed, maxLockupPeriod *big.Int, err error) {
	data, err := p.abi.Pack("operatorApprovals", token, client, operator)
	if err != nil {
		return false, nil, nil, nil, nil, nil, fmt.Errorf("failed to pack operatorApprovals call: %w", err)
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.address,
		Data: data,
	}, nil)
	if err != nil {
		return false, nil, nil, nil, nil, nil, fmt.Errorf("operatorApprovals call failed: %w", err)
	}

	values, err := p.abi.Unpack("operatorApprovals", result)
	if err != nil {
		return false, nil, nil, nil, nil, nil, fmt.Errorf("failed to unpack operatorApprovals result: %w", err)
	}

	return values[0].(bool), values[1].(*big.Int), values[2].(*big.Int), values[3].(*big.Int), values[4].(*big.Int), values[5].(*big.Int), nil
}


func (p *PaymentsContract) GetRail(ctx context.Context, railId *big.Int) (*RailViewResult, error) {
	data, err := p.abi.Pack("getRail", railId)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getRail call: %w", err)
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("getRail call failed: %w", err)
	}

	var raw getRailOutput
	if err := abix.UnpackSingleTuple(p.abi, "getRail", result, &raw); err != nil {
		return nil, fmt.Errorf("failed to unpack getRail result: %w", err)
	}

	return &RailViewResult{
		Token:               raw.Token,
		From:                raw.From,
		To:                  raw.To,
		Operator:            raw.Operator,
		Validator:           raw.Validator,
		PaymentRate:         raw.PaymentRate,
		LockupPeriod:        raw.LockupPeriod,
		LockupFixed:         raw.LockupFixed,
		SettledUpTo:         raw.SettledUpTo,
		EndEpoch:            raw.EndEpoch,
		CommissionRateBps:   raw.CommissionRateBps,
		ServiceFeeRecipient: raw.ServiceFeeRecipient,
	}, nil
}


func (p *PaymentsContract) GetRailsForPayerAndToken(ctx context.Context, payer, token common.Address, offset, limit *big.Int) ([]RailInfoResult, *big.Int, *big.Int, error) {
	data, err := p.abi.Pack("getRailsForPayerAndToken", payer, token, offset, limit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to pack getRailsForPayerAndToken call: %w", err)
	}

	result, err := p.client.CallContract(ctx, ethereum.CallMsg{
		To:   &p.address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getRailsForPayerAndToken call failed: %w", err)
	}

	values, err := p.abi.Unpack("getRailsForPayerAndToken", result)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unpack getRailsForPayerAndToken result: %w", err)
	}
	if len(values) != 3 {
		return nil, nil, nil, fmt.Errorf("unexpected getRailsForPayerAndToken result length: %d", len(values))
	}

	// values[0] is a tuple[]: json round-trip the whole slice instead of
	// asserting against the anonymous []struct{...} go-ethereum builds.
	buf, err := json.Marshal(values[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getRailsForPayerAndToken: marshal results: %w", err)
	}
	var rawResults []getRailsForPayerAndTokenItem
	if err := json.Unmarshal(buf, &rawResults); err != nil {
		return nil, nil, nil, fmt.Errorf("getRailsForPayerAndToken: decode results: %w", err)
	}

	results := make([]RailInfoResult, len(rawResults))
	for i, r := range rawResults {
		// getRailsForPayerAndTokenItem and RailInfoResult have identical
		// field names + types; struct tags don't affect Go's conversion
		// identity, so a direct conversion suffices. (gosimple S1016)
		results[i] = RailInfoResult(r)
	}

	nextOffset, ok := values[1].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("unexpected nextOffset type: %T", values[1])
	}
	total, ok := values[2].(*big.Int)
	if !ok {
		return nil, nil, nil, fmt.Errorf("unexpected total type: %T", values[2])
	}
	return results, nextOffset, total, nil
}


func (p *PaymentsContract) Deposit(opts *bind.TransactOpts, token, to common.Address, amount *big.Int) (*types.Transaction, error) {
	data, err := p.abi.Pack("deposit", token, to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack deposit call: %w", err)
	}

	return p.transact(opts, data)
}


func (p *PaymentsContract) Withdraw(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	data, err := p.abi.Pack("withdraw", token, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack withdraw call: %w", err)
	}

	return p.transact(opts, data)
}


func (p *PaymentsContract) SetOperatorApproval(opts *bind.TransactOpts, token, operator common.Address, approved bool, rateAllowance, lockupAllowance, maxLockupPeriod *big.Int) (*types.Transaction, error) {
	data, err := p.abi.Pack("setOperatorApproval", token, operator, approved, rateAllowance, lockupAllowance, maxLockupPeriod)
	if err != nil {
		return nil, fmt.Errorf("failed to pack setOperatorApproval call: %w", err)
	}

	return p.transact(opts, data)
}


func (p *PaymentsContract) SettleRail(opts *bind.TransactOpts, railId, untilEpoch *big.Int) (*types.Transaction, error) {
	data, err := p.abi.Pack("settleRail", railId, untilEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to pack settleRail call: %w", err)
	}

	return p.transact(opts, data)
}

func (p *PaymentsContract) transact(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	nonce, err := p.client.PendingNonceAt(opts.Context, opts.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := p.client.SuggestGasPrice(opts.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	value := opts.Value
	if value == nil {
		value = big.NewInt(0)
	}

	msg := ethereum.CallMsg{
		From:  opts.From,
		To:    &p.address,
		Value: value,
		Data:  data,
	}

	gasLimit, err := p.client.EstimateGas(opts.Context, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	tx := types.NewTransaction(nonce, p.address, value, gasLimit, gasPrice, data)

	signedTx, err := opts.Signer(opts.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = p.client.SendTransaction(opts.Context, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}
