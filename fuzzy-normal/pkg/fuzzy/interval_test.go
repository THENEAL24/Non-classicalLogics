package fuzzy

import (
	"math"
	"testing"
)

func TestAlphaCutHalfMax(t *testing.T) {
	f := MustNew(10, 2)
	iv, err := AlphaCut(f, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	wantLen := 2 * 2 * math.Sqrt(-math.Log(0.5)) // 2σ√(−ln α)
	if math.Abs(iv.Length-wantLen) > 1e-6 {
		t.Fatalf("length got %v want ~%v", iv.Length, wantLen)
	}
	if iv.Membership(10) != 1 {
		t.Fatal("mid should be in interval")
	}
}
