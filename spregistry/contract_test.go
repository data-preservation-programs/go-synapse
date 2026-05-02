package spregistry

import (
	"math/big"
	"strings"
	"testing"

	"github.com/data-preservation-programs/go-synapse/pkg/abix"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TestUnpackSingleTuple_GetProviderByAddress exercises the unpack path
// Contract.GetProviderByAddress uses, against a synthetic return blob.
// Reproduces the calibnet bug if abix.UnpackSingleTuple regresses to
// UnpackIntoInterface (which mishandles this shape).
func TestUnpackSingleTuple_GetProviderByAddress(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(SPRegistryABIJSON))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["getProviderByAddress"]
	if !ok {
		t.Fatalf("getProviderByAddress not found in ABI")
	}

	type infoT struct {
		ServiceProvider common.Address `abi:"serviceProvider"`
		Payee           common.Address `abi:"payee"`
		Name            string         `abi:"name"`
		Description     string         `abi:"description"`
		IsActive        bool           `abi:"isActive"`
	}
	type outT struct {
		ProviderID *big.Int `abi:"providerId"`
		Info       infoT    `abi:"info"`
	}
	want := outT{
		ProviderID: big.NewInt(24),
		Info: infoT{
			ServiceProvider: common.HexToAddress("0xE3e842B9D89ed2Ee3976b9b8916827302618c29e"),
			Payee:           common.HexToAddress("0xE3e842B9D89ed2Ee3976b9b8916827302618c29e"),
			Name:            "sp-playground",
			Description:     "calibnet test SP",
			IsActive:        true,
		},
	}

	payload, err := method.Outputs.Pack(want)
	if err != nil {
		t.Fatalf("pack synthetic return: %v", err)
	}

	var got getProviderByAddressOutput
	if err := abix.UnpackSingleTuple(parsedABI, "getProviderByAddress", payload, &got); err != nil {
		t.Fatalf("UnpackSingleTuple: %v", err)
	}

	if got.ProviderID == nil || got.ProviderID.Cmp(big.NewInt(24)) != 0 {
		t.Errorf("ProviderID = %v, want 24", got.ProviderID)
	}
	if got.Info.ServiceProvider != want.Info.ServiceProvider {
		t.Errorf("ServiceProvider = %s, want %s", got.Info.ServiceProvider, want.Info.ServiceProvider)
	}
	if got.Info.Payee != want.Info.Payee {
		t.Errorf("Payee = %s, want %s", got.Info.Payee, want.Info.Payee)
	}
	if got.Info.Name != want.Info.Name {
		t.Errorf("Name = %q, want %q", got.Info.Name, want.Info.Name)
	}
	if got.Info.Description != want.Info.Description {
		t.Errorf("Description = %q, want %q", got.Info.Description, want.Info.Description)
	}
	if !got.Info.IsActive {
		t.Errorf("IsActive = false, want true")
	}
}

// TestUnpackSingleTuple_GetProviderWithProduct exercises the same unpack
// path for getProviderWithProduct, which has a deeper return tuple
// (providerInfo + product + productCapabilityValues) and was missed by
// d91ba0c.
func TestUnpackSingleTuple_GetProviderWithProduct(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(SPRegistryABIJSON))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}

	method, ok := parsedABI.Methods["getProviderWithProduct"]
	if !ok {
		t.Fatalf("getProviderWithProduct not found in ABI")
	}

	type providerInfoT struct {
		ServiceProvider common.Address `abi:"serviceProvider"`
		Payee           common.Address `abi:"payee"`
		Name            string         `abi:"name"`
		Description     string         `abi:"description"`
		IsActive        bool           `abi:"isActive"`
	}
	type productT struct {
		ProductType    uint8    `abi:"productType"`
		CapabilityKeys []string `abi:"capabilityKeys"`
		IsActive       bool     `abi:"isActive"`
	}
	type outT struct {
		ProviderID              *big.Int      `abi:"providerId"`
		ProviderInfo            providerInfoT `abi:"providerInfo"`
		Product                 productT      `abi:"product"`
		ProductCapabilityValues [][]byte      `abi:"productCapabilityValues"`
	}
	want := outT{
		ProviderID: big.NewInt(24),
		ProviderInfo: providerInfoT{
			ServiceProvider: common.HexToAddress("0xE3e842B9D89ed2Ee3976b9b8916827302618c29e"),
			Payee:           common.HexToAddress("0xE3e842B9D89ed2Ee3976b9b8916827302618c29e"),
			Name:            "sp-playground",
			Description:     "calibnet test SP",
			IsActive:        true,
		},
		Product: productT{
			ProductType:    0,
			CapabilityKeys: []string{"serviceURL", "minPieceSizeInBytes"},
			IsActive:       true,
		},
		ProductCapabilityValues: [][]byte{
			[]byte("https://pdp.sp-playground.xyz"),
			{0x10, 0x00, 0x00},
		},
	}

	payload, err := method.Outputs.Pack(want)
	if err != nil {
		t.Fatalf("pack synthetic return: %v", err)
	}

	var got getProviderWithProductOutput
	if err := abix.UnpackSingleTuple(parsedABI, "getProviderWithProduct", payload, &got); err != nil {
		t.Fatalf("UnpackSingleTuple: %v", err)
	}

	if got.ProviderID == nil || got.ProviderID.Cmp(big.NewInt(24)) != 0 {
		t.Errorf("ProviderID = %v, want 24", got.ProviderID)
	}
	if got.ProviderInfo.Name != want.ProviderInfo.Name {
		t.Errorf("ProviderInfo.Name = %q, want %q", got.ProviderInfo.Name, want.ProviderInfo.Name)
	}
	if got.Product.ProductType != want.Product.ProductType {
		t.Errorf("Product.ProductType = %d, want %d", got.Product.ProductType, want.Product.ProductType)
	}
	if len(got.Product.CapabilityKeys) != len(want.Product.CapabilityKeys) {
		t.Errorf("CapabilityKeys len = %d, want %d", len(got.Product.CapabilityKeys), len(want.Product.CapabilityKeys))
	}
	if len(got.ProductCapabilityValues) != len(want.ProductCapabilityValues) {
		t.Errorf("ProductCapabilityValues len = %d, want %d", len(got.ProductCapabilityValues), len(want.ProductCapabilityValues))
	}
	if string(got.ProductCapabilityValues[0]) != string(want.ProductCapabilityValues[0]) {
		t.Errorf("ProductCapabilityValues[0] = %q, want %q", string(got.ProductCapabilityValues[0]), string(want.ProductCapabilityValues[0]))
	}
}
