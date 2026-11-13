package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fuzzy-normal/pkg/cluster"
	"fuzzy-normal/pkg/fuzzy"
)

type server struct {
	mu      sync.Mutex
	set     fuzzy.Set
	journal *fuzzy.Journal
}

func main() {
	logPath := env("LOG_PATH", "fuzzy_journal.log")
	s := &server{
		set:     fuzzy.NewSet("dataset"),
		journal: fuzzy.NewJournal(logPath),
	}
	s.journal.Append("Запуск веб-сервера")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/journal", s.handleJournal)
	mux.HandleFunc("GET /api/set", s.handleGetSet)
	mux.HandleFunc("POST /api/set/generate", s.handleGenerate)
	mux.HandleFunc("POST /api/ops", s.handleOps)
	mux.HandleFunc("POST /api/membership", s.handleMembership)
	mux.HandleFunc("POST /api/cluster", s.handleCluster)
	mux.HandleFunc("POST /api/cluster/refine", s.handleClusterRefine)
	mux.HandleFunc("POST /api/intervals/analyze", s.handleIntervalAnalyze)
	mux.HandleFunc("POST /api/intervals/cluster", s.handleIntervalCluster)
	mux.HandleFunc("POST /api/intervals/cluster/refine", s.handleIntervalClusterRefine)

	staticDir := env("STATIC_DIR", "./web/dist")
	if _, err := os.Stat(staticDir); err != nil {
		log.Printf("warning: static dir %q not found (build frontend or set STATIC_DIR)", staticDir)
	}
	mux.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel == "" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		target := filepath.Join(staticDir, filepath.Clean(rel))
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			http.ServeFile(w, r, target)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	addr := env("ADDR", ":8080")
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type fuzzyDTO struct {
	Mean  float64 `json:"mean"`
	Sigma float64 `json:"sigma"`
}

func (d fuzzyDTO) toFuzzy() (fuzzy.NormalFuzzy, error) {
	return fuzzy.New(d.Mean, d.Sigma)
}

func fromFuzzy(f fuzzy.NormalFuzzy) fuzzyDTO {
	return fuzzyDTO{Mean: f.Mean, Sigma: f.Sigma}
}

func normalClustersJSON(res cluster.Result) []map[string]any {
	clusters := make([]map[string]any, res.K)
	for i := 0; i < res.K; i++ {
		items := make([]fuzzyDTO, len(res.Clusters[i]))
		for j, v := range res.Clusters[i] {
			items[j] = fromFuzzy(v)
		}
		clusters[i] = map[string]any{
			"index":    i + 1,
			"size":     len(res.Clusters[i]),
			"centroid": fromFuzzy(res.Centroids[i]),
			"items":    items,
			"summary":  fmt.Sprintf("%d элементов, центроид %s", len(res.Clusters[i]), res.Centroids[i]),
		}
	}
	return clusters
}

func normalClusterStage(k int, crit cluster.Criterion, res cluster.Result) map[string]any {
	return map[string]any{
		"k":          res.K,
		"criterion":  crit.String(),
		"iterations": res.Iterations,
		"clusters":   normalClustersJSON(res),
		"report":     res.FormatSummary(),
	}
}

func intervalClustersJSON(res cluster.IntervalResult) []map[string]any {
	clusters := make([]map[string]any, res.K)
	for i := 0; i < res.K; i++ {
		clusters[i] = map[string]any{
			"index":    i + 1,
			"size":     len(res.Clusters[i]),
			"centroid": res.Centroids[i],
			"items":    res.Clusters[i],
		}
	}
	return clusters
}

func intervalClusterStage(k int, crit cluster.IntervalCriterion, res cluster.IntervalResult) map[string]any {
	return map[string]any{
		"k":          res.K,
		"criterion":  crit.String(),
		"iterations": res.Iterations,
		"clusters":   intervalClustersJSON(res),
		"report":     res.FormatSummary(),
	}
}

func (s *server) handleJournal(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	text := s.journal.Text()
	s.mu.Unlock()
	writeJSON(w, map[string]string{"text": text})
}

func (s *server) handleGetSet(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	vals := make([]fuzzyDTO, len(set.Values))
	for i, v := range set.Values {
		vals[i] = fromFuzzy(v)
	}
	writeJSON(w, map[string]any{
		"name":  set.Name,
		"count": len(vals),
		"items": vals,
	})
}

type generateReq struct {
	Count    int     `json:"count"`
	MeanMin  float64 `json:"meanMin"`
	MeanMax  float64 `json:"meanMax"`
	SigmaMin float64 `json:"sigmaMin"`
	SigmaMax float64 `json:"sigmaMax"`
}

func (s *server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req generateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Count == 0 {
		req.Count = 150
	}
	if req.MeanMax == 0 {
		req.MeanMin, req.MeanMax = 0, 100
	}
	if req.SigmaMax == 0 {
		req.SigmaMin, req.SigmaMax = 0.5, 8
	}
	set, err := fuzzy.GenerateSet(fuzzy.GenOptions{
		Count:    req.Count,
		MeanMin:  req.MeanMin,
		MeanMax:  req.MeanMax,
		SigmaMin: req.SigmaMin,
		SigmaMax: req.SigmaMax,
		Seed:     time.Now().UnixNano(),
		SetName:  "dataset",
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	s.mu.Lock()
	s.set = set
	s.journal.Append("Генерация: %d нечётких величин-параметров кривых (a, σ)", set.Len())
	s.mu.Unlock()
	s.handleGetSet(w, r)
}

type opsReq struct {
	Op string  `json:"op"`
	A  fuzzyDTO `json:"a"`
	B  fuzzyDTO `json:"b"`
}

func (s *server) handleOps(w http.ResponseWriter, r *http.Request) {
	var req opsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	va, err := req.A.toFuzzy()
	if err != nil {
		writeErr(w, http.StatusBadRequest, "A: "+err.Error())
		return
	}
	vb, err := req.B.toFuzzy()
	if err != nil {
		writeErr(w, http.StatusBadRequest, "B: "+err.Error())
		return
	}
	var res fuzzy.NormalFuzzy
	switch req.Op {
	case "add":
		res = fuzzy.Add(va, vb)
	case "sub":
		res = fuzzy.Sub(va, vb)
	case "mul":
		res = fuzzy.Mul(va, vb)
	case "div":
		res, err = fuzzy.Div(va, vb)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
	case "compare":
		info := fuzzy.ExplainCompare(va, vb)
		s.mu.Lock()
		s.journal.Append("Сравнение: N₁=%s N₂=%s → %s (%s)", va, vb, info.Relation, info.Title)
		s.mu.Unlock()
		writeJSON(w, map[string]any{
			"compare":       info.ResultText,
			"compareDetail": info,
			"a":             fromFuzzy(va),
			"b":             fromFuzzy(vb),
		})
		return
	default:
		writeErr(w, http.StatusBadRequest, "unknown op")
		return
	}
	s.mu.Lock()
	s.journal.Append("%s: A=%s B=%s → %s", req.Op, va, vb, res)
	s.mu.Unlock()
	cmp := fuzzy.ExplainCompare(va, vb)
	writeJSON(w, map[string]any{
		"result":        fromFuzzy(res),
		"compare":       cmp.ResultText,
		"compareDetail": cmp,
		"arithDetail":   fuzzy.ExplainArithmetic(req.Op, va, vb, res),
	})
}

type membershipReq struct {
	Fuzzy fuzzyDTO `json:"fuzzy"`
	X     float64  `json:"x"`
}

func (s *server) handleMembership(w http.ResponseWriter, r *http.Request) {
	var req membershipReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	f, err := req.Fuzzy.toFuzzy()
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	mu := f.Membership(req.X)
	s.mu.Lock()
	s.journal.Append("μ(%.4f)=%.6f, %s", req.X, mu, f)
	s.mu.Unlock()
	writeJSON(w, map[string]any{"mu": mu, "x": req.X})
}

type clusterReq struct {
	K         int     `json:"k"`
	Criterion string  `json:"criterion"`
	Beta      float64 `json:"beta"`
	Gamma     float64 `json:"gamma"`
}

func clusterCriterion(s string) cluster.Criterion {
	if s == "mean" {
		return cluster.ByMeanOnly
	}
	return cluster.ByMeanAndSigma
}

func (s *server) handleCluster(w http.ResponseWriter, r *http.Request) {
	var req clusterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	if set.Len() == 0 {
		writeErr(w, http.StatusBadRequest, "сначала сгенерируйте множество")
		return
	}
	crit := clusterCriterion(req.Criterion)
	baseRes, err := cluster.KMeans(set, cluster.Config{
		K:         req.K,
		Criterion: crit,
		Seed:      time.Now().UnixNano(),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	s.mu.Lock()
	s.journal.Append("Кластеризация K=%d (%s), итераций %d, N=%d", req.K, crit, baseRes.Iterations, set.Len())
	s.mu.Unlock()

	stage := normalClusterStage(req.K, crit, baseRes)
	writeJSON(w, map[string]any{
		"k":          baseRes.K,
		"criterion":  crit.String(),
		"iterations": baseRes.Iterations,
		"clusters":   stage["clusters"],
		"report":     baseRes.FormatSummary(),
		"baseline":   stage,
	})
}

func (s *server) handleClusterRefine(w http.ResponseWriter, r *http.Request) {
	var req clusterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	if set.Len() == 0 {
		writeErr(w, http.StatusBadRequest, "сначала сгенерируйте множество")
		return
	}
	crit := clusterCriterion(req.Criterion)
	filtered, frep, err := fuzzy.FilterSetByMagnitude(set, fuzzy.MagnitudeFilterOpts{
		Beta: req.Beta, Gamma: req.Gamma, NormalCriterion: req.Criterion,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out := map[string]any{"filter": frep}
	if !frep.Changed() {
		writeJSON(w, out)
		return
	}
	refRes, err := cluster.KMeans(filtered, cluster.Config{
		K:         req.K,
		Criterion: crit,
		Seed:      time.Now().UnixNano(),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out["refined"] = normalClusterStage(req.K, crit, refRes)
	writeJSON(w, out)
}

type intervalAnalyzeReq struct {
	Alpha float64 `json:"alpha"`
}

func (s *server) handleIntervalAnalyze(w http.ResponseWriter, r *http.Request) {
	var req intervalAnalyzeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	if set.Len() == 0 {
		writeErr(w, http.StatusBadRequest, "сначала сгенерируйте множество")
		return
	}
	report, err := fuzzy.AnalyzeSet(set, fuzzy.AnalyzeOptions{
		Alpha:          req.Alpha,
		PointsPerCurve: 80,
		PreviewPoints:  120,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	s.mu.Lock()
	s.journal.Append("Построено %d кривых μ(x); α=%.3f → %d интервалов [A,B]",
		report.CurveCount, report.Alpha, report.IntervalCount)
	s.mu.Unlock()
	writeJSON(w, report)
}

type intervalClusterReq struct {
	Alpha     float64 `json:"alpha"`
	K         int     `json:"k"`
	Criterion string  `json:"criterion"` // length | ab | midlen
	Beta      float64 `json:"beta"`
	Gamma     float64 `json:"gamma"`
}

func intervalCriterion(s string) cluster.IntervalCriterion {
	switch s {
	case "ab":
		return cluster.ByIntervalAB
	case "midlen":
		return cluster.ByMidAndLength
	default:
		return cluster.ByIntervalLength
	}
}

func (s *server) handleIntervalCluster(w http.ResponseWriter, r *http.Request) {
	var req intervalClusterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	if set.Len() == 0 {
		writeErr(w, http.StatusBadRequest, "сначала сгенерируйте множество")
		return
	}
	alpha := req.Alpha
	if alpha <= 0 {
		alpha = 0.5
	}
	iv, err := fuzzy.IntervalsFromSet(set, alpha)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	crit := intervalCriterion(req.Criterion)
	baseRes, err := cluster.KMeansIntervals(iv, cluster.IntervalConfig{
		K:         req.K,
		Criterion: crit,
		Seed:      time.Now().UnixNano(),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	s.mu.Lock()
	s.journal.Append("Кластеризация интервалов α=%.3f, K=%d (%s), N=%d", alpha, req.K, crit, iv.Len())
	s.mu.Unlock()

	stage := intervalClusterStage(req.K, crit, baseRes)
	writeJSON(w, map[string]any{
		"alpha":      alpha,
		"k":          baseRes.K,
		"criterion":  crit.String(),
		"iterations": baseRes.Iterations,
		"clusters":   stage["clusters"],
		"report":     baseRes.FormatSummary(),
		"baseline":   stage,
	})
}

func (s *server) handleIntervalClusterRefine(w http.ResponseWriter, r *http.Request) {
	var req intervalClusterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	s.mu.Lock()
	set := s.set.Clone()
	s.mu.Unlock()
	if set.Len() == 0 {
		writeErr(w, http.StatusBadRequest, "сначала сгенерируйте множество")
		return
	}
	alpha := req.Alpha
	if alpha <= 0 {
		alpha = 0.5
	}
	iv, err := fuzzy.IntervalsFromSet(set, alpha)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	crit := intervalCriterion(req.Criterion)
	filtered, frep, err := fuzzy.FilterIntervalSetByMagnitude(iv, fuzzy.MagnitudeFilterOpts{
		Beta: req.Beta, Gamma: req.Gamma, IntervalCriterion: req.Criterion,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out := map[string]any{"filter": frep, "alpha": alpha}
	if !frep.Changed() {
		writeJSON(w, out)
		return
	}
	refRes, err := cluster.KMeansIntervals(filtered, cluster.IntervalConfig{
		K:         req.K,
		Criterion: crit,
		Seed:      time.Now().UnixNano(),
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	out["refined"] = intervalClusterStage(req.K, crit, refRes)
	writeJSON(w, out)
}
