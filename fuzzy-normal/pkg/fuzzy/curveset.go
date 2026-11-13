package fuzzy

// CurveFromParams строит одну кривую μ(x) по параметрам нечёткой величины N(a, σ).
func CurveFromParams(f NormalFuzzy, points int) []CurvePoint {
	return SampleCurve(f, points)
}

// CurveMeta — одна из N кривых для общего графика.
type CurveMeta struct {
	Index   int          `json:"index"`
	Source  NormalFuzzy  `json:"source"`
	Preview []CurvePoint `json:"preview"`
}

// CurveBank — все кривые набора.
type CurveBank struct {
	Count          int         `json:"count"`
	PointsPerCurve int         `json:"pointsPerCurve"`
	PreviewPoints  int         `json:"previewPoints"`
	Items          []CurveMeta `json:"items"`
}

// BuildCurveBank строит сглаженную сетку точек для каждой кривой (без грубого прореживания).
func BuildCurveBank(set Set, pointsPerCurve, previewPoints int) CurveBank {
	if previewPoints < pointsPerCurve {
		previewPoints = pointsPerCurve
	}
	if previewPoints < 80 {
		previewPoints = 80
	}
	items := make([]CurveMeta, set.Len())
	for i, f := range set.Values {
		items[i] = CurveMeta{
			Index:   i,
			Source:  f,
			Preview: CurveFromParams(f, previewPoints),
		}
	}
	return CurveBank{
		Count:          set.Len(),
		PointsPerCurve: pointsPerCurve,
		PreviewPoints:  previewPoints,
		Items:          items,
	}
}
