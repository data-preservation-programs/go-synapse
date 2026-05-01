package spregistry

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TestUnpackSingleTuple_GetProviderByAddress exercises the unpack path
// Contract.GetProviderByAddress uses, against a synthetic return blob.
// Reproduces the calibnet bug if unpackSingleTuple regresses to
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
	if err := unpackSingleTuple(parsedABI, "getProviderByAddress", payload, &got); err != nil {
		t.Fatalf("unpackSingleTuple: %v", err)
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
