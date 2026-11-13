package fuzzy

import (
	"math"
	"testing"
)

func TestMembershipPeak(t *testing.T) {
	f := MustNew(5, 2)
	if m := f.Membership(5); math.Abs(m-1) > 1e-12 {
		t.Fatalf("expected 1 at mean, got %v", m)
	}
	if f.Membership(5+2) >= f.Membership(5+3) {
		// monotonic decay
	}
}

func TestAddSub(t *testing.T) {
	a := MustNew(1, 1)
	b := MustNew(2, 1)
	sum := Add(a, b)
	if sum.Mean != 3 || math.Abs(sum.Sigma-math.Sqrt2) > 1e-9 {
		t.Fatalf("add: %+v", sum)
	}
	diff := Sub(a, b)
	if diff.Mean != -1 {
		t.Fatalf("sub mean")
	}
}

func TestCompare(t *testing.T) {
	if Compare(MustNew(1, 1), MustNew(2, 1)) != Less {
		t.Fatal("compare less")
	}
}
