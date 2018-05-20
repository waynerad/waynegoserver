package streak

import (
	"accessdb"
	"fmt"
	"forms"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"login"
	"net/http"
	"static"
	"strconv"
	"streakui"
	"strings"
)

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

// getDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string
// this function was removed to simplify passing DB data without forcing everything to strings
// showForm() changed to getDBDataAndShowForm()
//type simpleWebForm interface {
//	getDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string
//	getDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string)
//	checkForErrors() map[string]string
//	saveForm(db mysql.Conn, userInput map[string]string, alreadyProcessed map[string]string) map[string]string
// }

// ----------------------------------------------------------------
// Streak Task List page
// ----------------------------------------------------------------

type taskEntryData struct {
	idTask      uint64
	name        string
	description string
	cycleDays   int
}
type taskListData []taskEntryData

func showStreakTaskListForm(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string, dbDataList taskListData) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Streaks</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>List of Streaks</h1>
`)
	fmt.Fprint(w, "dbDataList", dbDataList)
	started := false
	count := 0
	for _, streakTaskXYZ := range dbDataList {
		if count == 0 {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Cycle Days </th><th> Current Streak Length </th><th> Done Today! </th><th> Edit </th></tr>
`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		streakAsStr := uint64ToStr(streakTaskXYZ.idTask)
		fmt.Fprint(w, "<tr "+backgroundColor+">")
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+htmlize(streakTaskXYZ.name)+" </td><td> "+`<a href="taskedit?task=`+streakAsStr+`">Edit</a> </td>
	`)
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
        <p> <a href="taskedit?task=0">Add Streak</a> </p>
  </section>
</body>
</html>`)
}

func getStreakTaskListDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) taskListData {
	// mysql> DESCRIBE streak_task;
	// +---------+---------------------+------+-----+---------+----------------+
	// | Field   | Type                | Null | Key | Default | Extra          |
	// +---------+---------------------+------+-----+---------+----------------+
	// | id_task | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
	// | name    | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry taskEntryData
	theList := make(taskListData, 0)
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
		currentEntry.idTask = row.Uint64(0)
		currentEntry.name = row.Str(1)
		currentEntry.description = row.Str(2)
		currentEntry.cycleDays = row.Int(3)
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
	showStreakTaskListForm(w, r, op, userInfo, errorList, userInput, dbDataList)
}

func (self *taskListForm) ShowForm(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string, dbDataList map[string]string) map[string]string {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Streaks</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>List of Streaks</h1>
`)
	fmt.Fprint(w, "dbDataList", dbDataList)
	started := false
	count := 0
	if true {
		fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Edit </th></tr>
`)
		started = true
	}
	backgroundColor := " style=\"background-color: #FFFFFF;\""
	if (count & 1) == 1 {
		backgroundColor = " style=\"background-color: #E8F0E8;\""
	}
	// streakAsStr := uint64ToStr(streak.idStreak)
	fmt.Fprint(w, "<tr "+backgroundColor+">")
	// fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+htmlize(streak.name)+" </td><td> "+`<a href="editstreak?streak=`+streakAsStr+`">Edit</a> </td>
	// `)
	fmt.Fprint(w, `</tr>
`)
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
        <p> <a href="taskedit?streak=0">Add Streak</a> </p>
  </section>
</body>
</html>`)
	return nil
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
	// mysql> DESCRIBE streak_task;
	// +---------+---------------------+------+-----+---------+----------------+
	// | Field   | Type                | Null | Key | Default | Extra          |
	// +---------+---------------------+------+-----+---------+----------------+
	// | id_task | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
	// | name    | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry taskEntryData
	rv := make(map[string]string)
	sql = "SELECT id_task, name, description, cycle_days FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
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
		currentEntry.idTask = row.Uint64(0)
		currentEntry.name = row.Str(1)
		currentEntry.description = row.Str(2)
		currentEntry.cycleDays = row.Int(3)
		rv["task"] = uint64ToStr(taskid)
		rv["name"] = currentEntry.name
		rv["description"] = currentEntry.description
		rv["cycle_days"] = intToStr(currentEntry.cycleDays)
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
	// getStreakTaskEditDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) void
	fmt.Println("GetDBDataAndShowForm userInput", userInput)
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

func oldShowTaskListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	var streak struct {
		idStreak uint64
		name     string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Streaks</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>List of Streaks</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	sql = "SELECT id_streak, name FROM streak_streak WHERE (id_user = ?) ORDER BY id_streak ;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	count := 0
	started := false
	for _, row := range rows {
		streak.idStreak = row.Uint64(0)
		streak.name = row.Str(1)
		if !started {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Edit </th></tr>
`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		streakAsStr := uint64ToStr(streak.idStreak)
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+htmlize(streak.name)+" </td><td> "+`<a href="editstreak?streak=`+streakAsStr+`">Edit</a> </td>
`)
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
        <p> <a href="taskedit?streak=0">Add Streak</a> </p>
  </section>
</body>
</html>`)
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
