package review

type Context struct {
	MPS            float64
	AccuracyRate   float64 // 0.0 - 1.0
	TotalReviews   int
	LastReviewType string // "mcq", "match", "typing", "fill_blank"
}
