package session

type Item struct {
	WordID        string
	ReviewType    string
	PriorityScore float64
	Reason        string
}

type Session struct {
	UserID string
	Items  []Item
	Index  int
}
