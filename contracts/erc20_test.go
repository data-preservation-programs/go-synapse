package contracts

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func TestERC20ABI(t *testing.T) {
	t.Run("should parse ABI successfully", func(t *testing.T) {
		parsedABI, err := abi.JSON(strings.NewReader(ERC20ABIJSON))
		if err != nil {
			t.Fatalf("Failed to parse ABI: %v", err)
		}

		methods := []string{
			"name",
			"symbol",
			"decimals",
			"totalSupply",
			"balanceOf",
			"allowance",
			"approve",
			"transfer",
			"transferFrom",
			"nonces",
			"DOMAIN_SEPARATOR",
		}

		for _, method := range methods {
			if _, ok := parsedABI.Methods[method]; !ok {
				t.Errorf("Missing method: %s", method)
			}
		}
	})

	t.Run("should pack balanceOf correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		account := common.HexToAddress("0x1234567890123456789012345678901234567890")

		data, err := parsedABI.Pack("balanceOf", account)
		if err != nil {
			t.Fatalf("Failed to pack balanceOf: %v", err)
		}

		expectedLen := 4 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack allowance correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		owner := common.HexToAddress("0x1234567890123456789012345678901234567890")
		spender := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

		data, err := parsedABI.Pack("allowance", owner, spender)
		if err != nil {
			t.Fatalf("Failed to pack allowance: %v", err)
		}

		expectedLen := 4 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack approve correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		spender := common.HexToAddress("0x1234567890123456789012345678901234567890")
		amount := big.NewInt(1000000000000000000)

		data, err := parsedABI.Pack("approve", spender, amount)
		if err != nil {
			t.Fatalf("Failed to pack approve: %v", err)
		}

		expectedLen := 4 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack transfer correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		to := common.HexToAddress("0x1234567890123456789012345678901234567890")
		amount := big.NewInt(1000000000000000000)

		data, err := parsedABI.Pack("transfer", to, amount)
		if err != nil {
			t.Fatalf("Failed to pack transfer: %v", err)
		}

		expectedLen := 4 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack transferFrom correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		from := common.HexToAddress("0x1111111111111111111111111111111111111111")
		to := common.HexToAddress("0x2222222222222222222222222222222222222222")
		amount := big.NewInt(1000000000000000000)

		data, err := parsedABI.Pack("transferFrom", from, to, amount)
		if err != nil {
			t.Fatalf("Failed to pack transferFrom: %v", err)
		}

		expectedLen := 4 + 32 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack nonces correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

		owner := common.HexToAddress("0x1234567890123456789012345678901234567890")

		data, err := parsedABI.Pack("nonces", owner)
		if err != nil {
			t.Fatalf("Failed to pack nonces: %v", err)
		}

		expectedLen := 4 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})
}

func TestERC20MethodSelectors(t *testing.T) {
	parsedABI, _ := abi.JSON(strings.NewReader(ERC20ABIJSON))

	expectedSelectors := map[string]string{
		"name":         "06fdde03",
		"symbol":       "95d89b41",
		"decimals":     "313ce567",
		"totalSupply":  "18160ddd",
		"balanceOf":    "70a08231",
		"allowance":    "dd62ed3e",
		"approve":      "095ea7b3",
		"transfer":     "a9059cbb",
		"transferFrom": "23b872dd",
	}

	for method, expectedSelector := range expectedSelectors {
		t.Run(method, func(t *testing.T) {
			m, ok := parsedABI.Methods[method]
			if !ok {
				t.Fatalf("Method %s not found", method)
			}

			actualSelector := common.Bytes2Hex(m.ID)
			if actualSelector != expectedSelector {
				t.Errorf("Method %s: expected selector %s, got %s", method, expectedSelector, actualSelector)
			}
		})
	}
}
