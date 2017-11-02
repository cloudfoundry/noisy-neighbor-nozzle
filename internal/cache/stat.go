package cache

// Stat stores counter data
type Stat struct {
	ID    string `json:"id"`
	Count int64  `json:"count"`
}

// Stats is a collection of Stat for sorting purposes.
type Stats []Stat

func (s Stats) Len() int           { return len(s) }
func (s Stats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Stats) Less(i, j int) bool { return s[i].Count > s[j].Count }
