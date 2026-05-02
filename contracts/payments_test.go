package contracts

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/data-preservation-programs/go-synapse/pkg/abix"
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

// TestUnpackRail_GetRail exercises the unpack path GetRail uses against a
// synthetic return blob. Reproduces a regression if abix.UnpackSingleTuple
// or getRailOutput's json tags fall out of sync with the ABI.
func TestUnpackRail_GetRail(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(PaymentsABIJSON))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}
	method, ok := parsedABI.Methods["getRail"]
	if !ok {
		t.Fatalf("getRail not found in ABI")
	}

	type railT struct {
		Token               common.Address `abi:"token"`
		From                common.Address `abi:"from"`
		To                  common.Address `abi:"to"`
		Operator            common.Address `abi:"operator"`
		Validator           common.Address `abi:"validator"`
		PaymentRate         *big.Int       `abi:"paymentRate"`
		LockupPeriod        *big.Int       `abi:"lockupPeriod"`
		LockupFixed         *big.Int       `abi:"lockupFixed"`
		SettledUpTo         *big.Int       `abi:"settledUpTo"`
		EndEpoch            *big.Int       `abi:"endEpoch"`
		CommissionRateBps   *big.Int       `abi:"commissionRateBps"`
		ServiceFeeRecipient common.Address `abi:"serviceFeeRecipient"`
	}
	want := railT{
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
	payload, err := method.Outputs.Pack(want)
	if err != nil {
		t.Fatalf("pack synthetic return: %v", err)
	}

	var got getRailOutput
	if err := abix.UnpackSingleTuple(parsedABI, "getRail", payload, &got); err != nil {
		t.Fatalf("UnpackSingleTuple: %v", err)
	}
	if got.Token != want.Token {
		t.Errorf("Token = %s, want %s", got.Token, want.Token)
	}
	if got.From != want.From {
		t.Errorf("From = %s, want %s", got.From, want.From)
	}
	if got.PaymentRate == nil || got.PaymentRate.Cmp(want.PaymentRate) != 0 {
		t.Errorf("PaymentRate = %v, want %v", got.PaymentRate, want.PaymentRate)
	}
	if got.CommissionRateBps == nil || got.CommissionRateBps.Cmp(want.CommissionRateBps) != 0 {
		t.Errorf("CommissionRateBps = %v, want %v", got.CommissionRateBps, want.CommissionRateBps)
	}
	if got.ServiceFeeRecipient != want.ServiceFeeRecipient {
		t.Errorf("ServiceFeeRecipient = %s, want %s", got.ServiceFeeRecipient, want.ServiceFeeRecipient)
	}
}

// TestUnpackRail_GetRailsForPayerAndToken exercises the unpack path
// GetRailsForPayerAndToken uses against a synthetic return blob.
// getRailsForPayerAndToken returns 3 outputs (results[], nextOffset, total)
// so the json round-trip applies to values[0] only.
func TestUnpackRail_GetRailsForPayerAndToken(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(PaymentsABIJSON))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}
	method, ok := parsedABI.Methods["getRailsForPayerAndToken"]
	if !ok {
		t.Fatalf("getRailsForPayerAndToken not found in ABI")
	}

	type itemT struct {
		RailId       *big.Int `abi:"railId"`
		IsTerminated bool     `abi:"isTerminated"`
		EndEpoch     *big.Int `abi:"endEpoch"`
	}
	results := []itemT{
		{RailId: big.NewInt(7), IsTerminated: false, EndEpoch: big.NewInt(0)},
		{RailId: big.NewInt(11), IsTerminated: true, EndEpoch: big.NewInt(900000)},
	}
	nextOffset := big.NewInt(2)
	total := big.NewInt(2)

	payload, err := method.Outputs.Pack(results, nextOffset, total)
	if err != nil {
		t.Fatalf("pack synthetic return: %v", err)
	}

	values, err := parsedABI.Unpack("getRailsForPayerAndToken", payload)
	if err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}

	buf, err := json.Marshal(values[0])
	if err != nil {
		t.Fatalf("marshal results: %v", err)
	}
	var rawResults []getRailsForPayerAndTokenItem
	if err := json.Unmarshal(buf, &rawResults); err != nil {
		t.Fatalf("decode results: %v", err)
	}
	if len(rawResults) != 2 {
		t.Fatalf("len = %d, want 2", len(rawResults))
	}
	if rawResults[0].RailId == nil || rawResults[0].RailId.Cmp(big.NewInt(7)) != 0 {
		t.Errorf("results[0].RailId = %v, want 7", rawResults[0].RailId)
	}
	if rawResults[0].IsTerminated {
		t.Errorf("results[0].IsTerminated = true, want false")
	}
	if rawResults[1].RailId == nil || rawResults[1].RailId.Cmp(big.NewInt(11)) != 0 {
		t.Errorf("results[1].RailId = %v, want 11", rawResults[1].RailId)
	}
	if !rawResults[1].IsTerminated {
		t.Errorf("results[1].IsTerminated = false, want true")
	}
	if rawResults[1].EndEpoch == nil || rawResults[1].EndEpoch.Cmp(big.NewInt(900000)) != 0 {
		t.Errorf("results[1].EndEpoch = %v, want 900000", rawResults[1].EndEpoch)
	}

	gotNextOffset, ok := values[1].(*big.Int)
	if !ok || gotNextOffset.Cmp(nextOffset) != 0 {
		t.Errorf("nextOffset = %v, want %v", values[1], nextOffset)
	}
	gotTotal, ok := values[2].(*big.Int)
	if !ok || gotTotal.Cmp(total) != 0 {
		t.Errorf("total = %v, want %v", values[2], total)
	}
}
