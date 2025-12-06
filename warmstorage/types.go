package warmstorage

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DataSetInfo struct {
	PDPRailID       *big.Int
	CacheMissRailID *big.Int
	CDNRailID       *big.Int
	Payer           common.Address
	Payee           common.Address
	ServiceProvider common.Address
	CommissionBps   *big.Int
	ClientDataSetID *big.Int
	PDPEndEpoch     *big.Int
	ProviderID      *big.Int
	DataSetID       *big.Int
}
