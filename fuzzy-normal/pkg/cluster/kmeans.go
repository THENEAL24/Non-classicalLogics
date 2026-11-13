package cluster

import (
	"fmt"
	"math"
	"math/rand"

	"fuzzy-normal/pkg/fuzzy"
)

// Criterion — критерий расстояния при кластеризации.
type Criterion int

const (
	ByMeanOnly Criterion = iota
	ByMeanAndSigma
)

func (c Criterion) String() string {
	switch c {
	case ByMeanOnly:
		return "по центру (a)"
	case ByMeanAndSigma:
		return "по a и σ"
	default:
		return "unknown"
	}
}

// Result — результат кластерного анализа.
type Result struct {
	K          int
	Criterion  Criterion
	Clusters   [][]fuzzy.NormalFuzzy
	Centroids  []fuzzy.NormalFuzzy
	Labels     []int // индекс кластера для каждого элемента исходного множества
	Iterations int
}

// Config параметры k-means.
type Config struct {
	K         int
	Criterion Criterion
	MaxIter   int
	Seed      int64
	// SigmaWeight вес σ при критерии ByMeanAndSigma
	SigmaWeight float64
}

func (cfg *Config) defaults() {
	if cfg.K < 2 {
		cfg.K = 2
	}
	if cfg.MaxIter <= 0 {
		cfg.MaxIter = 100
	}
	if cfg.SigmaWeight <= 0 {
		cfg.SigmaWeight = 1
	}
}

// sigmaRangeWeight приводит вклад σ к масштабу a по размаху выборки (иначе при a∈[0,100], σ∈[0.5,8]
// евклидово расстояние с w=1 почти полностью определяется разностью центров).
func sigmaRangeWeight(vals []fuzzy.NormalFuzzy) float64 {
	if len(vals) == 0 {
		return 1
	}
	minA, maxA := vals[0].Mean, vals[0].Mean
	minS, maxS := vals[0].Sigma, vals[0].Sigma
	for _, v := range vals[1:] {
		if v.Mean < minA {
			minA = v.Mean
		}
		if v.Mean > maxA {
			maxA = v.Mean
		}
		if v.Sigma < minS {
			minS = v.Sigma
		}
		if v.Sigma > maxS {
			maxS = v.Sigma
		}
	}
	spanA := maxA - minA
	spanS := maxS - minS
	if spanA < 1e-12 || spanS < 1e-12 {
		return 1
	}
	return spanA / spanS
}

func dist(a, b fuzzy.NormalFuzzy, cfg Config) float64 {
	switch cfg.Criterion {
	case ByMeanOnly:
		return math.Abs(a.Mean - b.Mean)
	default:
		return fuzzy.Distance(a, b, cfg.SigmaWeight)
	}
}

// KMeans распределяет множество на K кластеров.
func KMeans(set fuzzy.Set, cfg Config) (Result, error) {
	cfg.defaults()
	if cfg.Criterion == ByMeanAndSigma {
		cfg.SigmaWeight *= sigmaRangeWeight(set.Values)
	}
	n := set.Len()
	if n == 0 {
		return Result{}, fmt.Errorf("empty set")
	}
	if cfg.K > n {
		return Result{}, fmt.Errorf("K=%d больше числа элементов %d", cfg.K, n)
	}

	r := rand.New(rand.NewSource(cfg.Seed))
	// начальные центроиды — случайные элементы
	perm := r.Perm(n)
	centroids := make([]fuzzy.NormalFuzzy, cfg.K)
	for i := 0; i < cfg.K; i++ {
		centroids[i] = set.Values[perm[i]]
	}
	labels := make([]int, n)

	var iter int
	for iter = 0; iter < cfg.MaxIter; iter++ {
		changed := false
		// назначение
		for i, v := range set.Values {
			best, bestD := 0, dist(v, centroids[0], cfg)
			for c := 1; c < cfg.K; c++ {
				d := dist(v, centroids[c], cfg)
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
		// пересчёт центроидов как среднее a и σ по кластеру
		newCent := make([]fuzzy.NormalFuzzy, cfg.K)
		counts := make([]int, cfg.K)
		sumA := make([]float64, cfg.K)
		sumS := make([]float64, cfg.K)
		for i, v := range set.Values {
			c := labels[i]
			sumA[c] += v.Mean
			sumS[c] += v.Sigma
			counts[c]++
		}
		for c := 0; c < cfg.K; c++ {
			if counts[c] == 0 {
				newCent[c] = set.Values[r.Intn(n)]
				changed = true
				continue
			}
			newCent[c] = fuzzy.MustNew(sumA[c]/float64(counts[c]), sumS[c]/float64(counts[c]))
		}
		centroids = newCent
		if !changed {
			iter++
			break
		}
	}

	clusters := make([][]fuzzy.NormalFuzzy, cfg.K)
	for i, v := range set.Values {
		clusters[labels[i]] = append(clusters[labels[i]], v)
	}

	return Result{
		K:          cfg.K,
		Criterion:  cfg.Criterion,
		Clusters:   clusters,
		Centroids:  centroids,
		Labels:     labels,
		Iterations: iter,
	}, nil
}

// FormatSummary — текстовый отчёт.
func (r Result) FormatSummary() string {
	var b fmtBuilder
	b.Printf("Кластерный анализ: K=%d, критерий: %s, итераций: %d\n",
		r.K, r.Criterion, r.Iterations)
	for i := 0; i < r.K; i++ {
		b.Printf("\nКластер %d (%d элементов), центроид %s\n",
			i+1, len(r.Clusters[i]), r.Centroids[i])
		for j := 0; j < len(r.Clusters[i]); j++ {
			b.Printf("  - %s\n", r.Clusters[i][j])
		}
	}
	return b.String()
}

type fmtBuilder struct{ s string }

func (f *fmtBuilder) Printf(format string, args ...any) {
	f.s += fmt.Sprintf(format, args...)
}

func (f *fmtBuilder) String() string { return f.s }
