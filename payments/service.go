package payments

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/data-preservation-programs/go-synapse/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)


type Service struct {
	client           *ethclient.Client
	privateKey       *ecdsa.PrivateKey
	address          common.Address
	chainID          *big.Int
	paymentsContract *contracts.PaymentsContract
	paymentsAddress  common.Address
	usdfcContract    *contracts.ERC20Contract
	usdfcAddress     common.Address
}


func NewService(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	chainID *big.Int,
	paymentsAddress common.Address,
) (*Service, error) {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	usdfcAddress, ok := USDFCAddresses[chainID.Int64()]
	if !ok {
		return nil, fmt.Errorf("USDFC address not found for chain ID %d", chainID.Int64())
	}

	paymentsContract, err := contracts.NewPaymentsContract(paymentsAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create payments contract: %w", err)
	}

	usdfcContract, err := contracts.NewERC20Contract(usdfcAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create USDFC contract: %w", err)
	}

	return &Service{
		client:           client,
		privateKey:       privateKey,
		address:          address,
		chainID:          chainID,
		paymentsContract: paymentsContract,
		paymentsAddress:  paymentsAddress,
		usdfcContract:    usdfcContract,
		usdfcAddress:     usdfcAddress,
	}, nil
}


func (s *Service) Address() common.Address {
	return s.address
}


func (s *Service) PaymentsAddress() common.Address {
	return s.paymentsAddress
}


func (s *Service) USDFCAddress() common.Address {
	return s.usdfcAddress
}


func (s *Service) Balance(ctx context.Context, token Token) (*big.Int, error) {
	tokenAddr := s.tokenAddress(token)
	funds, _, _, _, err := s.paymentsContract.Accounts(ctx, tokenAddr, s.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}
	return funds, nil
}


func (s *Service) WalletBalance(ctx context.Context, token Token) (*big.Int, error) {
	if token == TokenFIL {
		return s.client.BalanceAt(ctx, s.address, nil)
	}

	tokenAddr := s.tokenAddress(token)
	tokenContract, err := contracts.NewERC20Contract(tokenAddr, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create token contract: %w", err)
	}

	return tokenContract.BalanceOf(ctx, s.address)
}


func (s *Service) AccountInfo(ctx context.Context, token Token) (*AccountInfo, error) {
	tokenAddr := s.tokenAddress(token)

	funds, lockupCurrent, lockupRate, lockupLastSettled, err := s.paymentsContract.Accounts(ctx, tokenAddr, s.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	fundedUntilEpoch, _, availableFunds, currentLockupRate, err := s.paymentsContract.GetAccountInfoIfSettled(ctx, tokenAddr, s.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get settled account info: %w", err)
	}

	return &AccountInfo{
		Funds:             funds,
		LockupCurrent:     lockupCurrent,
		LockupRate:        lockupRate,
		LockupLastSettled: lockupLastSettled,
		FundedUntilEpoch:  fundedUntilEpoch,
		AvailableFunds:    availableFunds,
		CurrentLockupRate: currentLockupRate,
	}, nil
}


func (s *Service) Allowance(ctx context.Context, token Token) (*big.Int, error) {
	tokenAddr := s.tokenAddress(token)
	tokenContract, err := contracts.NewERC20Contract(tokenAddr, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create token contract: %w", err)
	}

	return tokenContract.Allowance(ctx, s.address, s.paymentsAddress)
}


func (s *Service) Approve(ctx context.Context, amount *big.Int, token Token) (common.Hash, error) {
	tokenAddr := s.tokenAddress(token)
	tokenContract, err := contracts.NewERC20Contract(tokenAddr, s.client)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to create token contract: %w", err)
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := tokenContract.Approve(opts, s.paymentsAddress, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) Deposit(ctx context.Context, amount *big.Int, token Token, opts *DepositOptions) (common.Hash, error) {
	tokenAddr := s.tokenAddress(token)

	allowance, err := s.Allowance(ctx, token)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to check allowance: %w", err)
	}

	if allowance.Cmp(amount) < 0 {
		_, err := s.Approve(ctx, amount, token)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to approve: %w", err)
		}
	}

	to := s.address
	if opts != nil && opts.To != (common.Address{}) {
		to = opts.To
	}

	txOpts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.paymentsContract.Deposit(txOpts, tokenAddr, to, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to deposit: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) Withdraw(ctx context.Context, amount *big.Int, token Token) (common.Hash, error) {
	tokenAddr := s.tokenAddress(token)

	info, err := s.AccountInfo(ctx, token)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get account info: %w", err)
	}

	if info.AvailableFunds.Cmp(amount) < 0 {
		return common.Hash{}, fmt.Errorf("insufficient available funds: have %s, want %s", info.AvailableFunds.String(), amount.String())
	}

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.paymentsContract.Withdraw(opts, tokenAddr, amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to withdraw: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) ApproveService(ctx context.Context, operator common.Address, rateAllowance, lockupAllowance, maxLockupPeriod *big.Int, token Token) (common.Hash, error) {
	tokenAddr := s.tokenAddress(token)

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.paymentsContract.SetOperatorApproval(opts, tokenAddr, operator, true, rateAllowance, lockupAllowance, maxLockupPeriod)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to approve service: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) RevokeService(ctx context.Context, operator common.Address, token Token) (common.Hash, error) {
	tokenAddr := s.tokenAddress(token)

	opts, err := s.transactOpts(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	tx, err := s.paymentsContract.SetOperatorApproval(opts, tokenAddr, operator, false, big.NewInt(0), big.NewInt(0), big.NewInt(0))
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to revoke service: %w", err)
	}

	return tx.Hash(), nil
}


func (s *Service) ServiceApproval(ctx context.Context, operator common.Address, token Token) (*OperatorApproval, error) {
	tokenAddr := s.tokenAddress(token)

	isApproved, rateAllowance, lockupAllowance, rateUsed, lockupUsed, maxLockupPeriod, err := s.paymentsContract.GetOperatorApproval(ctx, tokenAddr, s.address, operator)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator approval: %w", err)
	}

	return &OperatorApproval{
		IsApproved:      isApproved,
		RateAllowance:   rateAllowance,
		LockupAllowance: lockupAllowance,
		RateUsed:        rateUsed,
		LockupUsed:      lockupUsed,
		MaxLockupPeriod: maxLockupPeriod,
	}, nil
}


