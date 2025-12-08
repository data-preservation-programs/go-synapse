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
	PDPRailID        *big.Int      
	CacheMissRailID  *big.Int      
	CDNRailID        *big.Int      
	Payer            common.Address
	Payee            common.Address
	ServiceProvider  common.Address
	CommissionBps    uint16        
	ClientDataSetID  *big.Int
	PDPEndEpoch      *big.Int
	ProviderID       int
}
