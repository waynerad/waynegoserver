package streaklogic

import (
	"fmt"
	"forms"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"login"
	"net/http"
	"static"
	"strconv"
	"streak"
	"streakui"
	"strings"
	"time"
	"wutil"
)

//	"accessdb"
//	"io"
//	"math"
//	"math/rand"

type taskEntryData struct {
	IdTask      uint64
	Name        string
	Description string
	CycleDays   int
}

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
}

func getStyle() string {
	return `
<style>
body {
	font-family: "Trebuchet MS", "Helvetica", "Arial",  "Verdana", "sans-serif";
	font-size: 80%;
}
</style>
`
}

func uint64ToStr(z uint64) string {
	return strconv.FormatUint(z, 10)
}

func int64ToStr(z int64) string {
	return strconv.FormatInt(z, 10)
}

func intToStr(z int) string {
	return strconv.FormatInt(int64(z), 10)
}

func floatToStr(z float64) string {
	return strconv.FormatFloat(z, 'g', 5, 64)
}

func htmlize(z string) string {
	return html.EscapeString(z)
}

func strToUint64(z string) uint64 {
	rv, err := strconv.ParseUint(z, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("Parse Uint failed")
	}
	return rv
}

func strToInt64(z string) int64 {
	rv, err := strconv.ParseInt(z, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("Parse Int failed")
	}
	return rv
}

func strToInt(z string) int {
	rv, err := strconv.ParseInt(z, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("Parse Int failed")
	}
	return int(rv)
}

func trim(z string) string {
	return strings.Trim(z, " \r\n\t")
}

func getTimeZoneOffset(db mysql.Conn, userid uint64) int64 {
	var timeZoneOffset int64
	sql := "SELECT time_zone_offset FROM login_user WHERE id_user = " + strconv.FormatUint(userid, 10) + ";"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		timeZoneOffset = row.Int64(0)
	}
	return timeZoneOffset
}

func calculateDayNumber(timeCode uint64, timeZoneOffset int64, cycleDays int) (uint64, uint64) {
	// ourConstant := 270000 // this magic constant a) makes the day division 3:00am, and b) makes the 7-day cycles line up with human weeks
	ourConstant := 334800 // this magic constant a) makes the day division 3:00am, and b) makes the 7-day cycles line up with human weeks
	adjTime := uint64((int64(timeCode) + timeZoneOffset) + int64(ourConstant))
	adjDay := adjTime / uint64(cycleDays*86400) // note: integer division!
	boundary := (adjDay + 1) * uint64(cycleDays*86400)
	timeRemaining := boundary - adjTime
	return adjDay, timeRemaining
}

// returns length of current streak AND time remaining (though time remaining is not converted into a human-readable form)
func calculateCurrentStreakLen(db mysql.Conn, idTask uint64, cycleDays int, currentTime uint64, timeZoneOffset int64) (int, uint64) {
	today, timeRemaining := calculateDayNumber(currentTime, timeZoneOffset, cycleDays)
	limit := today - 1 // start from yesterday
	sql := "SELECT day_num FROM streak_day WHERE (id_task = ?) ORDER BY day_num DESC;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(idTask)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var dayNum uint64
	streakCount := 0
	keepGoing := true
	for _, row := range rows {
		if keepGoing { // we need to find a way to break completely out of this loop w/out reading whole result set
			dayNum = row.Uint64(0)
			if dayNum == today {
				timeRemaining = 0
			}
			if dayNum >= limit {
				streakCount++
				limit = dayNum - 1
			} else {
				keepGoing = false
			}
		}
	}
	return streakCount, timeRemaining
}

func convertTimeRemainingToEnglish(interval uint64) string {
	if interval <= 0 {
		return "Done!"
	}
	return wutil.TimeIntervalSecondsToEnglish(int(interval))
}

