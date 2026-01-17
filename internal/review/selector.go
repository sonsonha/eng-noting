package review

func SelectType(ctx Context) string {
	var selected string

	switch {
	case ctx.TotalReviews == 0:
		selected = "mcq"

	case ctx.AccuracyRate < 0.4:
		selected = "mcq"

	case ctx.AccuracyRate < 0.7:
		selected = "match"

	case ctx.AccuracyRate > 0.8 && ctx.TotalReviews >= 5:
		selected = "fill_blank"

	default:
		selected = "typing"
	}

	if selected == ctx.LastReviewType {
		return fallback(selected)
	}

	return selected
}
