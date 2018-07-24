package streak

type TaskDisplayData struct {
	IdTask           uint64
	Name             string
	Description      string
	CycleDays        int
	CurrentStreakLen int
	TimeRemaining    string
	ShowMarkDone     bool
}

type TaskListData []TaskDisplayData

type DayHistoryDisplayData struct {
	ActualTimeGmt uint64
	DayNum        uint64
	Consecutive   int
	Gap           int
}

type DayHistoryListData []DayHistoryDisplayData
