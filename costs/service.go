package costs

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/contracts"
	"github.com/data-preservation-programs/go-synapse/warmstorage"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Service struct {
	ethClient        *ethclient.Client
	chainID          int64
	fwss             *warmstorage.FWSSContract
	pdpVerifier      *contracts.PDPVerifier
	paymentsContract *contracts.PaymentsContract
	usdfcAddress     common.Address
	fwssAddress      common.Address
	pdpVerifierAddr  common.Address
}

type ServiceConfig struct {
	FWSSAddress        common.Address
	PDPVerifierAddress common.Address
	PaymentsAddress    common.Address
	USDFCAddress       common.Address
}

func NewService(client *ethclient.Client, chainID int64, config ServiceConfig) (*Service, error) {
	fwss, err := warmstorage.NewFWSSContract(config.FWSSAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create FWSS contract: %w", err)
	}

	pdpVerifier, err := contracts.NewPDPVerifier(config.PDPVerifierAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDP verifier contract: %w", err)
	}

	paymentsContract, err := contracts.NewPaymentsContract(config.PaymentsAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create payments contract: %w", err)
	}

	return &Service{
		ethClient:        client,
		chainID:          chainID,
		fwss:             fwss,
		pdpVerifier:      pdpVerifier,
		paymentsContract: paymentsContract,
		usdfcAddress:     config.USDFCAddress,
		fwssAddress:      config.FWSSAddress,
		pdpVerifierAddr:  config.PDPVerifierAddress,
	}, nil
}

func (s *Service) GetServicePrice(ctx context.Context) (*warmstorage.ServicePrice, error) {
	return s.fwss.GetServicePrice(ctx)
}

// GetUploadCosts computes costs for uploading uploadSizeBytes to a dataset
// that currently holds dataSetSizeBytes (0 for new).
func (s *Service) GetUploadCosts(
	ctx context.Context,
	payer common.Address,
	dataSetSizeBytes *big.Int,
	uploadSizeBytes *big.Int,
	opts *UploadCostOptions,
) (*UploadCosts, error) {
	if opts == nil {
		opts = &UploadCostOptions{}
	}
	runwayEpochs := opts.RunwayEpochs
	if runwayEpochs == 0 {
		runwayEpochs = DefaultRunwayEpochs
	}
	bufferEpochs := opts.BufferEpochs
	if bufferEpochs == 0 {
		bufferEpochs = DefaultBufferEpochs
	}

	var (
		pricing    *warmstorage.ServicePrice
		acctFunds  *big.Int
		acctLockup *big.Int
		acctRate   *big.Int
		acctSettle *big.Int

		fundedUntil    *big.Int
		currentFunds   *big.Int
		availableFunds *big.Int
		currentRate    *big.Int

		approved       bool
		rateAllowance  *big.Int
		lockAllowance  *big.Int
		rateUsed       *big.Int
		lockUsed       *big.Int
		maxLockPeriod  *big.Int

		usdfcSybilFee *big.Int

		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	wg.Add(4)

	go func() {
		defer wg.Done()
		p, err := s.fwss.GetServicePrice(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("getServicePrice: %w", err))
			return
		}
		pricing = p
	}()

	go func() {
		defer wg.Done()
		f, l, r, st, err := s.paymentsContract.Accounts(ctx, s.usdfcAddress, payer)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("accounts: %w", err))
			mu.Unlock()
			return
		}
		fu, cf, af, cr, err2 := s.paymentsContract.GetAccountInfoIfSettled(ctx, s.usdfcAddress, payer)
		mu.Lock()
		defer mu.Unlock()
		if err2 != nil {
			errs = append(errs, fmt.Errorf("getAccountInfoIfSettled: %w", err2))
			return
		}
		acctFunds, acctLockup, acctRate, acctSettle = f, l, r, st
		fundedUntil, currentFunds, availableFunds, currentRate = fu, cf, af, cr
	}()

	go func() {
		defer wg.Done()
		a, ra, la, ru, lu, ml, err := s.paymentsContract.GetOperatorApproval(
			ctx, s.usdfcAddress, payer, s.fwssAddress,
		)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("getOperatorApproval: %w", err))
			return
		}
		approved, rateAllowance, lockAllowance = a, ra, la
		rateUsed, lockUsed, maxLockPeriod = ru, lu, ml
	}()

	go func() {
		defer wg.Done()
		fee, err := s.readUsdfcSybilFee(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Errorf("usdfcSybilFee: %w", err))
			return
		}
		usdfcSybilFee = fee
	}()

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to fetch contract state: %w", errs[0])
	}

	_ = acctFunds
	_ = acctLockup
	_ = acctRate
	_ = acctSettle
	_ = fundedUntil
	_ = currentFunds
	_ = rateUsed
	_ = lockUsed
	_ = rateAllowance
	_ = lockAllowance

	totalSize := new(big.Int).Add(dataSetSizeBytes, uploadSizeBytes)
	rate := CalculateEffectiveRate(
		totalSize,
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		pricing.EpochsPerMonth.Int64(),
	)

	lockup := CalculateAdditionalLockupRequired(
		uploadSizeBytes,
		dataSetSizeBytes,
		pricing,
		DefaultLockupPeriod,
		usdfcSybilFee,
		opts.IsNewDataSet,
		opts.EnableCDN,
	)

	debt := new(big.Int)
	if availableFunds.Sign() < 0 {
		debt.Neg(availableFunds)
	}

	avail := new(big.Int)
	if availableFunds.Sign() > 0 {
		avail.Set(availableFunds)
	}

	depositNeeded := CalculateDepositNeeded(
		lockup.TotalLockup,
		lockup.RateDelta,
		currentRate,
		debt,
		avail,
		runwayEpochs,
		bufferEpochs,
		opts.IsNewDataSet,
	)

	needsApproval := !approved || maxLockPeriod.Cmp(big.NewInt(DefaultLockupPeriod)) < 0

	ready := depositNeeded.Sign() == 0 && !needsApproval

	return &UploadCosts{
		Rate:                 rate,
		DepositNeeded:        depositNeeded,
		NeedsFWSSMaxApproval: needsApproval,
		Ready:                ready,
	}, nil
}

