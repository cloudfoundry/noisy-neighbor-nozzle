package store

// Rate stores data for a single polling interval.
type Rate struct {
	Timestamp int64             `json:"timestamp"`
	Counts    map[string]uint64 `json:"counts"`
}

// Rates is a collection of Rate for sorting on timestamp and presentation purposes
type Rates []Rate

func (r Rates) Len() int           { return len(r) }
func (r Rates) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Rates) Less(i, j int) bool { return r[i].Timestamp < r[j].Timestamp }
