package blind

import (
	"errors"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"math"
	"net/http"
	"static"
	"strconv"
	"strings"
	"time"
)

type timenumbers struct {
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int
}

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
}

func getDbConnection() (mysql.Conn, error) {
	user := "webdata_user"
	pass := "97abcmt3teteej"
	dbname := "webdata"
	db := mysql.New("tcp", "", "127.0.0.1:3306", user, pass, dbname)
	err := db.Connect()
	return db, err
}

func checkIsNumeric(str string) bool {
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c < 48 {
			return false
		}
		if c > 57 {
			return false
		}
	}
	return true
}

func checkTimeFieldSyntaxError(contents string, allowSlash bool) bool {
	if contents == "" {
		return true
	}
	if contents == "*" {
		return false
	}
	if checkIsNumeric(contents) {
		return false
	}
	if !allowSlash {
		return true
	}
	i := strings.Index(contents, "/")
	if i == -1 {
		return true
	}
	if contents[0:i] == "*" {
		if contents[i+1:] == "" {
			return true
		}
		if checkIsNumeric(contents[i+1:]) {
			return false
		}
	}
	return true
}

func testCheckTimeFieldSyntaxError() {
	if !checkTimeFieldSyntaxError("", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on empty string.")
	}
	if !checkTimeFieldSyntaxError("A", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on single letter.")
	}
	if !checkTimeFieldSyntaxError("word", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on letters.")
	}
	if !checkTimeFieldSyntaxError("5A", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on number and letter.")
	}
	if checkTimeFieldSyntaxError("5", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on single-digit number.")
	}
	if checkTimeFieldSyntaxError("55", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on double digit number.")
	}
	if checkTimeFieldSyntaxError("*", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk.")
	}
	if !checkTimeFieldSyntaxError("**", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on double asterisk.")
	}
	if !checkTimeFieldSyntaxError("*5", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk and number.")
	}
	if !checkTimeFieldSyntaxError("*A", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk and letter.")
	}
	if !checkTimeFieldSyntaxError("*/", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash.")
	}
	if !checkTimeFieldSyntaxError("*/A", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash letter.")
	}
	if checkTimeFieldSyntaxError("*/5", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash digit.")
	}
	if checkTimeFieldSyntaxError("*/6", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash double digit.")
	}
	if checkTimeFieldSyntaxError("*/888", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash triple digit.")
	}
	if checkTimeFieldSyntaxError("*/1234569", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash 8-digit.")
	}
	if !checkTimeFieldSyntaxError(" *", true) {
		fmt.Println("checkTimeFieldSyntaxError test failed on leading space.")
	}
	if !checkTimeFieldSyntaxError("*/", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash.")
	}
	if !checkTimeFieldSyntaxError("*/A", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash letter.")
	}
	if !checkTimeFieldSyntaxError("*/3", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash digit + disallow slash.")
	}
	if !checkTimeFieldSyntaxError("*/22", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash double digit + disallow slash.")
	}
	if !checkTimeFieldSyntaxError("*/444", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash triple digit + disallow slash.")
	}
	if !checkTimeFieldSyntaxError(" *", false) {
		fmt.Println("checkTimeFieldSyntaxError test failed on leading space + disallow slash.")
	}
}

func timeCodeToTimeNumbers(dbtime uint64) timenumbers {
	to := time.Unix(int64(dbtime), 0)
	year, month, day := to.Date()
	hour, min, sec := to.Clock()
	var rv timenumbers
	rv.year = year
	rv.month = int(month)
	rv.day = day
	rv.hour = hour
	rv.minute = min
	rv.second = sec
	return rv
}

func timeNumbersToTimeCode(timeNums timenumbers) uint64 {
	return uint64(time.Date(timeNums.year, time.Month(timeNums.month), timeNums.day, timeNums.hour, timeNums.minute, timeNums.second, 0, time.UTC).Unix())
}

func twodigits(num int) string {
	if num < 10 {
		return "0" + strconv.FormatInt(int64(num), 10)
	} else {
		return strconv.FormatInt(int64(num), 10)
	}
}

func timeCodeToString(dbtime uint64, timeZoneOffset int64) string {
	tc := timeCodeToTimeNumbers(uint64(int64(dbtime) + timeZoneOffset))
	return strconv.FormatInt(int64(tc.year), 10) + "-" + twodigits(tc.month) + "-" + twodigits(tc.day) + " " + twodigits(tc.hour) + ":" + twodigits(tc.minute) + ":" + twodigits(tc.second)
}

func applyRule(currentValue int, rule string, fieldLimit int, fieldMin int, currentCarry int) (int, int) {
	if fieldLimit == 0 {
		panic("fieldLimit is zero.")
	}
	newCarry := 0
	if rule == "*" {
		currentValue = currentValue + currentCarry
		for currentValue >= (fieldLimit + fieldMin) {
			newCarry++
			currentValue = currentValue - fieldLimit
		}
		return currentValue, newCarry
	}
	if len(rule) >= 2 {
		if rule[0:2] == "*/" {
			denomStr := rule[2:]
			denomVal, err := strconv.ParseInt(denomStr, 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseInt failed")
			}
			newCarry = 0
			currentValue = currentValue + currentCarry
			for (currentValue % int(denomVal)) != 0 {
				currentValue++
				for currentValue >= (fieldLimit + fieldMin) {
					newCarry++
					currentValue = currentValue - fieldLimit
				}
			}
			return currentValue, newCarry
		}
	}
	value, err := strconv.ParseInt(rule, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("ParseInt failed")
	}
	currentValue = currentValue + currentCarry
	for currentValue >= (fieldLimit + fieldMin) {
		newCarry++
		fmt.Println("229 newCarry", newCarry)
		currentValue = currentValue - fieldLimit
		fmt.Println("231 currentValue", currentValue)
		fmt.Println("   fieldLimit + fieldMin", fieldLimit+fieldMin)
	}
	if currentValue <= int(value) {
		return int(value), newCarry
	}
	return int(value), newCarry + 1
}

func testApplyRule() {
	newval, carry := applyRule(0, "*", 60, 0, 0)
	if (newval != 0) || (carry != 0) {
		fmt.Println("test 1 failed for applyRule")
	}
	newval, carry = applyRule(0, "*", 60, 0, 1)
	if (newval != 1) || (carry != 0) {
		fmt.Println("test 2 failed for applyRule")
	}
	newval, carry = applyRule(59, "*", 60, 0, 1)
	if (newval != 0) || (carry != 1) {
		fmt.Println("test 3 failed for applyRule")
	}
	newval, carry = applyRule(99, "*", 60, 0, 1)
	if (newval != 40) || (carry != 1) {
		fmt.Println("test 4 failed for applyRule")
	}
	newval, carry = applyRule(59, "*", 60, 1, 1)
	if (newval != 60) || (carry != 0) {
		fmt.Println("test 5 failed for applyRule")
	}
	newval, carry = applyRule(1, "*/2", 60, 0, 0)
	if (newval != 2) || (carry != 0) {
		fmt.Println("test 6 failed for applyRule")
	}
	newval, carry = applyRule(1, "*/3", 60, 0, 0)
	if (newval != 3) || (carry != 0) {
		fmt.Println("test 7 failed for applyRule")
	}
	newval, carry = applyRule(0, "*/3", 60, 0, 1)
	if (newval != 3) || (carry != 0) {
		fmt.Println("test 8 failed for applyRule")
	}
	newval, carry = applyRule(0, "30", 60, 0, 0)
	if (newval != 30) || (carry != 0) {
		fmt.Println("test 9 failed for applyRule")
	}
	newval, carry = applyRule(0, "30", 60, 0, 1)
	if (newval != 30) || (carry != 0) {
		fmt.Println("test 10 failed for applyRule")
	}
	newval, carry = applyRule(45, "30", 60, 0, 0)
	if (newval != 30) || (carry != 1) {
		fmt.Println("test 11 failed for applyRule")
	}
	newval, carry = applyRule(45, "30", 60, 0, 1)
	if (newval != 30) || (carry != 1) {
		fmt.Println("test 12 failed for applyRule")
		fmt.Println("newval", newval)
		fmt.Println("carry", carry)
	}
}

