package review

func Reason(ctx Context, selected string) string {
	switch selected {
	case "mcq":
		if ctx.TotalReviews == 0 {
			return "This word is new — choose the correct meaning"
		}
		return "Let’s reinforce recognition before recall"

	case "match":
		return "You recognize this word — let’s strengthen understanding"

	case "typing":
		return "You know this word — recall it without hints"

	case "fill_blank":
		return "You’ve mastered this word — use it in context"

	default:
		return "Quick review"
	}
}
