package contracts

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func TestPaymentsABI(t *testing.T) {
	t.Run("should parse ABI successfully", func(t *testing.T) {
		parsedABI, err := abi.JSON(strings.NewReader(PaymentsABIJSON))
		if err != nil {
			t.Fatalf("Failed to parse ABI: %v", err)
		}

		methods := []string{
			"accounts",
			"getAccountInfoIfSettled",
			"deposit",
			"withdraw",
			"withdrawTo",
			"setOperatorApproval",
			"operatorApprovals",
			"getRail",
			"getRailsForPayerAndToken",
			"getRailsForPayeeAndToken",
			"settleRail",
			"settleTerminatedRailWithoutValidation",
		}

		for _, method := range methods {
			if _, ok := parsedABI.Methods[method]; !ok {
				t.Errorf("Missing method: %s", method)
			}
		}
	})

	t.Run("should pack accounts call correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		token := common.HexToAddress("0x1234567890123456789012345678901234567890")
		owner := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

		data, err := parsedABI.Pack("accounts", token, owner)
		if err != nil {
			t.Fatalf("Failed to pack accounts: %v", err)
		}

		expectedLen := 4 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack deposit call correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		token := common.HexToAddress("0x1234567890123456789012345678901234567890")
		to := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
		amount := big.NewInt(1000000000000000000) // 1 token

		data, err := parsedABI.Pack("deposit", token, to, amount)
		if err != nil {
			t.Fatalf("Failed to pack deposit: %v", err)
		}

		expectedLen := 4 + 32 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack setOperatorApproval correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		token := common.HexToAddress("0x1234567890123456789012345678901234567890")
		operator := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
		approved := true
		rateAllowance := big.NewInt(1000000000000000000)
		lockupAllowance := big.NewInt(5000000000000000000)
		maxLockupPeriod := big.NewInt(86400)

		data, err := parsedABI.Pack("setOperatorApproval", token, operator, approved, rateAllowance, lockupAllowance, maxLockupPeriod)
		if err != nil {
			t.Fatalf("Failed to pack setOperatorApproval: %v", err)
		}

		expectedLen := 4 + 32 + 32 + 32 + 32 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack getRail correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		railId := big.NewInt(123)

		data, err := parsedABI.Pack("getRail", railId)
		if err != nil {
			t.Fatalf("Failed to pack getRail: %v", err)
		}

		expectedLen := 4 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack settleRail correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		railId := big.NewInt(123)
		untilEpoch := big.NewInt(1000000)

		data, err := parsedABI.Pack("settleRail", railId, untilEpoch)
		if err != nil {
			t.Fatalf("Failed to pack settleRail: %v", err)
		}

		expectedLen := 4 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})

	t.Run("should pack getRailsForPayerAndToken correctly", func(t *testing.T) {
		parsedABI, _ := abi.JSON(strings.NewReader(PaymentsABIJSON))

		payer := common.HexToAddress("0x1234567890123456789012345678901234567890")
		token := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
		offset := big.NewInt(0)
		limit := big.NewInt(100)

		data, err := parsedABI.Pack("getRailsForPayerAndToken", payer, token, offset, limit)
		if err != nil {
			t.Fatalf("Failed to pack getRailsForPayerAndToken: %v", err)
		}

		expectedLen := 4 + 32 + 32 + 32 + 32
		if len(data) != expectedLen {
			t.Errorf("Expected %d bytes, got %d", expectedLen, len(data))
		}
	})
}

func TestRailViewResult(t *testing.T) {
	t.Run("should have all required fields", func(t *testing.T) {
		rail := RailViewResult{
			Token:               common.HexToAddress("0x1111111111111111111111111111111111111111"),
			From:                common.HexToAddress("0x2222222222222222222222222222222222222222"),
			To:                  common.HexToAddress("0x3333333333333333333333333333333333333333"),
			Operator:            common.HexToAddress("0x4444444444444444444444444444444444444444"),
			Validator:           common.HexToAddress("0x5555555555555555555555555555555555555555"),
			PaymentRate:         big.NewInt(1000000000000000000),
			LockupPeriod:        big.NewInt(2880),
			LockupFixed:         big.NewInt(0),
			SettledUpTo:         big.NewInt(1000000),
			EndEpoch:            big.NewInt(0),
			CommissionRateBps:   big.NewInt(500),
			ServiceFeeRecipient: common.HexToAddress("0x6666666666666666666666666666666666666666"),
		}

		if rail.PaymentRate.Cmp(big.NewInt(1000000000000000000)) != 0 {
			t.Error("PaymentRate mismatch")
		}
		if rail.LockupPeriod.Cmp(big.NewInt(2880)) != 0 {
			t.Error("LockupPeriod mismatch")
		}
		if rail.CommissionRateBps.Cmp(big.NewInt(500)) != 0 {
			t.Error("CommissionRateBps mismatch")
		}
		if rail.EndEpoch.Cmp(big.NewInt(0)) != 0 {
			t.Error("EndEpoch mismatch - expected 0 for active rail")
		}
	})
}

func TestRailInfoResult(t *testing.T) {
	t.Run("should represent active rail", func(t *testing.T) {
		rail := RailInfoResult{
			RailId:       big.NewInt(123),
			IsTerminated: false,
			EndEpoch:     big.NewInt(0),
		}

		if rail.IsTerminated {
			t.Error("Expected rail to not be terminated")
		}
		if rail.EndEpoch.Cmp(big.NewInt(0)) != 0 {
			t.Error("Active rail should have EndEpoch = 0")
		}
	})

	t.Run("should represent terminated rail", func(t *testing.T) {
		rail := RailInfoResult{
			RailId:       big.NewInt(456),
			IsTerminated: true,
			EndEpoch:     big.NewInt(2000000),
		}

		if !rail.IsTerminated {
			t.Error("Expected rail to be terminated")
		}
		if rail.EndEpoch.Cmp(big.NewInt(0)) <= 0 {
			t.Error("Terminated rail should have EndEpoch > 0")
		}
	})
}
