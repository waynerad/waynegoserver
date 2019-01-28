package wikitypes

type WikiPageDisplayData struct {
	IdPage  uint64
	Title   string
	Content string
}

type WikiPageListData []WikiPageDisplayData
