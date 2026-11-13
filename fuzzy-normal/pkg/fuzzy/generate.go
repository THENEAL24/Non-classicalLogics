package fuzzy

import (
	"fmt"
	"math/rand"
)

// GenOptions — параметры генерации набора 100–300 элементов.
type GenOptions struct {
	Count      int     // 100–300
	MeanMin    float64
	MeanMax    float64
	SigmaMin   float64
	SigmaMax   float64
	Seed       int64
	SetName    string
}

func (o *GenOptions) normalize() error {
	if o.Count < 100 {
		o.Count = 100
	}
	if o.Count > 300 {
		o.Count = 300
	}
	if o.MeanMax <= o.MeanMin {
		return fmt.Errorf("mean range invalid")
	}
	if o.SigmaMax <= o.SigmaMin || o.SigmaMin <= 0 {
		return fmt.Errorf("sigma range invalid")
	}
	if o.SetName == "" {
		o.SetName = "generated"
	}
	return nil
}

// GenerateSet создаёт случайное множество нечётких величин.
func GenerateSet(o GenOptions) (Set, error) {
	if err := o.normalize(); err != nil {
		return Set{}, err
	}
	r := rand.New(rand.NewSource(o.Seed))
	s := Set{Name: o.SetName, Values: make([]NormalFuzzy, 0, o.Count)}
	for i := 0; i < o.Count; i++ {
		mean := o.MeanMin + r.Float64()*(o.MeanMax-o.MeanMin)
		sigma := o.SigmaMin + r.Float64()*(o.SigmaMax-o.SigmaMin)
		s.Values = append(s.Values, MustNew(mean, sigma))
	}
	return s, nil
}
