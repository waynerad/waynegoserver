package streak

type TaskDisplayData struct {
	IdTask           uint64
	Name             string
	Description      string
	CycleDays        int
	CurrentStreakLen int
	TimeRemaining    string
}

type TaskListData []TaskDisplayData
