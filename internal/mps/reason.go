package mps

func generateReason(timeF, accF, confF, failF float64) string {
	switch {
	case accF > 0.5:
		return "You often answer this incorrectly"
	case failF > 0.4:
		return "You recently made mistakes with this word"
	case timeF > 0.7:
		return "You haven’t reviewed this word recently"
	case confF > 0.5:
		return "You marked this word as low confidence"
	default:
		return "This word needs a quick refresh"
	}
}
