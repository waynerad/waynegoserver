package zorpcore

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
	"strings"
	"zorptypes"
	"zorpui"
)

//	"time"
//	"wutil"

//	"accessdb"
//	"io"
//	"math"
//	"math/rand"

type shareEntryData struct {
	idShare     uint64
	name        string
	description string
}

//	path string
//	readWriteKey string
//	readOnlySey string

func uint64ToStr(z uint64) string {
	return strconv.FormatUint(z, 10)
}

func candeleteint64ToStr(z int64) string {
	return strconv.FormatInt(z, 10)
}

func candeleteintToStr(z int) string {
	return strconv.FormatInt(int64(z), 10)
}

func candeletefloatToStr(z float64) string {
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

func candeletestrToInt64(z string) int64 {
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

func candeletecalculateDayNumber(timeCode uint64, timeZoneOffset int64, cycleDays int) (uint64, uint64) {
	// ourConstant := 270000 // this magic constant a) makes the day division 3:00am, and b) makes the 7-day cycles line up with human weeks
	ourConstant := 334800 // this magic constant a) makes the day division 3:00am, and b) makes the 7-day cycles line up with human weeks
	adjTime := uint64((int64(timeCode) + timeZoneOffset) + int64(ourConstant))
	adjDay := adjTime / uint64(cycleDays*86400) // note: integer division!
	boundary := (adjDay + 1) * uint64(cycleDays*86400)
	timeRemaining := boundary - adjTime
	return adjDay, timeRemaining
}

////////////////////////////////////////////////////////////////
// New Form object system
////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------
// Zorp Share List page
// ----------------------------------------------------------------

func getZorpShareListDBData(db mysql.Conn, userInfo *login.UserInformationRecord) zorptypes.ShareListData {
	var currentEntry zorptypes.ShareDisplayData
	theList := make(zorptypes.ShareListData, 0)

	sql := "SELECT id_share, name, description FROM zorp_share WHERE (id_user = ?) ORDER BY id_share;"
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
		currentEntry.IdShare = row.Uint64(0)
		currentEntry.Name = row.Str(1)
		currentEntry.Description = row.Str(2)
		theList = append(theList, currentEntry)
	}
	return theList
}

func showShareListPage(w http.ResponseWriter, userInfo *login.UserInformationRecord, userInput map[string]string, dbDataList zorptypes.ShareListData) {
	_, editmode := userInput["edit"]
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Zorp Share List"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	zorpui.ShowZorpHeadHeader(w, displayInfo)
	zorpui.ShowZorpBodyHeader(w, displayInfo)
	zorpui.ShowZorpShareList(w, dbDataList, editmode)
	zorpui.ShowZorpFooter(w, displayInfo)
}

type shareListForm struct {
	objectName string
}

func (self *shareListForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	_, editmode := userInput["edit"]
	if editmode {
		rv := make(map[string]string)
		rv["edit"] = "1"
		return rv
	}
	return nil
}

func (self *shareListForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	dbDataList := getZorpShareListDBData(db, userInfo)
	showShareListPage(w, userInfo, userInput, dbDataList)
}

func (self *shareListForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	return nil, nil
}

func (self *shareListForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	return nil
}

// ----------------------------------------------------------------
// Zorp Share Edit page
// ----------------------------------------------------------------

type shareEditForm struct {
	objectName string
}

func (self *shareEditForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	var currentEntry shareEntryData
	rv := make(map[string]string)
	sql := "SELECT id_share, name, description FROM zorp_share WHERE (id_share = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	shareid := strToUint64(userInput["share"])
	sel.Bind(shareid, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	found := false
	for _, row := range rows {
		currentEntry.idShare = row.Uint64(0)
		currentEntry.name = row.Str(1)
		currentEntry.description = row.Str(2)
		rv["share"] = uint64ToStr(currentEntry.idShare)
		rv["name"] = currentEntry.name
		rv["description"] = currentEntry.description
		found = true
	}
	if !found {
		rv["share"] = uint64ToStr(shareid)
		rv["name"] = ""
		rv["description"] = ""
	}
	return rv
}

func (self *shareEditForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Edit Zorp Share"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	zorpui.ShowZorpHeadHeader(w, displayInfo)
	zorpui.ShowZorpBodyHeader(w, displayInfo)
	zorpui.ShowZorpShareEditForm(w, errorList, userInput, displayInfo)
	zorpui.ShowZorpFooter(w, displayInfo)
}

func (self *shareEditForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
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

func (self *shareEditForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	var zorpShareRecord struct {
		shareId     uint64
		name        string
		description string
		cycleDays   int
	}
	zorpShareRecord.shareId = strToUint64(userInput["share"])
	zorpShareRecord.name = trim(userInput["name"])
	zorpShareRecord.description = trim(userInput["description"])
	zorpShareRecord.cycleDays = strToInt(userInput["cycle_days"])

	var err error
	alreadyExists := false
	if zorpShareRecord.shareId > 0 {
		sql := "SELECT id_share FROM zorp_share WHERE (id_share = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(zorpShareRecord.shareId, userInfo.UserId)
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
		stmt, err := db.Prepare("UPDATE zorp_share SET name = ?, description = ?, cycle_days = ?  WHERE (id_share = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(zorpShareRecord.name, zorpShareRecord.description, zorpShareRecord.cycleDays, zorpShareRecord.shareId, userInfo.UserId)
		_, _, err = stmt.Exec()
	} else {
		stmt, err := db.Prepare("INSERT INTO zorp_share (id_user, name, description, cycle_days) VALUES (?, ?, ?, ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		stmt.Bind(userInfo.UserId, zorpShareRecord.name, zorpShareRecord.description, zorpShareRecord.cycleDays)
		_, _, err = stmt.Exec()
	}
	if err != nil {
		fmt.Println(err)
		panic("Exec failed")
	}
	return nil
}

// ----------------------------------------------------------------
// End of zorp share edit page
// ----------------------------------------------------------------

func doShareListPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject shareListForm
	formObject.objectName = "Share List Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "http://www.yahoo.com/")
}

func doShareEditPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject shareEditForm
	formObject.objectName = "Share Edit Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "sharelist")
}

// ----------------------------------------------------------------
// End of zorp time check page
// ----------------------------------------------------------------

func Handler(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	fmt.Println("op is", op)
	switch {
	case op == "sharelist":
		if userInfo.UserId != 0 {
			doShareListPage(w, r, op, userInfo)
		}
	case op == "shareedit":
		if userInfo.UserId != 0 {
			doShareEditPage(w, r, op, userInfo)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/zorp/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
