package payments

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)


type Token string

const (
	TokenUSDFC Token = "USDFC"
	TokenFIL   Token = "FIL"
)


type RailInfo struct {
	RailID       *big.Int
	IsTerminated bool
	EndEpoch     *big.Int
}


type RailView struct {
	Token               common.Address
	From                common.Address // Payer
	To                  common.Address // Payee
	Operator            common.Address // Service operator (Warm Storage)
	Validator           common.Address // PDPVerifier contract
	PaymentRate         *big.Int       // Amount per epoch
	LockupPeriod        *big.Int       // Epochs of lockup
	LockupFixed         *big.Int       // Fixed amount locked
	SettledUpTo         *big.Int       // Last settled epoch
	EndEpoch            *big.Int       // 0 if active, > 0 if terminated
	CommissionRateBps   *big.Int       // Commission in basis points
	ServiceFeeRecipient common.Address // Where commission goes
}


type AccountInfo struct {
	Funds              *big.Int
	LockupCurrent      *big.Int
	LockupRate         *big.Int
	LockupLastSettled  *big.Int
	FundedUntilEpoch   *big.Int
	AvailableFunds     *big.Int
	CurrentLockupRate  *big.Int
}


type SettlementResult struct {
	TotalSettledAmount     *big.Int
	TotalNetPayeeAmount    *big.Int
	TotalOperatorCommission *big.Int
	TotalNetworkFee        *big.Int
	FinalSettledEpoch      *big.Int
	Note                   string
}


type OperatorApproval struct {
	IsApproved       bool
	RateAllowance    *big.Int
	LockupAllowance  *big.Int
	RateUsed         *big.Int
	LockupUsed       *big.Int
	MaxLockupPeriod  *big.Int
}


type DepositOptions struct {
	To common.Address
}


type DataSetInfo struct {
	PDPRailID        *big.Int       // PDP payment rail ID
	CacheMissRailID  *big.Int       // CDN cache miss rail (if withCDN)
	CDNRailID        *big.Int       // CDN payment rail (if withCDN)
	Payer            common.Address // Address paying for storage
	Payee            common.Address // SP's beneficiary address
	ServiceProvider  common.Address // SP operator address
	CommissionBps    uint16         // Commission rate in basis points
	ClientDataSetID  *big.Int
	PDPEndEpoch      *big.Int // When PDP payments end (0 if active)
	ProviderID       int
}
