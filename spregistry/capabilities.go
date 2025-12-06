package spregistry

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const (
	CapServiceURL       = "serviceURL"
	CapMinPieceSize     = "minPieceSizeInBytes"
	CapMaxPieceSize     = "maxPieceSizeInBytes"
	CapIPNIPiece        = "ipniPiece"        // Optional
	CapIPNIIPFS         = "ipniIpfs"         // Optional
	CapStoragePrice     = "storagePricePerTibPerDay"
	CapMinProvingPeriod = "minProvingPeriodInEpochs"
	CapLocation         = "location"
	CapPaymentToken     = "paymentTokenAddress"
)

func DecodePDPCapabilities(capabilities map[string][]byte) *PDPOffering {
	offering := &PDPOffering{
		MinPieceSizeInBytes:      big.NewInt(0),
		MaxPieceSizeInBytes:      big.NewInt(0),
		StoragePricePerTiBPerDay: big.NewInt(0),
		MinProvingPeriodInEpochs: big.NewInt(0),
	}

	if v, ok := capabilities[CapServiceURL]; ok {
		offering.ServiceURL = string(v)
	}

	if v, ok := capabilities[CapMinPieceSize]; ok {
		offering.MinPieceSizeInBytes = new(big.Int).SetBytes(v)
	}

	if v, ok := capabilities[CapMaxPieceSize]; ok {
		offering.MaxPieceSizeInBytes = new(big.Int).SetBytes(v)
	}

	_, offering.IPNIPiece = capabilities[CapIPNIPiece]
	_, offering.IPNIIPFS = capabilities[CapIPNIIPFS]

	if v, ok := capabilities[CapStoragePrice]; ok {
		offering.StoragePricePerTiBPerDay = new(big.Int).SetBytes(v)
	}

	if v, ok := capabilities[CapMinProvingPeriod]; ok {
		offering.MinProvingPeriodInEpochs = new(big.Int).SetBytes(v)
	}

	if v, ok := capabilities[CapLocation]; ok {
		offering.Location = string(v)
	}

	if v, ok := capabilities[CapPaymentToken]; ok {
		if len(v) >= 20 {
			offering.PaymentTokenAddress = common.BytesToAddress(v[len(v)-20:])
		}
	}

	return offering
}

func EncodePDPCapabilities(offering *PDPOffering, extraCapabilities map[string]string) ([]string, [][]byte, error) {
	keys := make([]string, 0, 10)
	values := make([][]byte, 0, 10)

	keys = append(keys, CapServiceURL)
	values = append(values, []byte(offering.ServiceURL))

	keys = append(keys, CapMinPieceSize)
	values = append(values, bigIntToBytes(offering.MinPieceSizeInBytes))

	keys = append(keys, CapMaxPieceSize)
	values = append(values, bigIntToBytes(offering.MaxPieceSizeInBytes))

	if offering.IPNIPiece {
		keys = append(keys, CapIPNIPiece)
		values = append(values, []byte{0x01})
	}

	if offering.IPNIIPFS {
		keys = append(keys, CapIPNIIPFS)
		values = append(values, []byte{0x01})
	}

	keys = append(keys, CapStoragePrice)
	values = append(values, bigIntToBytes(offering.StoragePricePerTiBPerDay))

	keys = append(keys, CapMinProvingPeriod)
	values = append(values, bigIntToBytes(offering.MinProvingPeriodInEpochs))

	keys = append(keys, CapLocation)
	values = append(values, []byte(offering.Location))

	keys = append(keys, CapPaymentToken)
	values = append(values, offering.PaymentTokenAddress.Bytes())

	for k, v := range extraCapabilities {
		keys = append(keys, k)
		if v == "" {
			values = append(values, []byte{0x01})
		} else if strings.HasPrefix(v, "0x") {
			decoded, err := hex.DecodeString(v[2:])
			if err != nil {
				return nil, nil, fmt.Errorf("invalid hex value for capability %q: %w", k, err)
			}
			values = append(values, decoded)
		} else {
			values = append(values, []byte(v))
		}
	}

	return keys, values, nil
}

func CapabilitiesListToMap(keys []string, values [][]byte) map[string][]byte {
	result := make(map[string][]byte, len(keys))
	for i := 0; i < len(keys) && i < len(values); i++ {
		result[keys[i]] = values[i]
	}
	return result
}

func bigIntToBytes(n *big.Int) []byte {
	if n == nil {
		return []byte{0}
	}
	b := n.Bytes()
	if len(b) == 0 {
		return []byte{0}
	}
	return b
}
