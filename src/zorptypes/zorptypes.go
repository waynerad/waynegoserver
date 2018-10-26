package zorptypes

type ShareDisplayData struct {
	IdShare     uint64
	Name        string
	Description string
}

type ShareListData []ShareDisplayData