func (s *Service) GetRail(ctx context.Context, railID *big.Int) (*RailView, error) {
	rail, err := s.paymentsContract.GetRail(ctx, railID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rail: %w", err)
	}

	return &RailView{
		Token:               rail.Token,
		From:                rail.From,
		To:                  rail.To,
		Operator:            rail.Operator,
		Validator:           rail.Validator,
		PaymentRate:         rail.PaymentRate,
		LockupPeriod:        rail.LockupPeriod,
		LockupFixed:         rail.LockupFixed,
		SettledUpTo:         rail.SettledUpTo,
		EndEpoch:            rail.EndEpoch,
		CommissionRateBps:   rail.CommissionRateBps,
		ServiceFeeRecipient: rail.ServiceFeeRecipient,
	}, nil
}


func (s *Service) GetRailsAsPayer(ctx context.Context, token Token) ([]RailInfo, error) {
	tokenAddr := s.tokenAddress(token)

	var allRails []RailInfo
	offset := big.NewInt(0)
	limit := big.NewInt(100)

	for {
		results, nextOffset, _, err := s.paymentsContract.GetRailsForPayerAndToken(ctx, s.address, tokenAddr, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get rails: %w", err)
		}

		for _, r := range results {
			allRails = append(allRails, RailInfo{
				RailID:       r.RailId,
				IsTerminated: r.IsTerminated,
				EndEpoch:     r.EndEpoch,
			})
		}

		if nextOffset.Cmp(big.NewInt(0)) == 0 || len(results) < int(limit.Int64()) {
			break
		}
		offset = nextOffset
	}

	return allRails, nil
}


func (s *Service) Settle(ctx context.Context, railID, untilEpoch *big.Int) (*SettlementResult, error) {
	opts, err := s.transactOpts(ctx)
	if err != nil {
		return nil, err
	}

	opts.Value = SettlementFee

	tx, err := s.paymentsContract.SettleRail(opts, railID, untilEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to settle rail: %w", err)
	}

	return &SettlementResult{
		Note: fmt.Sprintf("Settlement transaction submitted: %s", tx.Hash().Hex()),
	}, nil
}

func (s *Service) tokenAddress(token Token) common.Address {
	switch token {
	case TokenUSDFC:
		return s.usdfcAddress
	case TokenFIL:
		return common.Address{}
	default:
		return common.HexToAddress(string(token))
	}
}

func (s *Service) transactOpts(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	opts.Context = ctx
	return opts, nil
}