func recalculateDayNumbers(db mysql.Conn, userid uint64, taskid uint64, timeZoneOffset int64, cycleDays int) {
	type updateInfo struct {
		idDayTask uint64
		dayNum    uint64
	}
	var updateEntry updateInfo
	var updateList []updateInfo

	sql := "SELECT id_day_task, actual_time_gmt FROM streak_day WHERE (id_task = ?) ORDER BY day_num;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(taskid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var actualTimeGmt uint64
	for _, row := range rows {
		updateEntry.idDayTask = row.Uint64(0)
		actualTimeGmt = row.Uint64(1)
		updateEntry.dayNum, _ = calculateDayNumber(actualTimeGmt, timeZoneOffset, cycleDays)
		updateList = append(updateList, updateEntry)
	}
	for _, updateItem := range updateList {
		stmt, err := db.Prepare("UPDATE streak_day SET day_num = ?  WHERE (id_day_task = ?) AND (id_task = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(updateItem.dayNum, updateItem.idDayTask, taskid, userid)
		_, _, err = stmt.Exec()
	}
}

////////////////////////////////////////////////////////////////
// New Form object system
////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------
// Streak Task List page
// ----------------------------------------------------------------

func getStreakTaskListDBData(db mysql.Conn, userInfo *login.UserInformationRecord) streak.TaskListData {
	// mysql> DESCRIBE streak_task;
	// +---------+---------------------+------+-----+---------+----------------+
	// | Field   | Type                | Null | Key | Default | Extra          |
	// +---------+---------------------+------+-----+---------+----------------+
	// | id_task | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
	// | name    | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry streak.TaskDisplayData
	theList := make(streak.TaskListData, 0)
	currentTime := uint64(time.Now().Unix())
	timeZoneOffset := getTimeZoneOffset(db, userInfo.UserId)
	var interval uint64

	sql = "SELECT id_task, name, description, cycle_days FROM streak_task WHERE (id_user = ?) ORDER BY seq_num, id_task;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		currentEntry.IdTask = row.Uint64(0)
		currentEntry.Name = row.Str(1)
		currentEntry.Description = row.Str(2)
		currentEntry.CycleDays = row.Int(3)
		currentEntry.CurrentStreakLen, interval = calculateCurrentStreakLen(db, currentEntry.IdTask, currentEntry.CycleDays, currentTime, timeZoneOffset)
		currentEntry.TimeRemaining = convertTimeRemainingToEnglish(interval)
		theList = append(theList, currentEntry)
	}
	return theList
}

type taskListForm struct {
	objectName string
}

func (self *taskListForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	_, editmode := userInput["edit"]
	if editmode {
		rv := make(map[string]string)
		rv["edit"] = "1"
		return rv
	}
	return nil
}

func (self *taskListForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	dbDataList := getStreakTaskListDBData(db, userInfo)
	_, editmode := userInput["edit"]
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Streak Task List"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	streakui.ShowHeadHeader(w, displayInfo)
	streakui.ShowBodyHeader(w, displayInfo)
	streakui.ShowStreakTaskList(w, dbDataList, editmode)
	streakui.ShowFooter(w, displayInfo)
}

func (self *taskListForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	return nil, nil
}

func (self *taskListForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	return nil
}

// ----------------------------------------------------------------
// Streak Task Edit page
// ----------------------------------------------------------------

type taskEditForm struct {
	objectName string
}

