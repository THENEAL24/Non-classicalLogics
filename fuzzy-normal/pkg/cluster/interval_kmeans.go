package cluster

import (
	"fmt"
	"math"
	"math/rand"

	"fuzzy-normal/pkg/fuzzy"
)

// IntervalCriterion — критерий кластеризации интервалов [A, B].
type IntervalCriterion int

const (
	ByIntervalLength IntervalCriterion = iota
	ByIntervalAB
	ByMidAndLength
)

func (c IntervalCriterion) String() string {
	switch c {
	case ByIntervalLength:
		return "по длине L=B−A"
	case ByIntervalAB:
		return "по границам A и B"
	case ByMidAndLength:
		return "по середине и длине"
	default:
		return "unknown"
	}
}

// IntervalResult — кластеризация интервальных нечётких величин.
type IntervalResult struct {
	K          int
	Criterion  IntervalCriterion
	Clusters   [][]fuzzy.IntervalFuzzy
	Centroids  []fuzzy.IntervalFuzzy
	Labels     []int
	Iterations int
}

type IntervalConfig struct {
	K         int
	Criterion IntervalCriterion
	MaxIter   int
	Seed      int64
}

func (cfg *IntervalConfig) defaults() {
	if cfg.K < 2 {
		cfg.K = 2
	}
	if cfg.MaxIter <= 0 {
		cfg.MaxIter = 100
	}
}

func intervalDist(a, b fuzzy.IntervalFuzzy, c IntervalCriterion) float64 {
	switch c {
	case ByIntervalLength:
		return math.Abs(a.Length - b.Length)
	case ByIntervalAB:
		dA := a.A - b.A
		dB := a.B - b.B
		return math.Hypot(dA, dB)
	default:
		dM := a.Mid - b.Mid
		dL := a.Length - b.Length
		return math.Hypot(dM, dL)
	}
}

func intervalCentroid(items []fuzzy.IntervalFuzzy, alpha float64) fuzzy.IntervalFuzzy {
	if len(items) == 0 {
		return fuzzy.IntervalFuzzy{}
	}
	var sA, sB, sL, sM float64
	for _, it := range items {
		sA += it.A
		sB += it.B
		sL += it.Length
		sM += it.Mid
	}
	n := float64(len(items))
	cA, cB := sA/n, sB/n
	return fuzzy.IntervalFuzzy{
		A: cA, B: cB, Alpha: alpha,
		Length: sL / n, Mid: sM / n,
	}
}

// KMeansIntervals кластеризует интервалы [A, B].
func KMeansIntervals(set fuzzy.IntervalSet, cfg IntervalConfig) (IntervalResult, error) {
	cfg.defaults()
	n := set.Len()
	if n == 0 {
		return IntervalResult{}, fmt.Errorf("пустое множество интервалов")
	}
	if cfg.K > n {
		return IntervalResult{}, fmt.Errorf("K=%d больше числа интервалов %d", cfg.K, n)
	}
	items := set.Items
	r := rand.New(rand.NewSource(cfg.Seed))
	perm := r.Perm(n)
	centroids := make([]fuzzy.IntervalFuzzy, cfg.K)
	for i := 0; i < cfg.K; i++ {
		centroids[i] = items[perm[i]]
	}
	labels := make([]int, n)
	var iter int
	for iter = 0; iter < cfg.MaxIter; iter++ {
		changed := false
		for i, v := range items {
			best, bestD := 0, intervalDist(v, centroids[0], cfg.Criterion)
			for c := 1; c < cfg.K; c++ {
				d := intervalDist(v, centroids[c], cfg.Criterion)
				if d < bestD {
					bestD = d
					best = c
				}
			}
			if labels[i] != best {
				changed = true
				labels[i] = best
			}
		}
		newCent := make([]fuzzy.IntervalFuzzy, cfg.K)
		counts := make([]int, cfg.K)
		buckets := make([][]fuzzy.IntervalFuzzy, cfg.K)
		for i, v := range items {
			c := labels[i]
			buckets[c] = append(buckets[c], v)
			counts[c]++
		}
		for c := 0; c < cfg.K; c++ {
			if counts[c] == 0 {
				newCent[c] = items[r.Intn(n)]
				changed = true
				continue
			}
			newCent[c] = intervalCentroid(buckets[c], set.Alpha)
		}
		centroids = newCent
		if !changed {
			iter++
			break
		}
	}
	clusters := make([][]fuzzy.IntervalFuzzy, cfg.K)
	for i, v := range items {
		clusters[labels[i]] = append(clusters[labels[i]], v)
	}
	return IntervalResult{
		K: cfg.K, Criterion: cfg.Criterion,
		Clusters: clusters, Centroids: centroids,
		Labels: labels, Iterations: iter,
	}, nil
}

func (r IntervalResult) FormatSummary() string {
	var b fmtBuilder
	b.Printf("Кластеризация интервалов: K=%d, %s, итераций %d\n", r.K, r.Criterion, r.Iterations)
	for i := 0; i < r.K; i++ {
		b.Printf("\nКластер %d (%d интервалов), центроид [%.4f, %.4f], L=%.4f\n",
			i+1, len(r.Clusters[i]), r.Centroids[i].A, r.Centroids[i].B, r.Centroids[i].Length)
		for j := 0; j < len(r.Clusters[i]); j++ {
			b.Printf("  - %s\n", r.Clusters[i][j])
		}
	}
	return b.String()
}
