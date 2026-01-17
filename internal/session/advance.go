package session

func (s *Session) Current() *Item {
	if s.Index >= len(s.Items) {
		return nil
	}
	return &s.Items[s.Index]
}

func (s *Session) Advance() {
	if s.Index < len(s.Items) {
		s.Index++
	}
}

func (s *Session) Done() bool {
	return s.Index >= len(s.Items)
}
