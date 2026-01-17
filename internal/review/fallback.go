package review

func fallback(t string) string {
	switch t {
	case "fill_blank":
		return "typing"
	case "typing":
		return "match"
	case "match":
		return "mcq"
	default:
		return "mcq"
	}
}
