package spregistry

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type ProductType int

const (
	ProductTypePDP ProductType = 0
)

type PDPOffering struct {
	ServiceURL              string
	MinPieceSizeInBytes     *big.Int
	MaxPieceSizeInBytes     *big.Int
	IPNIPiece               bool
	IPNIIPFS                bool
	StoragePricePerTiBPerDay *big.Int
	MinProvingPeriodInEpochs *big.Int
	Location                string
	PaymentTokenAddress     common.Address
}

type ServiceProduct struct {
	Type         string // "PDP"
	IsActive     bool
	Capabilities map[string][]byte
	Data         *PDPOffering
}

type ProviderInfo struct {
	ID              int
	ServiceProvider common.Address
	Payee           common.Address
	Name            string
	Description     string
	Active          bool
	Products        map[string]*ServiceProduct // keyed by product type ("PDP")
}

type ProviderRegistrationInfo struct {
	Payee        common.Address
	Name         string
	Description  string
	PDPOffering  PDPOffering
	Capabilities map[string]string // additional key-value pairs
}

type PDPServiceInfo struct {
	Offering     PDPOffering
	Capabilities map[string][]byte
	IsActive     bool
}

type RawProviderInfo struct {
	ServiceProvider common.Address
	Payee           common.Address
	Name            string
	Description     string
	IsActive        bool
}

type RawProduct struct {
	ProductType    uint8
	CapabilityKeys []string
	IsActive       bool
}
