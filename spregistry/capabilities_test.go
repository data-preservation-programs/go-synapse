package spregistry

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDecodePDPCapabilities(t *testing.T) {
	tests := []struct {
		name         string
		capabilities map[string][]byte
		expected     PDPOffering
	}{
		{
			name: "full capabilities",
			capabilities: map[string][]byte{
				CapServiceURL:       []byte("https://provider.example.com"),
				CapMinPieceSize:     big.NewInt(1024).Bytes(),
				CapMaxPieceSize:     big.NewInt(1073741824).Bytes(), // 1 GiB
				CapIPNIPiece:        {0x01},
				CapIPNIIPFS:         {0x01},
				CapStoragePrice:     big.NewInt(1000000).Bytes(),
				CapMinProvingPeriod: big.NewInt(2880).Bytes(),
				CapLocation:         []byte("US-EAST"),
				CapPaymentToken:     common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0").Bytes(),
			},
			expected: PDPOffering{
				ServiceURL:               "https://provider.example.com",
				MinPieceSizeInBytes:      big.NewInt(1024),
				MaxPieceSizeInBytes:      big.NewInt(1073741824),
				IPNIPiece:                true,
				IPNIIPFS:                 true,
				StoragePricePerTiBPerDay: big.NewInt(1000000),
				MinProvingPeriodInEpochs: big.NewInt(2880),
				Location:                 "US-EAST",
				PaymentTokenAddress:      common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
			},
		},
		{
			name: "minimal capabilities without optional flags",
			capabilities: map[string][]byte{
				CapServiceURL:       []byte("https://provider.example.com"),
				CapMinPieceSize:     big.NewInt(127).Bytes(),
				CapMaxPieceSize:     big.NewInt(1073741824).Bytes(),
				CapStoragePrice:     big.NewInt(500000).Bytes(),
				CapMinProvingPeriod: big.NewInt(1440).Bytes(),
				CapLocation:         []byte("EU"),
				CapPaymentToken:     common.Address{}.Bytes(),
			},
			expected: PDPOffering{
				ServiceURL:               "https://provider.example.com",
				MinPieceSizeInBytes:      big.NewInt(127),
				MaxPieceSizeInBytes:      big.NewInt(1073741824),
				IPNIPiece:                false,
				IPNIIPFS:                 false,
				StoragePricePerTiBPerDay: big.NewInt(500000),
				MinProvingPeriodInEpochs: big.NewInt(1440),
				Location:                 "EU",
				PaymentTokenAddress:      common.Address{},
			},
		},
		{
			name:         "empty capabilities",
			capabilities: map[string][]byte{},
			expected: PDPOffering{
				ServiceURL:               "",
				MinPieceSizeInBytes:      big.NewInt(0),
				MaxPieceSizeInBytes:      big.NewInt(0),
				IPNIPiece:                false,
				IPNIIPFS:                 false,
				StoragePricePerTiBPerDay: big.NewInt(0),
				MinProvingPeriodInEpochs: big.NewInt(0),
				Location:                 "",
				PaymentTokenAddress:      common.Address{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodePDPCapabilities(tt.capabilities)

			if result.ServiceURL != tt.expected.ServiceURL {
				t.Errorf("ServiceURL = %s, want %s", result.ServiceURL, tt.expected.ServiceURL)
			}
			if result.MinPieceSizeInBytes.Cmp(tt.expected.MinPieceSizeInBytes) != 0 {
				t.Errorf("MinPieceSizeInBytes = %s, want %s", result.MinPieceSizeInBytes, tt.expected.MinPieceSizeInBytes)
			}
			if result.MaxPieceSizeInBytes.Cmp(tt.expected.MaxPieceSizeInBytes) != 0 {
				t.Errorf("MaxPieceSizeInBytes = %s, want %s", result.MaxPieceSizeInBytes, tt.expected.MaxPieceSizeInBytes)
			}
			if result.IPNIPiece != tt.expected.IPNIPiece {
				t.Errorf("IPNIPiece = %v, want %v", result.IPNIPiece, tt.expected.IPNIPiece)
			}
			if result.IPNIIPFS != tt.expected.IPNIIPFS {
				t.Errorf("IPNIIPFS = %v, want %v", result.IPNIIPFS, tt.expected.IPNIIPFS)
			}
			if result.StoragePricePerTiBPerDay.Cmp(tt.expected.StoragePricePerTiBPerDay) != 0 {
				t.Errorf("StoragePricePerTiBPerDay = %s, want %s", result.StoragePricePerTiBPerDay, tt.expected.StoragePricePerTiBPerDay)
			}
			if result.MinProvingPeriodInEpochs.Cmp(tt.expected.MinProvingPeriodInEpochs) != 0 {
				t.Errorf("MinProvingPeriodInEpochs = %s, want %s", result.MinProvingPeriodInEpochs, tt.expected.MinProvingPeriodInEpochs)
			}
			if result.Location != tt.expected.Location {
				t.Errorf("Location = %s, want %s", result.Location, tt.expected.Location)
			}
			if result.PaymentTokenAddress != tt.expected.PaymentTokenAddress {
				t.Errorf("PaymentTokenAddress = %s, want %s", result.PaymentTokenAddress, tt.expected.PaymentTokenAddress)
			}
		})
	}
}

func TestEncodePDPCapabilities(t *testing.T) {
	offering := PDPOffering{
		ServiceURL:               "https://provider.example.com",
		MinPieceSizeInBytes:      big.NewInt(1024),
		MaxPieceSizeInBytes:      big.NewInt(1073741824),
		IPNIPiece:                true,
		IPNIIPFS:                 false,
		StoragePricePerTiBPerDay: big.NewInt(1000000),
		MinProvingPeriodInEpochs: big.NewInt(2880),
		Location:                 "US-EAST",
		PaymentTokenAddress:      common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
	}

	keys, values, err := EncodePDPCapabilities(&offering, nil)
	if err != nil {
		t.Fatalf("EncodePDPCapabilities failed: %v", err)
	}

	expectedKeys := []string{
		CapServiceURL,
		CapMinPieceSize,
		CapMaxPieceSize,
		CapIPNIPiece, // IPNIPiece is true
		CapStoragePrice,
		CapMinProvingPeriod,
		CapLocation,
		CapPaymentToken,
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("len(keys) = %d, want %d", len(keys), len(expectedKeys))
	}

	for i, key := range expectedKeys {
		if keys[i] != key {
			t.Errorf("keys[%d] = %s, want %s", i, keys[i], key)
		}
	}

	capMap := CapabilitiesListToMap(keys, values)
	decoded := DecodePDPCapabilities(capMap)

	if decoded.ServiceURL != offering.ServiceURL {
		t.Errorf("decoded.ServiceURL = %s, want %s", decoded.ServiceURL, offering.ServiceURL)
	}
	if decoded.MinPieceSizeInBytes.Cmp(offering.MinPieceSizeInBytes) != 0 {
		t.Errorf("decoded.MinPieceSizeInBytes = %s, want %s", decoded.MinPieceSizeInBytes, offering.MinPieceSizeInBytes)
	}
	if decoded.MaxPieceSizeInBytes.Cmp(offering.MaxPieceSizeInBytes) != 0 {
		t.Errorf("decoded.MaxPieceSizeInBytes = %s, want %s", decoded.MaxPieceSizeInBytes, offering.MaxPieceSizeInBytes)
	}
	if decoded.IPNIPiece != offering.IPNIPiece {
		t.Errorf("decoded.IPNIPiece = %v, want %v", decoded.IPNIPiece, offering.IPNIPiece)
	}
	if decoded.IPNIIPFS != offering.IPNIIPFS {
		t.Errorf("decoded.IPNIIPFS = %v, want %v", decoded.IPNIIPFS, offering.IPNIIPFS)
	}
}

func TestEncodePDPCapabilities_WithExtraCapabilities(t *testing.T) {
	offering := PDPOffering{
		ServiceURL:               "https://provider.example.com",
		MinPieceSizeInBytes:      big.NewInt(1024),
		MaxPieceSizeInBytes:      big.NewInt(1073741824),
		IPNIPiece:                false,
		IPNIIPFS:                 false,
		StoragePricePerTiBPerDay: big.NewInt(1000000),
		MinProvingPeriodInEpochs: big.NewInt(2880),
		Location:                 "US-EAST",
		PaymentTokenAddress:      common.Address{},
	}

	extras := map[string]string{
		"customKey":  "customValue",
		"emptyValue": "",
		"hexValue":   "0xabcd",
	}

	keys, values, err := EncodePDPCapabilities(&offering, extras)
	if err != nil {
		t.Fatalf("EncodePDPCapabilities failed: %v", err)
	}

	expectedBaseCount := 7 // no IPNIPiece or IPNIIPFS since both are false
	expectedTotal := expectedBaseCount + len(extras)

	if len(keys) != expectedTotal {
		t.Errorf("len(keys) = %d, want %d", len(keys), expectedTotal)
	}
	if len(values) != expectedTotal {
		t.Errorf("len(values) = %d, want %d", len(values), expectedTotal)
	}

	capMap := CapabilitiesListToMap(keys, values)

	if _, ok := capMap["customKey"]; !ok {
		t.Error("Expected customKey in capabilities")
	}
	if _, ok := capMap["emptyValue"]; !ok {
		t.Error("Expected emptyValue in capabilities")
	}
	if _, ok := capMap["hexValue"]; !ok {
		t.Error("Expected hexValue in capabilities")
	}
}

func TestEncodePDPCapabilities_InvalidHex(t *testing.T) {
	offering := PDPOffering{
		ServiceURL:               "https://provider.example.com",
		MinPieceSizeInBytes:      big.NewInt(1024),
		MaxPieceSizeInBytes:      big.NewInt(1073741824),
		StoragePricePerTiBPerDay: big.NewInt(1000000),
		MinProvingPeriodInEpochs: big.NewInt(2880),
		Location:                 "US-EAST",
	}

	extras := map[string]string{
		"badHex": "0xZZZZ", // Invalid hex
	}

	_, _, err := EncodePDPCapabilities(&offering, extras)
	if err == nil {
		t.Error("Expected error for invalid hex value, got nil")
	}
}

func TestCapabilitiesListToMap(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}
	values := [][]byte{[]byte("value1"), []byte("value2"), []byte("value3")}

	result := CapabilitiesListToMap(keys, values)

	if len(result) != 3 {
		t.Errorf("len(result) = %d, want 3", len(result))
	}

	for i, key := range keys {
		if string(result[key]) != string(values[i]) {
			t.Errorf("result[%s] = %s, want %s", key, result[key], values[i])
		}
	}
}

func TestCapabilitiesListToMap_MismatchedLengths(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}
	values := [][]byte{[]byte("value1"), []byte("value2")} // only 2 values

	result := CapabilitiesListToMap(keys, values)

	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}
