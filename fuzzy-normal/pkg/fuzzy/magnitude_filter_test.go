package fuzzy

import "testing"

func TestFilterSetBySigmaBeta(t *testing.T) {
	set := Set{
		Values: []NormalFuzzy{
			MustNew(1, 2),
			MustNew(2, 10),
			MustNew(3, 5),
		},
	}
	out, rep, err := FilterSetBySigmaBeta(set, 0.8)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Before != 3 || rep.After != 1 {
		t.Fatalf("count: before=%d after=%d", rep.Before, rep.After)
	}
	if rep.Threshold != 8 || rep.Max != 10 {
		t.Fatalf("threshold=%v max=%v", rep.Threshold, rep.Max)
	}
	if len(out.Values) != 1 || out.Values[0].Sigma != 10 {
		t.Fatalf("unexpected filter result: %+v", out.Values)
	}
	_, rep2, err := FilterSetBySigmaBeta(set, 1)
	if err != nil || rep2.After != 3 {
		t.Fatalf("beta=1: err=%v after=%d", err, rep2.After)
	}
}

func TestFilterSetLowBeta(t *testing.T) {
	set := Set{
		Values: []NormalFuzzy{
			MustNew(1, 1),
			MustNew(2, 5),
			MustNew(3, 10),
		},
	}
	_, rep, err := FilterSetBySigmaBeta(set, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	if rep.After != 3 {
		t.Fatalf("beta=0.1 should keep all (порог 1.0), after=%d", rep.After)
	}
}

func TestFilterSetGammaRho(t *testing.T) {
	set := Set{
		Values: []NormalFuzzy{
			MustNew(10, 8),
			MustNew(20, 2),
			MustNew(30, 4),
			MustNew(40, 1),
		},
	}
	_, rep, err := FilterSetByMagnitude(set, MagnitudeFilterOpts{
		Beta: 1, Gamma: 0.5, NormalCriterion: "both",
	})
	if err != nil {
		t.Fatal(err)
	}
	if rep.After != 2 {
		t.Fatalf("gamma=0.5 want 2 items, got %d", rep.After)
	}
}

func TestFilterSetByMeanBeta(t *testing.T) {
	set := Set{
		Values: []NormalFuzzy{
			MustNew(5, 1),
			MustNew(50, 10),
			MustNew(40, 2),
		},
	}
	out, rep, err := FilterSetByMagnitude(set, MagnitudeFilterOpts{
		Beta: 0.9, Gamma: 1, NormalCriterion: "mean",
	})
	if err != nil {
		t.Fatal(err)
	}
	if rep.After != 1 || rep.Metric != "mean" || out.Values[0].Mean != 50 {
		t.Fatalf("want 1 by max a=50; after=%d metric=%s", rep.After, rep.Metric)
	}
}

func TestFilterIntervalSetByLengthBeta(t *testing.T) {
	set := IntervalSet{
		Alpha: 0.5,
		Items: []IntervalFuzzy{
			{A: 0, B: 4, Length: 4},
			{A: 0, B: 10, Length: 10},
			{A: 0, B: 7, Length: 7},
		},
	}
	out, rep, err := FilterIntervalSetByLengthBeta(set, 0.8)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Before != 3 || rep.After != 1 || rep.Threshold != 8 {
		t.Fatalf("rep=%+v", rep)
	}
	if len(out.Items) != 1 || out.Items[0].Length != 10 {
		t.Fatalf("items=%v", out.Items)
	}
}
