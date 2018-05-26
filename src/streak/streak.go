package streak

type TaskEntryData struct {
	IdTask      uint64
	Name        string
 	Description string
 	CycleDays   int
}

type TaskListData []TaskEntryData

