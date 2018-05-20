package forms

import (
	"accessdb"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"login"
	"net/http"
	"strconv"
	"strings"
)

func trim(z string) string {
	return strings.Trim(z, " \r\n\t")
}

////////////////////////////////////////////////////////////////
// New Form object system
////////////////////////////////////////////////////////////////

// this function was removed to simplify passing DB data without forcing everything to strings:
// getDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string
//
// showForm() changed to getDBDataAndShowForm()

type SimpleWebForm interface {
	GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string
	GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string)
	CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string)
	SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string
}

func GetUserInput(r *http.Request) map[string]string {
	rv := make(map[string]string)
	theForm := r.Form
	for name, valset := range theForm {
		if len(valset) == 1 {
			rv[name] = trim(valset[0])
		} else {
			lvals := len(valset)
			for i := 0; i < lvals; i++ {
				rv[name+"_"+strconv.FormatInt(int64(i), 10)] = trim(valset[i])
			}
		}
	}
	return rv
}

func HandleStandaloneForm(formObject SimpleWebForm, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, redirectTarget string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	alreadyProcessed := make(map[string]string)
	userInput := make(map[string]string)
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("parseform failed")
	}
	var dbConn mysql.Conn
	dbSet := false
	method := r.Method
	if method == "GET" {
		showform = true
		dbConn = accessdb.GetDbConnection()
		dbSet = true
		defer dbConn.Close()
		userInput = GetUserInput(r)
		userInput = formObject.GetDefaults(dbConn, userInfo, userInput)
	}
	if method == "POST" {
		if !dbSet {
			dbConn = accessdb.GetDbConnection()
			dbSet = true
		}
		userInput = GetUserInput(r)
		errorList, alreadyProcessed = formObject.CheckForErrors(dbConn, userInput)
		if len(errorList) > 0 {
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			_ = formObject.SaveForm(dbConn, userInfo, userInput, alreadyProcessed)
			http.Redirect(w, r, redirectTarget, 302)
		}
	}
	if showform {
		if !dbSet {
			dbConn = accessdb.GetDbConnection()
			dbSet = true
		}
		formObject.GetDBDataAndShowForm(dbConn, w, r, op, userInfo, errorList, userInput)
	}
}
