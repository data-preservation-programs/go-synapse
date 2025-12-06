package storage

import (
	"testing"

	"github.com/ipfs/go-cid"
)

var zeroPieceCidFixtures = []struct {
	RawSize    int
	PaddedSize int
	V1PieceCID string
}{
	{96, 128, "baga6ea4seaqdomn3tgwgrh3g532zopskstnbrd2n3sxfqbze7rxt7vqn7veigmy"},
	{126, 128, "baga6ea4seaqdomn3tgwgrh3g532zopskstnbrd2n3sxfqbze7rxt7vqn7veigmy"},
	{127, 128, "baga6ea4seaqdomn3tgwgrh3g532zopskstnbrd2n3sxfqbze7rxt7vqn7veigmy"},
	{192, 256, "baga6ea4seaqgiktap34inmaex4wbs6cghlq5i2j2yd2bb2zndn5ep7ralzphkdy"},
	{253, 256, "baga6ea4seaqgiktap34inmaex4wbs6cghlq5i2j2yd2bb2zndn5ep7ralzphkdy"},
	{254, 256, "baga6ea4seaqgiktap34inmaex4wbs6cghlq5i2j2yd2bb2zndn5ep7ralzphkdy"},
	{255, 512, "baga6ea4seaqfpirydiugkk7up5v666wkm6n6jlw6lby2wxht5mwaqekerdfykjq"},
	{256, 512, "baga6ea4seaqfpirydiugkk7up5v666wkm6n6jlw6lby2wxht5mwaqekerdfykjq"},
	{384, 512, "baga6ea4seaqfpirydiugkk7up5v666wkm6n6jlw6lby2wxht5mwaqekerdfykjq"},
	{507, 512, "baga6ea4seaqfpirydiugkk7up5v666wkm6n6jlw6lby2wxht5mwaqekerdfykjq"},
	{508, 512, "baga6ea4seaqfpirydiugkk7up5v666wkm6n6jlw6lby2wxht5mwaqekerdfykjq"},
	{509, 1024, "baga6ea4seaqb66wjlfkrbye6uqoemcyxmqylwmrm235uclwfpsyx3ge2imidoly"},
	{512, 1024, "baga6ea4seaqb66wjlfkrbye6uqoemcyxmqylwmrm235uclwfpsyx3ge2imidoly"},
	{768, 1024, "baga6ea4seaqb66wjlfkrbye6uqoemcyxmqylwmrm235uclwfpsyx3ge2imidoly"},
	{1015, 1024, "baga6ea4seaqb66wjlfkrbye6uqoemcyxmqylwmrm235uclwfpsyx3ge2imidoly"},
	{1016, 1024, "baga6ea4seaqb66wjlfkrbye6uqoemcyxmqylwmrm235uclwfpsyx3ge2imidoly"},
	{1017, 2048, "baga6ea4seaqpy7usqklokfx2vxuynmupslkeutzexe2uqurdg5vhtebhxqmpqmy"},
	{1024, 2048, "baga6ea4seaqpy7usqklokfx2vxuynmupslkeutzexe2uqurdg5vhtebhxqmpqmy"},
}

func TestCalculatePieceCID_ZeroData(t *testing.T) {
	for _, fixture := range zeroPieceCidFixtures {
		t.Run("", func(t *testing.T) {
			zeroBytes := make([]byte, fixture.RawSize)

			pieceCID, err := CalculatePieceCID(zeroBytes)
			if err != nil {
				t.Fatalf("CalculatePieceCID failed for size %d: %v", fixture.RawSize, err)
			}

			expectedV1, err := cid.Decode(fixture.V1PieceCID)
			if err != nil {
				t.Fatalf("Failed to parse expected CID %s: %v", fixture.V1PieceCID, err)
			}

			if pieceCID.String() != expectedV1.String() {
				t.Errorf("PieceCID mismatch for size %d:\nExpected: %s\nActual:   %s",
					fixture.RawSize, expectedV1.String(), pieceCID.String())
			}
		})
	}
}

func TestCalculatePieceCID_NonZeroData(t *testing.T) {
	data1 := []byte("Hello, World!")
	data2 := []byte("Hello, World?")

	cid1, err := CalculatePieceCID(data1)
	if err != nil {
		t.Fatalf("CalculatePieceCID failed for data1: %v", err)
	}

	cid2, err := CalculatePieceCID(data2)
	if err != nil {
		t.Fatalf("CalculatePieceCID failed for data2: %v", err)
	}

	if cid1.String() == cid2.String() {
		t.Error("Different data produced the same PieceCID")
	}
}

func TestCalculatePieceCID_Deterministic(t *testing.T) {
	data := []byte("Deterministic test data")

	cid1, err := CalculatePieceCID(data)
	if err != nil {
		t.Fatalf("First CalculatePieceCID failed: %v", err)
	}

	cid2, err := CalculatePieceCID(data)
	if err != nil {
		t.Fatalf("Second CalculatePieceCID failed: %v", err)
	}

	if cid1.String() != cid2.String() {
		t.Errorf("Same data produced different CIDs:\nFirst:  %s\nSecond: %s", cid1.String(), cid2.String())
	}
}

func TestCalculatePieceCID_VariousSizes(t *testing.T) {
	testSizes := []int{1, 50, 100, 127, 128, 500, 1000, 2048, 4096, 8192}

	for _, size := range testSizes {
		t.Run("", func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			pieceCID, err := CalculatePieceCID(data)
			if err != nil {
				t.Fatalf("CalculatePieceCID failed for size %d: %v", size, err)
			}

			if pieceCID.String() == "" {
				t.Errorf("Empty CID string for size %d", size)
			}

			_, err = cid.Decode(pieceCID.String())
			if err != nil {
				t.Errorf("Generated CID cannot be decoded for size %d: %v", size, err)
			}
		})
	}
}

func TestCalculatePieceCID_EmptyData(t *testing.T) {
	emptyData := []byte{}

	_, err := CalculatePieceCID(emptyData)
	if err == nil {
		t.Error("Expected error for empty data, but got nil")
	}
}
