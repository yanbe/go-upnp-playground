package soap

type Action struct{}

func (a Action) Browse(ObjectID string, BrowseFlag string, Filter string, StartingIndex int, RequestedCount int, SortCriteria string) (string, int, int, int) {
	// Result, NumberReturned, TotalMatches, UpdateID
	return "result", 1, 2, 3
}