const usdfcSybilFeeABIJSON = `[{
	"type": "function",
	"name": "USDFC_SYBIL_FEE",
	"inputs": [],
	"outputs": [{"name": "", "type": "uint256"}],
	"stateMutability": "view"
}]`

var usdfcSybilFeeABI abi.ABI

func init() {
	var err error
	usdfcSybilFeeABI, err = abi.JSON(strings.NewReader(usdfcSybilFeeABIJSON))
	if err != nil {
		panic("failed to parse USDFC_SYBIL_FEE ABI: " + err.Error())
	}
}

func (s *Service) readUsdfcSybilFee(ctx context.Context) (*big.Int, error) {
	data, err := usdfcSybilFeeABI.Pack("USDFC_SYBIL_FEE")
	if err != nil {
		return nil, fmt.Errorf("failed to pack USDFC_SYBIL_FEE call: %w", err)
	}

	result, err := s.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &s.pdpVerifierAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call USDFC_SYBIL_FEE: %w", err)
	}

	values, err := usdfcSybilFeeABI.Unpack("USDFC_SYBIL_FEE", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack USDFC_SYBIL_FEE result: %w", err)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("empty result from USDFC_SYBIL_FEE")
	}

	fee, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected type for USDFC_SYBIL_FEE: %T", values[0])
	}

	return fee, nil
}

func NetworkConfig(network constants.Network) (ServiceConfig, error) {
	fwss, ok := constants.WarmStorageAddresses[network]
	if !ok {
		return ServiceConfig{}, fmt.Errorf("no FWSS address for network %s", network)
	}
	pdp, ok := constants.PDPVerifierAddresses[network]
	if !ok {
		return ServiceConfig{}, fmt.Errorf("no PDPVerifier address for network %s", network)
	}
	pay, ok := constants.PaymentsAddresses[network]
	if !ok {
		return ServiceConfig{}, fmt.Errorf("no Payments address for network %s", network)
	}
	usdfc, ok := constants.USDFCAddresses[network]
	if !ok {
		return ServiceConfig{}, fmt.Errorf("no USDFC address for network %s", network)
	}
	return ServiceConfig{
		FWSSAddress:        fwss,
		PDPVerifierAddress: pdp,
		PaymentsAddress:    pay,
		USDFCAddress:       usdfc,
	}, nil
}
