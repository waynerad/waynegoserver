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
)

//	"accessdb"
//	"io"
//	"math"
//	"math/rand"
//	"time"

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

////////////////////////////////////////////////////////////////
// New Form object system
////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------
// Streak Task List page
// ----------------------------------------------------------------

func showStreakTaskList(w http.ResponseWriter, userInfo *login.UserInformationRecord, dbDataList streak.TaskListData) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Streak Task List"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	streakui.ShowHeadHeader(w, displayInfo)
	streakui.ShowBodyHeader(w, displayInfo)
	streakui.ShowStreakTaskList(w, dbDataList)
	streakui.ShowFooter(w, displayInfo)
}

func getStreakTaskListDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) streak.TaskListData {
	// mysql> DESCRIBE streak_task;
	// +---------+---------------------+------+-----+---------+----------------+
	// | Field   | Type                | Null | Key | Default | Extra          |
	// +---------+---------------------+------+-----+---------+----------------+
	// | id_task | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
	// | name    | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry streak.TaskEntryData
	theList := make(streak.TaskListData, 0)
	sql = "SELECT id_task, name, description, cycle_days FROM streak_task WHERE (id_user = ?) ORDER BY id_task;"
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
		theList = append(theList, currentEntry)
	}
	return theList
}

type taskListForm struct {
	objectName string
}

func (self *taskListForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	return nil
}

func (self *taskListForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	dbDataList := getStreakTaskListDBData(db, userInfo, userInput)
	showStreakTaskList(w , userInfo, dbDataList)
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

func showTaskEditForm(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Edit Streak Task"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	streakui.ShowHeadHeader(w, displayInfo)
	streakui.ShowBodyHeader(w, displayInfo)
	streakui.ShowTaskEditForm(w, errorList, userInput, displayInfo)
	streakui.ShowFooter(w, displayInfo)
}

type taskEditForm struct {
	objectName string
}

func (self *taskEditForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	var currentEntry streak.TaskEntryData
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
	showTaskEditForm(w, r, op, userInfo, errorList, userInput)
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
	fmt.Println("Save: streakTaskRecord", streakTaskRecord)
	if alreadyExists {
		stmt, err := db.Prepare("UPDATE streak_task SET name = ?, description = ?, cycle_days = ?  WHERE (id_task = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(streakTaskRecord.name, streakTaskRecord.description, streakTaskRecord.cycleDays, streakTaskRecord.taskId, userInfo.UserId)
		_, _, err = stmt.Exec()
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
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/streak/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}