func countOfDaysInThisMonth(timeCode time.Time) int {
	count := 0
	start := timeCode.Month()
	tc := timeCode
	check := tc.Month()
	for check == start {
		count++
		tc = tc.Add(-604800000000000) // 1 week
		check = tc.Month()
	}
	return count
}

func calculateNextCurrentTimeForEvent(currentTimeCode uint64, yearRule string, monthRule string, domRule string, dowRule string, nthRule string, hourRule string, minuteRule string, secondRule string) uint64 {
	timeNumbers := timeCodeToTimeNumbers(currentTimeCode)

	var result timenumbers

	// forward clearing pass

	fmt.Println("")
	fmt.Println("timeNumbers at beginning", timeNumbers)

	foundDiff := false
	newval, _ := applyRule(timeNumbers.year, yearRule, 2147483647, 0, 0)
	if newval > timeNumbers.year {
		fmt.Println("325 !=")
		timeNumbers.year = newval
		fmt.Println("327 timeNumbers.year", timeNumbers.year)
		foundDiff = true
		fmt.Println("329 foundDiff", foundDiff)
	}

	newval, _ = applyRule(timeNumbers.month, monthRule, 12, 1, 0)
	if foundDiff {
		fmt.Println("336 in foundDiff")
		timeNumbers.month = 1
		fmt.Println("338 timeNumbers.month", timeNumbers.month)
	} else {
		if newval > timeNumbers.month {
			timeNumbers.month = newval
			foundDiff = true
		}
	}

	monthlengs := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	daysinmonth := monthlengs[timeNumbers.month-1]
	newval, _ = applyRule(timeNumbers.day, domRule, daysinmonth, 1, 0)
	if foundDiff {
		timeNumbers.day = 1
	} else {
		if newval > timeNumbers.day {
			fmt.Println("365 newval", newval)
			timeNumbers.day = newval
			fmt.Println("367 timeNumbers.day", timeNumbers.day)
			foundDiff = true
			fmt.Println("369 foundDiff", foundDiff)
		}
	}

	newval, _ = applyRule(timeNumbers.hour, hourRule, 60, 0, 0)
	if foundDiff {
		timeNumbers.hour = 0
	} else {
		if newval > timeNumbers.hour {
			timeNumbers.hour = newval
			foundDiff = true
		}
	}

	newval, _ = applyRule(timeNumbers.minute, minuteRule, 60, 0, 0)
	if foundDiff {
		timeNumbers.minute = 0
	} else {
		fmt.Println("401 else")
		if newval > timeNumbers.minute {
			fmt.Println("403 !=")
			timeNumbers.minute = newval
			fmt.Println("405 timeNumbers.minute", timeNumbers.minute)
			foundDiff = true
			fmt.Println("407 foundDiff", foundDiff)
		}
		fmt.Println("409")
	}

	newval, _ = applyRule(timeNumbers.second, secondRule, 60, 0, 0)
	if foundDiff {
		timeNumbers.second = 0
	} else {
		fmt.Println("420 else")
		if newval > timeNumbers.second {
			fmt.Println("422 !=")
			timeNumbers.second = newval
			fmt.Println("424 timeNumbers.second", timeNumbers.second)
			foundDiff = true
			fmt.Println("426 foundDiff", foundDiff)
		}
		fmt.Println("428")
	}

	fmt.Println("timeNumbers after forward clearing", timeNumbers)

	// backwards add with carry

	newval, minuteCarry := applyRule(timeNumbers.second, secondRule, 60, 0, 0)
	result.second = newval

	newval, hourCarry := applyRule(timeNumbers.minute, minuteRule, 60, 0, minuteCarry)
	result.minute = newval

	newval, dayCarry := applyRule(timeNumbers.hour, hourRule, 60, 0, hourCarry)
	result.hour = newval

	// monthlengs := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 0, 30, 31}
	daysinmonth = monthlengs[timeNumbers.month-1]
	newval, monthCarry := applyRule(timeNumbers.day, domRule, daysinmonth, 1, dayCarry)
	result.day = newval

	newval, yearCarry := applyRule(timeNumbers.month, monthRule, 12, 1, monthCarry)
	result.month = newval

	newval, _ = applyRule(timeNumbers.year, yearRule, 2147483647, 0, yearCarry)
	result.year = newval

	fmt.Println("result after backwards add with carry", result)

	// move forward to weekday if weekday specificed

	if dowRule != "*" {
		dowNum, err := strconv.ParseUint(dowRule, 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
		offsetDays := 0
		dayTime := time.Date(result.year, time.Month(result.month), result.day, 12, 0, 0, 0, time.UTC)
		breakloop := 0
		for dayTime.Weekday() != time.Weekday(dowNum) {
			dayTime = dayTime.Add(86400000000000) // 1 day
			offsetDays++
			breakloop++
			if breakloop == 12 {
				panic("Stuck in endless loop.")
			}
		}
		// if nth weekday of month rule specified, move forward to the nth weekday
		if nthRule != "*" {
			nthNum, err := strconv.ParseUint(nthRule, 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			for countOfDaysInThisMonth(dayTime) != int(nthNum) {
				dayTime = dayTime.Add(604800000000000) // 1 week
				offsetDays = offsetDays + 7
			}
		}
		return timeNumbersToTimeCode(result) + uint64(offsetDays*86400)
		// if monthRule is fixed and not "*", then this can return the wrong month!!
	}

	return timeNumbersToTimeCode(result)
}

func getTimeZoneOffset(db mysql.Conn, userid uint64) int64 {
	var timeZoneOffset int64
	sql := "SELECT time_zone_offset FROM login_user WHERE id_user = " + strconv.FormatUint(userid, 10) + ";"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	// sel.Bind(entryid, userid) // nothing to bind for this one
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		timeZoneOffset = row.Int64(0)
		fmt.Println("timeZoneOffset", timeZoneOffset)
	}
	return timeZoneOffset
}

func showEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		title       string
		description string
		year        string
		month       string
		dom         string
		dow         string
		nth         string
		hour        string
		minute      string
		second      string
	}
	var entryid uint64
	entryid = 0
	if method == "GET" {
		// set defaults
		ui.title = ""
		ui.description = ""
		ui.year = "*"
		ui.month = "*"
		ui.dom = "*"
		ui.dow = "*"
		ui.nth = "*"
		ui.hour = "*"
		ui.minute = "0"
		ui.second = "0"
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, entry := getform["entry"]
		if entry {
			entryid, err = strconv.ParseUint(getform["entry"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db, err := getDbConnection()
			if err != nil {
				fmt.Println(err)
				panic("getDbConnection failed")
			}
			defer db.Close()
			sql := "SELECT title, description, year, month, dom, dow, nth, hour, minute, second FROM calcron_entry WHERE (id_cal_ent = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(entryid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.title = row.Str(0)
				ui.description = row.Str(1)
				ui.year = row.Str(2)
				ui.month = row.Str(3)
				ui.dom = row.Str(4)
				ui.dow = row.Str(5)
				ui.nth = row.Str(6)
				ui.hour = row.Str(7)
				ui.minute = row.Str(8)
				ui.second = row.Str(9)
			}
		}
	}
	if method == "POST" {
		fmt.Println("point 553")
		// set from form post
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		postform := r.Form
		// error checking
		entryid, err = strconv.ParseUint(postform["entry"][0], 10, 64)
		if err != nil {
			entryid = 0
		}
		ui.title = strings.Trim(postform["title"][0], " \r\n\t")
		ui.description = strings.Trim(postform["description"][0], " \r\n\t")
		ui.year = strings.Trim(postform["year"][0], " \r\n\t")
		ui.month = strings.Trim(postform["month"][0], " \r\n\t")
		ui.dom = strings.Trim(postform["dom"][0], " \r\n\t")
		ui.dow = strings.Trim(postform["dow"][0], " \r\n\t")
		ui.nth = strings.Trim(postform["nth"][0], " \r\n\t")
		ui.hour = strings.Trim(postform["hour"][0], " \r\n\t")
		ui.minute = strings.Trim(postform["minute"][0], " \r\n\t")
		ui.second = strings.Trim(postform["second"][0], " \r\n\t")
		if ui.title == "" {
			errorList["title"] = "Please specify a title."
		}
		if checkTimeFieldSyntaxError(ui.year, true) {
			errorList["year"] = "Year is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.month, true) {
			errorList["month"] = "Month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.dom, true) {
			errorList["dom"] = "Day of month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.dow, false) {
			errorList["dow"] = "Day of week is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.nth, false) {
			errorList["nth"] = "Nth day of week in month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.hour, true) {
			errorList["hour"] = "Hour is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.minute, true) {
			errorList["minute"] = "Minute is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.second, true) {
			errorList["second"] = "Second is invalid"
			errorOccurred = true
		}
		if ui.nth != "*" {
			if ui.dow == "*" {
				errorList["dow"] = "If you put something in the nth day of week in month field, you must specify a weekday."
				errorOccurred = true
			}
		}
		if ui.dow != "*" {
			if ui.dom != "*" {
				errorList["dow"] = "If you specify a day of week, you cannot specify a day of month."
				errorOccurred = true
			}
		}
		if errorOccurred {
			showform = true
		} else {
			fmt.Println("point 623")
			// dbConnect!!
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			var timeZoneOffset int64
			timeZoneOffset = getTimeZoneOffset(db, userid)
			fmt.Println("point 634")
			var save struct {
				idCalEnt    uint64
				idUser      uint64
				title       string
				description string
				starttime   uint64
				year        string
				month       string
				dom         string
				dow         string
				nth         string
				hour        string
				minute      string
				second      string
				currenttime uint64
			}
			save.idCalEnt = entryid
			save.idUser = userid
			save.title = ui.title
			save.description = ui.description
			save.starttime = uint64(time.Now().Unix())
			save.year = ui.year
			save.month = ui.month
			save.dom = ui.dom
			save.dow = ui.dow
			save.nth = ui.nth
			save.hour = ui.hour
			save.minute = ui.minute
			save.second = ui.second
			// save.currenttime = uint64(time.Now().Unix())
			// save.currenttime = uint64(int64(calculateNextCurrentTimeForEvent(uint64(int64(save.starttime) + timeZoneOffset), save.year, save.month, save.dom, save.dow, save.nth, save.hour, save.minute, save.second)) - timeZoneOffset)
			fmt.Println("point 669")
			save.currenttime = calculateNextCurrentTimeForEvent(uint64(int64(save.starttime)+timeZoneOffset), save.year, save.month, save.dom, save.dow, save.nth, save.hour, save.minute, save.second)
			fmt.Println("point 671")
			save.currenttime = uint64(int64(save.currenttime) - timeZoneOffset)
			// query, if there, update, if not, create new
			alreadyExists := false
			if entryid > 0 {
				sql := "SELECT id_cal_ent FROM calcron_entry WHERE (id_cal_ent = ?) AND (id_user = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(entryid, userid)
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
				stmt, err := db.Prepare("UPDATE calcron_entry SET title = ?, description = ?, year = ?, month = ?, dom = ?, dow = ?, nth = ?, hour = ?, minute = ?, second = ?, currenttime = ?, is_defunct = 0 WHERE (id_cal_ent = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.title, save.description, save.year, save.month, save.dom, save.dow, save.nth, save.hour, save.minute, save.second, save.currenttime, entryid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO calcron_entry (id_user, title, description, starttime, year, month, dom, dow, nth, hour, minute, second, currenttime, is_defunct) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idUser, save.title, save.description, save.starttime, save.year, save.month, save.dom, save.dow, save.nth, save.hour, save.minute, save.second, save.currenttime)
				_, _, err = stmt.Exec()
			}
			if err != nil {
				fmt.Println(err)
				panic("Exec failed")
			}
			http.Redirect(w, r, "list", 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer db.Close()
		fmt.Fprint(w, `<title>CalCron Entry</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<script>
/*jslint browser: true, devel: true */
// (C) 2013-2016 Wayne Radinsky
// jsabort, cx, and ctstr are debugging functions that will be REMOVED from the final product
function jsabort(msg) {
    "use strict";
    window.alert(msg);
    throw new Error('Aborting javascript: ' + msg);
}
function cx(stringparameter) {
    "use strict";
    // this function exists just so I don't have to type 'console.log' over and over (was originally just 'cx', the '' prefix made it longer, d'oh!), plus prevents output from changing!
    console.log(stringparameter);
}
function ctstr(anyparameter) {
    "use strict";
    // a supposedly simple function to make a string from a var for debug dumps, that has gotten way complicated over time...
    var x, rv, z, name, isArr;
    if (typeof anyparameter === 'object') {
        if (anyparameter === null) {
            return 'null';
        }
        name = '';
        isArr = false;
        // apc = anyparameter.constructor; / so we can examine in debugger
        if (anyparameter.constructor !== undefined) {
            name = anyparameter.constructor.toString();
        }
        if (name === 'function Object() { [native code] }') {
            name = '';
        }
        if (name === 'function Array() { [native code] }') {
            name = '';
            isArr = true;
        }
        if (name === 'function AudioContext() { [native code] }') {
            return 'object AudioContext';
        }
        if (name === '') {
            rv = '';
            for (z in anyparameter) {
                if (anyparameter.hasOwnProperty(z)) {
                    rv = rv + ', ' + z + ': ' + ctstr(anyparameter[z]);
                }
            }
            if (isArr) {
                return 'array [ ' + rv.substring(2) + ' ]';
            }
            return 'object { ' + rv.substring(2) + ' }';
        }
        return '{[Code Object]}';
    }
    if (typeof anyparameter === 'number') {
        return 'number ' + anyparameter.toString();
    }
    if (typeof anyparameter === 'string') {
        return 'string "' + anyparameter + '"';
    }
    if (typeof anyparameter === 'boolean') {
        return 'boolean ' + anyparameter.toString();
    }
    x = typeof anyparameter;
    window.alert(x);
    console.log('x', x);
    jsabort('typeof returned an undefined value');
}
// main
</script>
</head>
<body>
  <section>
    <h1>CalCron Entry</h1>
<form action="edit" method="post">
<input type="hidden" name="entry" value="`+strconv.FormatUint(entryid, 10)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<table border="0" cellpadding="4">
<tr><td align="right"> Title: </td><td> <input class="biginput" name="title" id="title" type="text" value="`+html.EscapeString(ui.title)+`" /> </td></tr>
<tr><td align="right"> Description: </td><td> <input class="biginput" name="description" id="description" type="text" value="`+html.EscapeString(ui.description)+`" /> </td></tr>
<tr><td align="right"> Year: </td><td> <input class="biginput" name="year" id="year" type="text" value="`+html.EscapeString(ui.year)+`" /> </td></tr>
<tr><td align="right"> Month: </td><td> <input class="biginput" name="month" id="month" type="text" value="`+html.EscapeString(ui.month)+`" /> </td></tr>
<tr><td align="right"> Day of month: </td><td> <input class="biginput" name="dom" id="dom" type="text" value="`+html.EscapeString(ui.dom)+`" /> </td></tr>
<tr><td colspan="2" align="center" style="border-bottom: 1px solid #000000;"> </td></tr>
<tr><td align="right"> Day of week: </td><td> <input class="biginput" name="dow" id="dow" type="text" value="`+html.EscapeString(ui.dow)+`" /> </td></tr>
<tr><td align="right"> Nth day of day of week this month: </td><td> <input class="biginput" name="nth" id="nth" type="text" value="`+html.EscapeString(ui.nth)+`" /> </td></tr>
<tr><td colspan="2" align="center" style="border-bottom: 1px solid #000000;"> </td></tr>
<tr><td align="right"> Hour: </td><td> <input class="biginput" name="hour" id="hour" type="text" value="`+html.EscapeString(ui.hour)+`" /> </td></tr>
<tr><td align="right"> Minute: </td><td> <input class="biginput" name="minute" id="minute" type="text" value="`+html.EscapeString(ui.minute)+`" /> </td></tr>
<tr><td align="right"> Second: </td><td> <input class="biginput" name="second" id="second" type="text" value="`+html.EscapeString(ui.second)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

// mysql> DESCRIBE blind_set;
// +-------------+---------------------+------+-----+---------+----------------+
// | Field       | Type                | Null | Key | Default | Extra          |
// +-------------+---------------------+------+-----+---------+----------------+
// | id_set      | int(10) unsigned    | NO   | PRI | NULL    | auto_increment |
// | id_user     | bigint(20) unsigned | NO   |     | 0       |                |
// | id_group    | int(10) unsigned    | NO   |     | 0       |                |
// | created_gmt | bigint(20) unsigned | NO   | MUL | 0       |                |
// | name        | varchar(255)        | NO   |     |         |                |
// | pop_fact    | float               | NO   |     | 0       |                |
// +-------------+---------------------+------+-----+---------+----------------+
// 6 rows in set (0.00 sec)
// mysql> DESCRIBE blind_book;
// +---------+---------------------+------+-----+---------+----------------+
// | Field   | Type                | Null | Key | Default | Extra          |
// +---------+---------------------+------+-----+---------+----------------+
// | id_book | int(10) unsigned    | NO   | PRI | NULL    | auto_increment |
// | id_user | bigint(20) unsigned | NO   |     | 0       |                |
// | id_set  | int(10) unsigned    | NO   | MUL | 0       |                |
// | title   | varchar(255)        | NO   |     |         |                |
// | authors | varchar(255)        | NO   |     |         |                |
// | stars5  | int(10) unsigned    | NO   |     | 0       |                |
// | stars4  | int(10) unsigned    | NO   |     | 0       |                |
// | stars3  | int(10) unsigned    | NO   |     | 0       |                |
// | stars2  | int(10) unsigned    | NO   |     | 0       |                |
// | stars1  | int(10) unsigned    | NO   |     | 0       |                |
// | laplace | float               | NO   |     | 0       |                |
// +---------+---------------------+------+-----+---------+----------------+
// 11 rows in set (0.01 sec)

func floattostr(f float64) string {
	return strconv.FormatFloat(f, 'g', 5, 64)
}

func inttostr(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func int64tostr(i int64) string {
	return strconv.FormatInt(i, 10)
}

func uint64tostr(u uint64) string {
	return strconv.FormatUint(u, 10)
}

func htm(s string) string {
	return html.EscapeString(s)
}

func recalculateLaplace(db mysql.Conn, userid uint64, setid uint64, popFactor float64) {
	var book struct {
		idBook uint64
		stars5 uint64
		stars4 uint64
		stars3 uint64
		stars2 uint64
		stars1 uint64
	}
	stmt, err := db.Prepare("UPDATE blind_set SET pop_fact = ? WHERE (id_user = ?) AND (id_set = ?);")
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	stmt.Bind(popFactor, userid, setid)
	_, _, err = stmt.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec failed")
	}
	bookList := make([]uint64, 0, 0)
	laplaceList := make([]float64, 0, 0)
	sel, err := db.Prepare("SELECT id_book, stars5, stars4, stars3, stars2, stars1 FROM blind_book WHERE (id_user = ?) AND (id_set = ?);")
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, setid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	count := 0
	for _, row := range rows {
		book.idBook = row.Uint64(0)
		book.stars5 = row.Uint64(1)
		book.stars4 = row.Uint64(2)
		book.stars3 = row.Uint64(3)
		book.stars2 = row.Uint64(4)
		book.stars1 = row.Uint64(5)
		number := book.stars5 + book.stars4 + book.stars3 + book.stars2 + book.stars1
		numerator := float64((book.stars5 * 5) + (book.stars4 * 4) + (book.stars3 * 3) + (book.stars2 * 2) + (book.stars1))
		laplaceNum := (numerator + (3.0*popFactor))/(float64(number)+popFactor)
		bookList = append(bookList, book.idBook)
		laplaceList = append(laplaceList, laplaceNum)
		count++
	}
	for i := 0; i < count; i++ {
		bookId := bookList[i]
		laplaceNum := laplaceList[i]
		stmt, err := db.Prepare("UPDATE blind_book SET laplace = ? WHERE (id_user = ?) AND (id_set = ?) AND (id_book = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(laplaceNum, userid, setid, bookId)
		_, _, err = stmt.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Exec failed")
		}
	}

}

func showBlindMenuBar(w http.ResponseWriter) {
	fmt.Fprint(w, `
<p><a href="scrape">Scrape</a></p>
`)
}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var sql string
	var book struct {
		idBook  uint64
		title   string
		authors string
		stars5  uint64
		stars4  uint64
		stars3  uint64
		stars2  uint64
		stars1  uint64
		laplace float64
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Books</title>
<link rel="stylesheet" type="text/css" href="/style.css">

</head>
<body>
  <section>
`)
	if userid == 1 {
		showBlindMenuBar(w)
	}
	fmt.Fprint(w, `
    <h1>List of Books</h1>
`)
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	_, setExists := getform["set"]
	var setid uint64
	if setExists {
		setid, err = strconv.ParseUint(getform["set"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
	} else {
		fmt.Println("set ID is missing")
		panic("set ID is missing")
	}
	var popFactor float64
	_, popExists := getform["pop"]
	if popExists {
		popFactor, err = strconv.ParseFloat(getform["pop"][0], 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseFloat failed")
		}
		recalculateLaplace(db, userid, setid, popFactor)
	} else {
		sql = "SELECT pop_fact FROM blind_set WHERE (id_set = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(setid, userid)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Bind/Exec failed")
		}
		for _, row := range rows {
			popFactor = row.Float(0)
		}

	}
	fmt.Fprint(w, `<form action="list" method="get" /><input type="hidden" name="set" value="`+uint64tostr(setid)+`" />
<p>
	Popularity factor: <input type="text" name="pop" value="`+floattostr(popFactor)+`" /> <input type="submit" name="submit" value="Go" />
</p>
`)
	sql = "SELECT id_book, title, authors, stars5, stars4, stars3, stars2, stars1, laplace FROM blind_book WHERE (id_user = ?) AND (id_set = ?) ORDER BY laplace DESC;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, setid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	started := false
	count := 0
	for _, row := range rows {
		book.idBook = row.Uint64(0)
		book.title = row.Str(1)
		book.authors = row.Str(2)
		book.stars5 = row.Uint64(3)
		book.stars4 = row.Uint64(4)
		book.stars3 = row.Uint64(5)
		book.stars2 = row.Uint64(6)
		book.stars1 = row.Uint64(7)
		book.laplace = row.Float(8)

		if !started {
			fmt.Fprint(w, `<table border="0" cellpadding="6">
<tr>
<th> Rank </th><th> Title </th><th> Authors </th><th> 5 Stars </th><th> 4 Stars </th><th> 3 Stars </th><th> 2 Stars </th><th> 1 Stars </th><th> Average </th><th> L score </th>
</tr>
`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		count++
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+inttostr(count)+" </td><td> "+htm(book.title)+" </td><td> "+html.EscapeString(book.authors)+" </td>")
		average := float64((book.stars5*5)+(book.stars4*4)+(book.stars3*3)+(book.stars2*2)+book.stars1) / float64(book.stars5+book.stars4+book.stars3+book.stars2+book.stars1)

		fmt.Fprint(w, `<td align="right">`+strconv.FormatUint(book.stars5, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars4, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars3, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars2, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars1, 10)+` </td><td align="right">`+floattostr(average)+` </td><td align="right">`+floattostr(book.laplace)+` </td></tr>
`)
	}
	if started {
		fmt.Fprint(w, "</table></form>")
	}

	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func recalculateAllEvents(db mysql.Conn, userid uint64, recalcDefunct bool) {
	var entry struct {
		idCalEnt uint64
		year     string
		month    string
		dom      string
		dow      string
		nth      string
		hour     string
		minute   string
		second   string
	}
	sql := "SELECT COUNT(*) FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	isDefunct := 0
	if recalcDefunct {
		isDefunct = 1
	}
	sel.Bind(userid, isDefunct)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var size uint64
	for _, row := range rows {
		size = row.Uint64(0)
	}
	type updateItem struct {
		entryid     uint64
		currenttime uint64
	}
	var updateList []updateItem
	updateList = make([]updateItem, size)

	count := 0
	var timeZoneOffset int64
	var currenttime uint64
	timeZoneOffset = getTimeZoneOffset(db, userid)
	rightNowWithTimeZone := uint64(time.Now().Unix() + timeZoneOffset)

	sql = "SELECT id_cal_ent, year, month, dom, dow, nth, hour, minute, second FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?) ORDER BY id_cal_ent;"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, isDefunct)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		entry.idCalEnt = row.Uint64(0)
		entry.year = row.Str(1)
		entry.month = row.Str(2)
		entry.dom = row.Str(3)
		entry.dow = row.Str(4)
		entry.nth = row.Str(5)
		entry.hour = row.Str(6)
		entry.minute = row.Str(7)
		entry.second = row.Str(8)
		currenttime = calculateNextCurrentTimeForEvent(rightNowWithTimeZone, entry.year, entry.month, entry.dom, entry.dow, entry.nth, entry.hour, entry.minute, entry.second)
		currenttime = uint64(int64(currenttime) - timeZoneOffset)
		updateList[count] = updateItem{entryid: entry.idCalEnt, currenttime: currenttime}
		count++
	}
	var save struct {
		idCalEnt    uint64
		currenttime uint64
		isDefunct   int
	}
	cutoff := rightNowWithTimeZone - 259200 // 3 days
	for i := 0; i < count; i++ {
		stmt, err := db.Prepare("UPDATE calcron_entry SET currenttime = ?, is_defunct = ? WHERE id_cal_ent = ?;")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		save.idCalEnt = updateList[i].entryid
		save.currenttime = updateList[i].currenttime
		save.isDefunct = 0
		fmt.Println("comparing ", save.currenttime, " to ", cutoff)
		if save.currenttime < cutoff {
			save.isDefunct = 1
			fmt.Println("isDefunct set to true!!!")
		}
		stmt.Bind(save.currenttime, save.isDefunct, save.idCalEnt)
		_, _, err = stmt.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Exec failed")
		}
	}
}

func showHumanIntervalJS(w http.ResponseWriter) {
	fmt.Fprint(w, `

function humanInterval(seconds) {
    "use strict";
    var interval, minutes, hours, days, weeks, years, intText, nonZero;
    intText = "";
    if (seconds < 0) {
        seconds = Math.abs(seconds);
        intText = " - ";
    }
    minutes = Math.floor(seconds / 60);
    seconds = seconds - (minutes * 60);
    hours = Math.floor(minutes / 60);
    minutes = minutes - (hours * 60);
    days = Math.floor(hours / 24);
    hours = hours - (days * 24);
    weeks = Math.floor(days / 7);
    days = days - (weeks * 7);
    years = Math.floor(weeks / 52);
    weeks = weeks - (years * 52);
    nonZero = false;
    if (years !== 0) {
        intText = intText + " " + years + " years";
        nonZero = true;
    }
    if ((weeks !== 0) || nonZero) {
        intText = intText + " " + weeks + " weeks";
        nonZero = true;
    }
    if ((days !== 0) || nonZero) {
        intText = intText + " " + days + " days";
        nonZero = true;
    }
    if ((hours !== 0) || nonZero) {
        intText = intText + " " + hours + " hours";
        nonZero = true;
    }
    if ((minutes !== 0) || nonZero) {
        intText = intText + " " + minutes + " min";
        nonZero = true;
    }
    if ((seconds !== 0) || nonZero) {
        intText = intText + " " + seconds + " sec";
        nonZero = true;
    }
    intText = intText.substring(1);
    return intText;
}

`)
}

func showChimesPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	db, err := getDbConnection()
	if err != nil {
		fmt.Println(err)
		panic("getDbConnection failed")
	}
	recalculateAllEvents(db, userid, false)
	var entry struct {
		title       string
		description string
		currenttime uint64
	}
	rightNow := uint64(time.Now().Unix())
	sql := "SELECT title, description, currenttime FROM calcron_entry WHERE currenttime >= ? ORDER BY currenttime LIMIT 1;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(rightNow) // nothing to bind for this one
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		entry.title = row.Str(0)
		entry.description = row.Str(1)
		entry.currenttime = row.Uint64(2)
	}
	timeZoneOffset := getTimeZoneOffset(db, userid)
	timeNums := timeCodeToTimeNumbers(uint64(int64(entry.currenttime) + timeZoneOffset))
	dateTimeString := timeCodeToString(entry.currenttime, timeZoneOffset)
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Chimes: `+html.EscapeString(entry.title)+`</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>

<script>

// (C) 2016 Wayne Radinsky

/*jslint browser: true, devel: true, passfail: true */

// have to pre-declare some globals to pass jslint
/*global jQuery */

var gChimesData;

function chimesGetAudioContext() {
    "use strict";
    if (!window.AudioContext) {
        if (!window.webkitAudioContext) {
            window.alert("Your browser does not support the WebAudio API (HTML5).\nPlease use a compatible browser e.g. Google Chrome. If your WebAudio is disabled in Chrome, enable it in page about:flags");
            return;
        }
        window.AudioContext = window.webkitAudioContext;
    }
    var ctx = new window.AudioContext();
    return ctx;
}

function chimesGenerateANote(frequency, duration, amplitude) {
    "use strict";
    // consult https://dvcs.w3.org/hg/audio/raw-file/tip/webaudio/specification.html
    var numSamples, theBuffer, bufData, i, fade;
    numSamples = duration * gChimesData.SAMPLERATE;
    if (numSamples === 0) {
        window.alert("we hit the zero samples condition.");
    }
    console.log("allocating a buffer");
    theBuffer = gChimesData.globalCtx.createBuffer(1, numSamples, gChimesData.SAMPLERATE); // numberOfChannels, length, sampleRate
    bufData = theBuffer.getChannelData(0);
    for (i = 0; i <= numSamples; i = i + 1) {
        fade = 1 - (i / numSamples);
        bufData[i] = Math.sin((gChimesData.TAU * i * frequency) / gChimesData.SAMPLERATE) * ((amplitude * fade) / 2);
    }
    return theBuffer;
}

function chimesQueUpANote(fromMoment, frequency, startTime, duration, amplitude) {
    "use strict";
    var idxname, theBuffer, node, ourStart, startMmt;
    idxname = frequency + "x" + duration + "x" + amplitude;
    if (gChimesData.cachedNotes.hasOwnProperty(idxname)) {
        theBuffer = gChimesData.cachedNotes[idxname];
    } else {
        theBuffer = chimesGenerateANote(frequency, duration, amplitude);
        gChimesData.cachedNotes[idxname] = theBuffer;
    }
    node = gChimesData.globalCtx.createBufferSource();
    node.buffer = theBuffer;
    node.connect(gChimesData.globalCtx.destination);
    ourStart = Number(startTime);
    startMmt = fromMoment + ourStart;
    node.start(startMmt);
}

function chimesFindMostHarmonicNote(pitch, scale) {
    "use strict";
    var octave, minIdx, minDist, i, dist, rv, numOfNotes;
    octave = Math.floor(pitch);
    pitch = pitch - octave;
    minIdx = -1;
    minDist = 1.0 - pitch;
    numOfNotes = scale.length;
    for (i = 0; i < numOfNotes; i = i + 1) {
        dist = Math.abs(scale[i].fpt - pitch);
        if (dist < minDist) {
            minIdx = i;
            minDist = dist;
        }
    }
    if (minIdx === -1) {
        octave = octave + 1;
        rv = {octave: octave, scalenote: 0, num: scale[0].num, dnom: scale[0].dnom, fpt: scale[0].fpt };
    } else {
        rv = {octave: octave, scalenote: minIdx, num: scale[minIdx].num, dnom: scale[minIdx].dnom, fpt: scale[minIdx].fpt };
    }
    return rv;
}

function chimesGetScale7MajorReg() {
    "use strict";
    var scale = [ {num: 1, dnom: 1}, {num: 9, dnom: 8}, {num: 5, dnom: 4}, {num: 4, dnom: 3}, {num: 3, dnom: 2}, {num: 5, dnom: 3}, {num: 15, dnom: 8} ];
    return scale;
}

function chimesAddFloatsToScale(originalScale) {
    "use strict";
    var num, dnom, x, newScale, i;
    newScale = [];
    for (i = 0; i < originalScale.length; i = i + 1) {
        num = originalScale[i].num;
        dnom = originalScale[i].dnom;
        x = num / dnom;
        x = Math.log(x) / gChimesData.LOG2;
        newScale[i] = {num: num, dnom: dnom, fpt: x};
    }
    return newScale;
}

`)
	showHumanIntervalJS(w)
	fmt.Fprint(w, `

function chimesDoTimerPulse() {
    "use strict";
    var curtim, curdat, interval, xl, fromMoment, pitch, duration, amplitude, frequency, startTime, harmonizedInfo, harmOctave, harmNum, harmDnom, theScale, seconds, intText;
    // console.log("timer pulse!");
    curdat = new Date();
    curtim = curdat.getTime();
    interval = gChimesData.nextThing - curtim;
    // interval = Math.abs(interval);
    seconds = Math.floor(interval / 1000);
    if (seconds !== gChimesData.lastSecs) {
        // intText = years + " years " + weeks + " weeks " + days + " days " + hours + " hours " + minutes + " min " + seconds + " sec";
        // document.getElementById("interval").value = intText;
        intText = humanInterval(seconds);
        if (intText.substring(0, 1) === "-") {
            document.getElementById("interv_txt").innerHTML = '<font color="red">' + intText + '</font>';
        } else {
            document.getElementById("interv_txt").innerHTML = intText;
        }
    }
    if (interval < 0) {
        // window.alert("hello");
        // gChimesData.timerGoing = false;
        amplitude = 0.9;
        frequency = Math.exp((gChimesData.centernoteLog + 9) * gChimesData.LOG2);
        startTime = 0;
        duration = 0.1;
        fromMoment = gChimesData.globalCtx.currentTime;
        chimesQueUpANote(fromMoment, frequency, startTime, duration, amplitude);
        xl = Math.log(Math.abs(interval)) / gChimesData.LOG2;
        xl = xl * 5;
        document.getElementById("xl").value = xl;
        xl = Math.floor(xl);
    } else {
        xl = Math.log(interval) / gChimesData.LOG2;
        pitch = 9.0 - (xl / 4.7);
        xl = xl * 5;
        document.getElementById("xl").value = xl;
        xl = Math.floor(xl);
        if (xl !== gChimesData.lastXl) {
            fromMoment = gChimesData.globalCtx.currentTime;
            if (gChimesData.lastMoment === 0) {
                duration = 1.1;
            } else {
                duration = (fromMoment - gChimesData.lastMoment) / 2;
                if (duration <= 0) {
                    duration = 0.1; // but why does this ever happen?
                } else {
                    if (duration > 20) {
                        duration = 20;
                    }
                }
            }
            gChimesData.lastMoment = fromMoment;
            amplitude = 0.3333;
            theScale = chimesAddFloatsToScale(chimesGetScale7MajorReg());
            harmonizedInfo = chimesFindMostHarmonicNote(pitch, theScale);
            harmOctave = harmonizedInfo.octave;
            harmNum = harmonizedInfo.num;
            harmDnom = harmonizedInfo.dnom;
            frequency = Math.exp((gChimesData.centernoteLog + harmOctave) * gChimesData.LOG2) * (harmNum / harmDnom);
            startTime = 0;
            chimesQueUpANote(fromMoment, frequency, startTime, duration, amplitude);
            gChimesData.lastXl = xl;
        }
    }
    if (gChimesData.timerGoing) {
        if (xl < 40) {
            window.setTimeout(chimesDoTimerPulse, 20);
        } else {
            if (xl < 70) {
                window.setTimeout(chimesDoTimerPulse, 100);
            } else {
                if (xl < 80) {
                    window.setTimeout(chimesDoTimerPulse, 250);
                } else {
                    if (xl < 100) {
                        window.setTimeout(chimesDoTimerPulse, 1000);
                    } else {
                        if (interval < 0) {
                            window.setTimeout(chimesDoTimerPulse, 10000);
                        } else {
                            window.setTimeout(chimesDoTimerPulse, 1000);
                        }
                    }
                }
            }
        }
    }
}

function chimesConvertTimeCodeToOurString(timeCode) {
    "use strict";
    var str, i;
    str = timeCode.toISOString();
    i = str.indexOf(".");
    str = str.substring(0, i);
    i = str.indexOf("T");
    str = str.substring(0, i) + " " + str.substring(i + 1);
    return str;
}

function chimesConvertOurStringToTimeCode(datestr) {
    "use strict";
    var i;
    i = datestr.indexOf(" ");
    datestr = datestr.substring(0, i) + "T" + datestr.substring(i + 1);
    datestr = datestr + ".000Z";
    return Date.parse(datestr);
}

function chimesExecParseAndSet() {
    "use strict";
    var datestr, thetime, addMinutes;
    datestr = document.getElementById("as_text").value;
    // thetime = Date.parse(datestr) + 25200000;
    thetime = chimesConvertOurStringToTimeCode(datestr) + 25200000;
    addMinutes = document.getElementById("add_minutes").value;
    if (addMinutes !== "") {
        addMinutes = parseInt(addMinutes, 10);
        addMinutes = addMinutes * 60 * 1000;
        thetime = thetime + addMinutes;
    }
    // document.getElementById("next_event").value = thetime;
    gChimesData.nextThing = thetime;
    if (!gChimesData.timerGoing) {
        window.setTimeout(chimesDoTimerPulse, 1000);
        gChimesData.timerGoing = true;
    }
    gChimesData.centernoteLog = (Math.log(432) / Math.log(2)) - 5;
}

function chimesExecUseTimeNow() {
    "use strict";
    var curtim;
    curtim = new Date(); // year, month [, day, hour, minute, second, millisecond]);
    curtim.setTime(curtim.getTime() - 25200000);
    document.getElementById("next_event").value = curtim.getTime();
    // document.getElementById("as_text").value = curtim.toISOString();
    document.getElementById("as_text").value = chimesConvertTimeCodeToOurString(curtim);
    document.getElementById("title").innerHTML = "";
    document.getElementById("description").innerHTML = "";
}

// main

gChimesData = {SAMPLERATE: 44100, TAU: Math.PI * 2, LOG2: Math.log(2), globalCtx: chimesGetAudioContext(), cachedNotes: {}, timerGoing: false, lastXl: 0, lastMoment: 0, lastSecs: 0 };

jQuery(function () {
    "use strict";
    // var curtim;
    // curtim = new Date(); // year, month [, day, hour, minute, second, millisecond]);
    // curtim.setTime(curtim.getTime() - 25200000);
    // document.getElementById("next_event").value = curtim.getTime();
    // document.getElementById("as_text").value = curtim.toISOString();
    chimesExecParseAndSet();
});

</script>

</head>
<body>
  <section>
<form>

<p> Next event: `)
	fmt.Fprint(w, html.EscapeString(dateTimeString))
	fmt.Fprint(w, `</p>
    <h1 id="title"><b>`)
	fmt.Fprint(w, html.EscapeString(entry.title))
	fmt.Fprint(w, `</b></h1>
<p id="description">
`)
	fmt.Fprint(w, html.EscapeString(entry.description))
	fmt.Fprint(w, `
</p>
<p> Time code: 
    <input type="textbox" id="as_text" name="as_text" value="`)
	// format example 2016-04-25T16:32:34.995Z
	// fmt.Fprintf(w, "%d-%s-%sT%s:%s:%s.000Z", timeNums.year, twodigits(timeNums.month), twodigits(timeNums.day), twodigits(timeNums.hour), twodigits(timeNums.minute), twodigits(timeNums.second))
	fmt.Fprintf(w, "%d-%s-%s %s:%s:%s", timeNums.year, twodigits(timeNums.month), twodigits(timeNums.day), twodigits(timeNums.hour), twodigits(timeNums.minute), twodigits(timeNums.second))
	fmt.Fprint(w, `" />
    <input type="button" id="use_time_now" name="use_time_now" value="Use time now" />
</p><p style="display:none;">
    <input type="textbox" id="next_event" name="next_event" value="" />
</p><p> Add minutes:
    <input type="textbox" id="add_minutes" name="add_minutes" value="" />
</p><p>
    <input type="button" id="parse_and_set" name="parse_and_set" value="Parse And Set Time" />
</p><p><table style="width:300px;"><tr><td align="right"> <span id="interv_txt"></span> </td></tr></table>
</p><p style="display:none;">
    <input type="textbox" id="interval" name="interval" value="interval" style="width:400px;" />
</p><p style="display:block;">
    <input type="textbox" id="xl" name="xl" value="lx" />
</p>

</form>

  </section>

<script>

var chimesListenElement;
chimesListenElement = document.getElementById("parse_and_set");
chimesListenElement.addEventListener("click", chimesExecParseAndSet, true);

chimesListenElement = document.getElementById("use_time_now");
chimesListenElement.addEventListener("click", chimesExecUseTimeNow, true);

</script>

</body>
</html>`)
}

func showViewPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var ui struct {
		title       string
		description string
		year        string
		month       string
		dom         string
		dow         string
		nth         string
		hour        string
		minute      string
		second      string
	}
	var entryid uint64
	entryid = 0

	// set defaults
	ui.title = ""
	ui.description = ""
	ui.year = "*"
	ui.month = "*"
	ui.dom = "*"
	ui.dow = "*"
	ui.nth = "*"
	ui.hour = "*"
	ui.minute = "0"
	ui.second = "0"
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("parseform failed")
	}
	getform := r.Form
	_, entry := getform["entry"]
	if entry {
		entryid, err = strconv.ParseUint(getform["entry"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
		db, err := getDbConnection()
		if err != nil {
			fmt.Println(err)
			panic("getDbConnection failed")
		}
		defer db.Close()
		sql := "SELECT title, description, year, month, dom, dow, nth, hour, minute, second FROM calcron_entry WHERE (id_cal_ent = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(entryid, userid)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Bind/Exec failed")
		}
		for _, row := range rows {
			ui.title = row.Str(0)
			ui.description = row.Str(1)
			ui.year = row.Str(2)
			ui.month = row.Str(3)
			ui.dom = row.Str(4)
			ui.dow = row.Str(5)
			ui.nth = row.Str(6)
			ui.hour = row.Str(7)
			ui.minute = row.Str(8)
			ui.second = row.Str(9)
		}
	}

	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer db.Close()
	fmt.Fprint(w, `<title>CalCron Entry</title>
<link rel="stylesheet" type="text/css" href="/style.css">

</head>
<body>
  <section>
    <h1>CalCron Entry</h1>

`)
	fmt.Fprint(w, `

<table border="0" cellpadding="4">
<tr><td align="right"> Title: </td><td> `+html.EscapeString(ui.title)+` </td></tr>
<tr><td align="right"> Description: </td><td> `+html.EscapeString(ui.description)+` </td></tr>
<tr><td align="right"> Year: </td><td> `+html.EscapeString(ui.year)+` </td></tr>
<tr><td align="right"> Month: </td><td> `+html.EscapeString(ui.month)+` </td></tr>
<tr><td align="right"> Day of month: </td><td> `+html.EscapeString(ui.dom)+` </td></tr>
<tr><td colspan="2" align="center" style="border-bottom: 1px solid #000000;"> </td></tr>
<tr><td align="right"> Day of week: </td><td> `+html.EscapeString(ui.dow)+` </td></tr>
<tr><td align="right"> Nth day of day of week this month: </td><td> `+html.EscapeString(ui.nth)+` </td></tr>
<tr><td colspan="2" align="center" style="border-bottom: 1px solid #000000;"> </td></tr>
<tr><td align="right"> Hour: </td><td> `+html.EscapeString(ui.hour)+` </td></tr>
<tr><td align="right"> Minute: </td><td> `+html.EscapeString(ui.minute)+` </td></tr>
<tr><td align="right"> Second: </td><td> `+html.EscapeString(ui.second)+` </td></tr>

<tr><td colspan="2" align="center"> <a href="edit?entry=`+strconv.FormatUint(entryid, 10)+`">Edit</a> </td></tr>

</table>
</form>

  </section>

</body>
</html>`)
}

func parseStarsPct(stars string, starToFind string) (float64, error) {
	idx1 := strings.Index(stars, starToFind)
	if idx1 == -1 {
		return 0.0, errors.New("Could not find star to find")
	}
	idx2 := strings.Index(stars[idx1:], "%")
	if idx2 == -1 {
		return 0.0, errors.New("Could not find percent sign")
	}
	idx2 = idx2 + idx1
	pct, err := strconv.ParseFloat(stars[idx1+10:idx2], 64)
	if err != nil {
		return 0.0, err
	}
	return pct, nil
}

func round(x float64) int {
	fmt.Println("round: ", x)
	fmt.Println("    floor: ", math.Floor(x+0.5))
	fmt.Println("    to int: ", int(math.Floor(x+0.5)))
	return int(math.Floor(x + 0.5))
}

func showScrapePage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		titleauthors string
		stars        string
	}
	var bookid uint64
	bookid = 0
	existingTitle := ""
	existingAuthors := ""
	if method == "GET" {
		// set defaults
		ui.titleauthors = ""
		ui.stars = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, bookspecified := getform["book"]
		if bookspecified {
			bookid, err = strconv.ParseUint(getform["book"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db, err := getDbConnection()
			if err != nil {
				fmt.Println(err)
				panic("getDbConnection failed")
			}
			defer db.Close()
			sql := "SELECT title, authors FROM blind_book WHERE (id_book = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(bookid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				existingTitle = row.Str(0)
				existingAuthors = row.Str(1)
			}
		}
	}
	title := ""
	authors := ""
	stars5 := 0
	stars4 := 0
	stars3 := 0
	stars2 := 0
	stars1 := 0
	if method == "POST" {
		fmt.Println("point 1725")
		// set from form post
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		postform := r.Form
		// error checking
		bookid, err = strconv.ParseUint(postform["book"][0], 10, 64)
		if err != nil {
			bookid = 0
		}
		ui.titleauthors = strings.Trim(postform["titleauthors"][0], " \r\n\t")
		ui.stars = strings.Trim(postform["stars"][0], " \r\n\t")
		if ui.titleauthors == "" {
			errorList["titleauthors"] = "The Title & Authors box is empty."
			errorOccurred = true
		}
		if ui.stars == "" {
			errorList["stars"] = "The stars box is empty."
			errorOccurred = true
		}
		// Ok! Here's our actual parsing rountine! If anything goes wrong, we set the error
		byIdx := strings.Index(ui.titleauthors, "by")
		if byIdx == -1 {
			errorList["titleauthors"] = "Can't find by"
			errorOccurred = true
		} else {
			ttlIdx := byIdx
			anotherIdx := strings.Index(ui.titleauthors, "Hardcover")
			if anotherIdx != -1 {
				if anotherIdx < ttlIdx {
					ttlIdx = anotherIdx
				}
			}
			anotherIdx = strings.Index(ui.titleauthors, "Kindle")
			if anotherIdx != -1 {
				if anotherIdx < ttlIdx {
					ttlIdx = anotherIdx
				}
			}
			anotherIdx = strings.Index(ui.titleauthors, "Paperback")
			if anotherIdx != -1 {
				if anotherIdx < ttlIdx {
					ttlIdx = anotherIdx
				}
			}
			title = strings.Trim(ui.titleauthors[:ttlIdx], " \r\n\t")
			authors = strings.Trim(ui.titleauthors[byIdx+3:], " \r\n\t")
			outOfIdx := strings.Index(ui.stars, "out of")
			if outOfIdx == -1 {
				errorList["stars"] = "Could not find out of in stars field."
				errorOccurred = true
			} else {
				starsAfterOutOf := ui.stars[outOfIdx+14:]
				st5pct, pct5err := parseStarsPct(starsAfterOutOf, "5 star")
				st4pct, pct4err := parseStarsPct(starsAfterOutOf, "4 star")
				st3pct, pct3err := parseStarsPct(starsAfterOutOf, "3 star")
				st2pct, pct2err := parseStarsPct(starsAfterOutOf, "2 star")
				st1pct, pct1err := parseStarsPct(starsAfterOutOf, "1 star")
				if (pct5err != nil) || (pct4err != nil) || (pct3err != nil) || (pct2err != nil) || (pct1err != nil) {
					errorList["stars"] = "One of the stars percentages failed to parse."
					errorOccurred = true
				} else {
					seeAllIdx := strings.Index(ui.stars, "See all")
					reviewsIdx := strings.Index(ui.stars, "reviews")
					if (seeAllIdx == -1) || (reviewsIdx == -1) {
						errorList["stars"] = "The reviews count failed to parse."
						errorOccurred = true
					} else {
						reviewsCount, err := strconv.ParseInt(ui.stars[seeAllIdx+8:reviewsIdx-1], 10, 64)
						if err != nil {
							errorList["stars"] = "The reviews count was found but the found count failed to parse."
							errorOccurred = true
						}
						stars5 = round((st5pct / 100.0) * float64(reviewsCount))
						stars4 = round((st4pct / 100.0) * float64(reviewsCount))
						stars3 = round((st3pct / 100.0) * float64(reviewsCount))
						stars2 = round((st2pct / 100.0) * float64(reviewsCount))
						stars1 = round((st1pct / 100.0) * float64(reviewsCount))
					}
				}
			}
		}
		if errorOccurred {
			showform = true
		} else {
			// dbConnect!!
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			var save struct {
				idBook  uint64
				idUser  uint64
				idSet   uint64
				title   string
				authors string
				stars5  int
				stars4  int
				stars3  int
				stars2  int
				stars1  int
			}
			save.idBook = bookid
			save.idUser = userid
			save.idSet = 3 // hardcoded for now
			save.title = title
			save.authors = authors
			save.stars5 = stars5
			save.stars4 = stars4
			save.stars3 = stars3
			save.stars2 = stars2
			save.stars1 = stars1
			// query, if there, update, if not, create new
			alreadyExists := false
			if bookid > 0 {
				sql := "SELECT id_book FROM blind_book WHERE (id_book = ?) AND (id_user = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(bookid, userid)
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
				stmt, err := db.Prepare("UPDATE blind_book SET title = ?, authors = ?, stars5 = ?, stars4 = ?, stars3 = ?, stars2 = ?, stars1 = ? WHERE (id_book = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.title, save.authors, save.stars5, save.stars4, save.stars3, save.stars2, save.stars1, bookid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO blind_book (id_user, id_set, title, authors, stars5, stars4, stars3, stars2, stars1) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idUser, save.idSet, save.title, save.authors, save.stars5, save.stars4, save.stars3, save.stars2, save.stars1)
				_, _, err = stmt.Exec()
			}
			if err != nil {
				fmt.Println(err)
				panic("Exec failed")
			}
			setid := 3 // BUGBUG this is hardcoded but shouldn't be
			http.Redirect(w, r, "list?set="+inttostr(setid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer db.Close()
		fmt.Fprint(w, `<title>Blind Book Scrape</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
    <h1>Blind Book Scrape</h1>
`)
		if bookid != 0 {
			fmt.Fprintln(w, `<p>Modifying existing book ID `+strconv.FormatUint(bookid, 10)+`: `+html.EscapeString(existingTitle)+` by `+html.EscapeString(existingAuthors)+`</p>`)

		}
		fmt.Fprint(w, `
<form action="scrape" method="post">
<input type="hidden" name="book" value="`+strconv.FormatUint(bookid, 10)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<h2>Title & Author(s)</h2>
<p>
<textarea name="titleauthors" id="titleauthors" cols="120" rows="5">`+html.EscapeString(ui.titleauthors)+`</textarea>
</p>

<h2>Stars</h2>
<p>
<textarea name="stars" id="stars" cols="120" rows="5">`+html.EscapeString(ui.stars)+`</textarea>
</p>

<p>
<input type="submit"> 
</p>

</form>
  </section>
</body>
</html>`)
	}
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	fmt.Println("We are in the blind handler, and op is", op)
	testCheckTimeFieldSyntaxError()
	testApplyRule()
	switch {
	case op == "scrape":
		if userid != 0 {
			showScrapePage(w, r, op, userid)
		}
	case op == "list":
		if userid != 0 {
			showListPage(w, r, op, userid)
		}
	case op == "edit":
		if userid != 0 {
			showEditPage(w, r, op, userid)
		}
	case op == "add":
		if userid != 0 {
			showEditPage(w, r, op, userid)
		}
	case op == "chimes":
		if userid != 0 {
			showChimesPage(w, r, op, userid)
		}
	case op == "view":
		if userid != 0 {
			showViewPage(w, r, op, userid)
		}
	default:
		// fmt.Fprintln(w, "Could not find page:", op)
		filename := "/home/ec2-user/wayneserver/staticappcontent/blind/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