func (self *taskEditForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	var currentEntry taskEntryData
	rv := make(map[string]string)
	sql := "SELECT id_task, name, description, cycle_days FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	taskid := strToUint64(userInput["task"])
	sel.Bind(taskid, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	found := false
	for _, row := range rows {
		currentEntry.IdTask = row.Uint64(0)
		currentEntry.Name = row.Str(1)
		currentEntry.Description = row.Str(2)
		currentEntry.CycleDays = row.Int(3)
		rv["task"] = uint64ToStr(currentEntry.IdTask)
		rv["name"] = currentEntry.Name
		rv["description"] = currentEntry.Description
		rv["cycle_days"] = intToStr(currentEntry.CycleDays)
		found = true
	}
	if !found {
		rv["task"] = uint64ToStr(taskid)
		rv["name"] = ""
		rv["description"] = ""
		rv["cycle_days"] = "1"
	}
	return rv
}

func (self *taskEditForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Edit Streak Task"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	streakui.ShowHeadHeader(w, displayInfo)
	streakui.ShowBodyHeader(w, displayInfo)
	streakui.ShowTaskEditForm(w, errorList, userInput, displayInfo)
	streakui.ShowFooter(w, displayInfo)
}

func (self *taskEditForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	fmt.Println("CheckForErrors: userInput", userInput)
	errorList := make(map[string]string)
	if userInput["name"] == "" {
		errorList["name"] = "Please specify a name."
	}
	if userInput["description"] == "" {
		errorList["description"] = "Please specify a description."
	}
	cycleDaysNum, err := strconv.ParseUint(userInput["cycle_days"], 10, 64)
	if err != nil {
		errorList["cycle_days"] = "Cycle days must be a number."
	} else {
		if cycleDaysNum == 0 {
			errorList["cycle_days"] = "Cycle days must not be zero."
		}
	}
	return errorList, nil
}

func (self *taskEditForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	fmt.Println("SaveForm: userInput", userInput)
	var streakTaskRecord struct {
		taskId      uint64
		name        string
		description string
		cycleDays   int
	}
	streakTaskRecord.taskId = strToUint64(userInput["task"])
	streakTaskRecord.name = trim(userInput["name"])
	streakTaskRecord.description = trim(userInput["description"])
	streakTaskRecord.cycleDays = strToInt(userInput["cycle_days"])

	var err error
	alreadyExists := false
	if streakTaskRecord.taskId > 0 {
		sql := "SELECT id_task FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(streakTaskRecord.taskId, userInfo.UserId)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Exec() failed")
		}
		for _, _ = range rows {
			alreadyExists = true
		}
	}
	if alreadyExists {
		stmt, err := db.Prepare("UPDATE streak_task SET name = ?, description = ?, cycle_days = ?  WHERE (id_task = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(streakTaskRecord.name, streakTaskRecord.description, streakTaskRecord.cycleDays, streakTaskRecord.taskId, userInfo.UserId)
		_, _, err = stmt.Exec()
		timeZoneOffset := getTimeZoneOffset(db, userInfo.UserId)
		recalculateDayNumbers(db, userInfo.UserId, streakTaskRecord.taskId, timeZoneOffset, streakTaskRecord.cycleDays)
	} else {
		stmt, err := db.Prepare("INSERT INTO streak_task (id_user, name, description, cycle_days) VALUES (?, ?, ?, ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		stmt.Bind(userInfo.UserId, streakTaskRecord.name, streakTaskRecord.description, streakTaskRecord.cycleDays)
		_, _, err = stmt.Exec()
	}
	if err != nil {
		fmt.Println(err)
		panic("Exec failed")
	}
	return nil
}

// ----------------------------------------------------------------
// End of streak task edit page
// ----------------------------------------------------------------

// ----------------------------------------------------------------
// Streak Time Check page
// ----------------------------------------------------------------

type markDoneForm struct {
	objectName string
}

func (self *markDoneForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	rv := make(map[string]string)
	taskid := strToUint64(userInput["task"])
	rv["task"] = uint64ToStr(taskid)
	return rv
}

func (self *markDoneForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Mark Done"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"

	taskId := strToUint64(userInput["task"])
	currentTime := uint64(time.Now().Unix())
	timeZoneOffset := getTimeZoneOffset(db, userInfo.UserId)
	var currentEntry streak.TaskDisplayData
	var interval uint64

	sql := "SELECT id_task, name, description, cycle_days FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(taskId, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		currentEntry.IdTask = row.Uint64(0)
		currentEntry.Name = row.Str(1)
		currentEntry.Description = row.Str(2)
		currentEntry.CycleDays = row.Int(3)
		currentEntry.CurrentStreakLen, interval = calculateCurrentStreakLen(db, currentEntry.IdTask, currentEntry.CycleDays, currentTime, timeZoneOffset)
		currentEntry.TimeRemaining = convertTimeRemainingToEnglish(interval)
	}

	streakui.ShowHeadHeader(w, displayInfo)
	streakui.ShowBodyHeader(w, displayInfo)
	streakui.ShowMarkDoneForm(w, errorList, userInput, currentEntry)
	streakui.ShowFooter(w, displayInfo)
}

