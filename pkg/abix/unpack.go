// Package abix contains small utilities around go-ethereum's accounts/abi
// that work around or guard against fragile patterns observed in practice.
package abix

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// UnpackSingleTuple decodes an ABI method's single-tuple return into dst via
// abi.Unpack + json round-trip. UnpackIntoInterface mishandles this shape;
// Unpack returns the right anonymous struct, json copies it into dst by
// matching json tags. dst must be a pointer to a tagged struct.
func UnpackSingleTuple(parsed abi.ABI, method string, payload []byte, dst any) error {
	out, err := parsed.Unpack(method, payload)
	if err != nil {
		return err
	}
	if len(out) != 1 {
		return fmt.Errorf("%s: expected 1 output, got %d", method, len(out))
	}
	buf, err := json.Marshal(out[0])
	if err != nil {
		return fmt.Errorf("%s: marshal unpacked tuple: %w", method, err)
	}
	if err := json.Unmarshal(buf, dst); err != nil {
		return fmt.Errorf("%s: decode into %T: %w", method, dst, err)
	}
	return nil
}