func (self *markDoneForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	fmt.Println("CheckForErrors: userInput", userInput)
	errorList := make(map[string]string)
	return errorList, nil
}

func (self *markDoneForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	// mysql> DESCRIBE streak_day;
	// +-----------------+---------------------+------+-----+---------+----------------+
	// | Field           | Type                | Null | Key | Default | Extra          |
	// +-----------------+---------------------+------+-----+---------+----------------+
	// | id_day_task     | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
	// | id_user         | bigint(20) unsigned | NO   |     | 0       |                |
	// | id_task         | bigint(20) unsigned | NO   | MUL | 0       |                |
	// | actual_time_gmt | bigint(20) unsigned | NO   |     | 0       |                |
	// | day_num         | bigint(20) unsigned | NO   |     | 0       |                |
	// +-----------------+---------------------+------+-----+---------+----------------+

	userid := userInfo.UserId
	timeZoneOffset := getTimeZoneOffset(db, userid)
	taskid := strToUint64(userInput["task"])

	taskExists := false
	cycleDays := 0
	sql := "SELECT id_task, cycle_days FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(taskid, userid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, row := range rows {
		taskExists = true
		cycleDays = row.Int(1)
	}
	if !taskExists {
		panic("task Id does not exist")
	}

	// var idDayTask uint64
	var streakDay struct {
		idUser        uint64
		idTask        uint64
		actualTimeGmt uint64
		dayNum        uint64
	}

	currentTime := uint64(time.Now().Unix())

	dayNum, _ := calculateDayNumber(currentTime, timeZoneOffset, cycleDays)

	streakDay.idUser = userid
	streakDay.idTask = taskid
	streakDay.actualTimeGmt = currentTime
	streakDay.dayNum = dayNum

	alreadyExists := false

	sql = "SELECT id_day_task FROM streak_day WHERE (id_user = ?) AND (id_task = ?) AND (day_num = ?);"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, taskid, dayNum)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, _ = range rows {
		alreadyExists = true
		// idDayTask = row.Uint64(0) // Would be used if we did an update, but for this func, update is a no-op!
	}

	fmt.Println("Save: streakDay", streakDay)
	if alreadyExists {
		// if it already exists, we do nothing! no-op
	} else {
		stmt, err := db.Prepare("INSERT INTO streak_day (id_user, id_task, actual_time_gmt, day_num) VALUES (?, ?, ?, ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		stmt.Bind(streakDay.idUser, streakDay.idTask, streakDay.actualTimeGmt, streakDay.dayNum)
		_, _, err = stmt.Exec()
	}
	if err != nil {
		fmt.Println(err)
		panic("Exec failed")
	}
	return nil
}

// ----------------------------------------------------------------
// End of streak time check page
// ----------------------------------------------------------------

func showTaskListPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject taskListForm
	formObject.objectName = "Task List Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "http://www.yahoo.com/")
}

func showTaskEditPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject taskEditForm
	formObject.objectName = "Task Edit Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "tasklist")
}

func showMarkDonePage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject markDoneForm
	formObject.objectName = "Mark Done Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "tasklist")
}

// func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
func Handler(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	fmt.Println("op is", op)
	switch {
	case op == "tasklist":
		if userInfo.UserId != 0 {
			showTaskListPage(w, r, op, userInfo)
		}
	case op == "taskedit":
		if userInfo.UserId != 0 {
			showTaskEditPage(w, r, op, userInfo)
		}
	case op == "markdone":
		if userInfo.UserId != 0 {
			showMarkDonePage(w, r, op, userInfo)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/streak/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
