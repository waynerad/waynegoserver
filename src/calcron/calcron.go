package calcron

import (
	"accessdb"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
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

func htm(s string) string {
	return html.EscapeString(s)
}

func strToInt(str string) int {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("ParseInt failed")
	}
	return int(value)
}

func intToStr(z int) string {
	return strconv.FormatInt(int64(z), 10)
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

func checkTimeFieldSyntaxError(contents string) bool {
	if contents == "" {
		return true
	}
	if contents == "*" {
		return false
	}
	if checkIsNumeric(contents) {
		return false
	}
	hasComma := strings.Index(contents, ",")
	if hasComma > 0 {
		valStrs := strings.Split(contents, ",")
		num := len(valStrs)
		for i := 0; i < num; i++ {
			_, err := strconv.ParseInt(valStrs[i], 10, 64)
			if err != nil {
				return true
			}
		}
		return false
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
	if checkIsNumeric(contents[0:i]) {
		return false
	}
	return true
}

func testCheckTimeFieldSyntaxError() {
	if !checkTimeFieldSyntaxError("") {
		fmt.Println("checkTimeFieldSyntaxError test failed on empty string.")
	}
	if !checkTimeFieldSyntaxError("A") {
		fmt.Println("checkTimeFieldSyntaxError test failed on single letter.")
	}
	if !checkTimeFieldSyntaxError("word") {
		fmt.Println("checkTimeFieldSyntaxError test failed on letters.")
	}
	if !checkTimeFieldSyntaxError("5A") {
		fmt.Println("checkTimeFieldSyntaxError test failed on number and letter.")
	}
	if checkTimeFieldSyntaxError("5") {
		fmt.Println("checkTimeFieldSyntaxError test failed on single-digit number.")
	}
	if checkTimeFieldSyntaxError("55") {
		fmt.Println("checkTimeFieldSyntaxError test failed on double digit number.")
	}
	if checkTimeFieldSyntaxError("*") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk.")
	}
	if !checkTimeFieldSyntaxError("**") {
		fmt.Println("checkTimeFieldSyntaxError test failed on double asterisk.")
	}
	if !checkTimeFieldSyntaxError("*5") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk and number.")
	}
	if !checkTimeFieldSyntaxError("*A") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk and letter.")
	}
	if !checkTimeFieldSyntaxError("*/") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash.")
	}
	if !checkTimeFieldSyntaxError("*/A") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash letter.")
	}
	if checkTimeFieldSyntaxError("*/5") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash digit.")
	}
	if checkTimeFieldSyntaxError("*/6") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash double digit.")
	}
	if checkTimeFieldSyntaxError("*/888") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash triple digit.")
	}
	if checkTimeFieldSyntaxError("*/1234569") {
		fmt.Println("checkTimeFieldSyntaxError test failed on asterisk slash 8-digit.")
	}
	if !checkTimeFieldSyntaxError(" *") {
		fmt.Println("checkTimeFieldSyntaxError test failed on leading space.")
	}
	if checkTimeFieldSyntaxError("1,2,3") {
		fmt.Println("checkTimeFieldSyntaxError test failed on comma separated + allow slash.")
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

func breakSlashRuleIntoNumAndDnom(rule string) (num int, dnom int) {
	i := strings.Index(rule, "/")
	if i == -1 {
		panic("slash rule without slash")
	}
	if rule[0:i] == "*" {
		num = 0
	} else {
		num = strToInt(rule[0:i])
	}
	dnom = strToInt(rule[i+1:])
	return num, dnom
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
	hasSlash := strings.Index(rule, "/")
	if hasSlash > 0 {
		num, dnom := breakSlashRuleIntoNumAndDnom(rule)
		newCarry = 0
		currentValue = currentValue + currentCarry
		for (currentValue % dnom) != num {
			currentValue++
			for currentValue >= (fieldLimit + fieldMin) {
				newCarry++
				currentValue = currentValue - fieldLimit
			}
		}
		return currentValue, newCarry
	}
	hasComma := strings.Index(rule, ",")
	if hasComma > 0 {
		valStrs := strings.Split(rule, ",")
		num := len(valStrs)
		valInts := make([]int, num)
		for i := 0; i < num; i++ {
			valInts[i] = strToInt(valStrs[i])
		}
		currentPos := 0
		newCarry = 0
		currentValue = currentValue + currentCarry
		for valInts[currentPos] < currentValue {
			currentPos++
			if currentPos >= num {
				newCarry++
				currentValue = currentValue - fieldLimit
				if currentValue < fieldMin {
					currentValue = fieldMin
				}
				currentPos = 0
			}
		}
		currentValue = valInts[currentPos]
		return currentValue, newCarry
	}

	value, err := strconv.ParseInt(rule, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("ParseInt failed")
	}
	currentValue = currentValue + currentCarry
	for currentValue >= (fieldLimit + fieldMin) {
		newCarry++
		currentValue = currentValue - fieldLimit
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
	newval, carry = applyRule(1, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 0)
	if (newval != 2) || (carry != 0) {
		fmt.Println("test 13 failed for applyRule")
	}
	newval, carry = applyRule(3, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 0)
	if (newval != 3) || (carry != 0) {
		fmt.Println("test 14 failed for applyRule")
	}
	newval, carry = applyRule(3, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 1)
	if (newval != 5) || (carry != 0) {
		fmt.Println("test 15 failed for applyRule")
	}
	newval, carry = applyRule(30, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 0)
	if (newval != 31) || (carry != 0) {
		fmt.Println("test 16 failed for applyRule")
	}
	newval, carry = applyRule(31, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 1)
	if (newval != 37) || (carry != 0) {
		fmt.Println("test 17 failed for applyRule")
	}
	newval, carry = applyRule(45, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 0)
	if (newval != 47) || (carry != 0) {
		fmt.Println("test 18 failed for applyRule")
	}
	newval, carry = applyRule(45, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 1)
	if (newval != 47) || (carry != 0) {
		fmt.Println("test 19 failed for applyRule")
	}
	newval, carry = applyRule(59, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 0)
	if (newval != 59) || (carry != 0) {
		fmt.Println("test 20 failed for applyRule")
	}
	newval, carry = applyRule(59, "2,3,5,7,11,13,17,19,23,29,31,37,41,43,47,53,59", 60, 0, 1)
	if (newval != 2) || (carry != 1) {
		fmt.Println("test 21 failed for applyRule")
	}
	newval, carry = applyRule(45, "1/8", 60, 0, 0)
	if (newval != 49) || (carry != 0) {
		fmt.Println("test 22 failed for applyRule")
	}
	newval, carry = applyRule(45, "1/17", 60, 0, 0)
	if (newval != 52) || (carry != 0) {
		fmt.Println("test 23 failed for applyRule")
	}
	newval, carry = applyRule(45, "7/17", 60, 0, 0)
	if (newval != 58) || (carry != 0) {
		fmt.Println("test 24 failed for applyRule")
	}
	newval, carry = applyRule(53, "5/17", 60, 0, 0)
	if (newval != 56) || (carry != 0) {
		fmt.Println("test 25 failed for applyRule")
	}
	newval, carry = applyRule(59, "3/7", 60, 0, 1)
	if (newval != 3) || (carry != 1) {
		fmt.Println("test 26 failed for applyRule")
	}
	newval, carry = applyRule(59, "3/8", 60, 0, 1)
	if (newval != 3) || (carry != 1) {
		fmt.Println("test 27 failed for applyRule")
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

func makeIntListFromRule(rule string, starLimit int) []int {
	hasComma := strings.Index(rule, ",")
	if hasComma > 0 {
		valStrs := strings.Split(rule, ",")
		num := len(valStrs)
		resultList := make([]int, num)
		for i := 0; i < num; i++ {
			resultList[i] = strToInt(valStrs[i])
		}
		return resultList
	}
	if len(rule) >= 2 {
		hasSlash := strings.Index(rule, "/")
		if hasSlash > 0 {
			num, dnom := breakSlashRuleIntoNumAndDnom(rule)
			howMany := starLimit / dnom // integer division! Remainder is discarded
			resultList := make([]int, howMany)
			for i := 0; i < howMany; i++ {
				resultList[i] = (i * dnom) + num
			}
		}
	}
	if rule == "*" {
		resultList := make([]int, starLimit)
		for i := 0; i < starLimit; i++ {
			resultList[i] = i
		}
		return resultList
	}
	resultList := make([]int, 1)
	resultList[0] = strToInt(rule)
	return resultList
}

func calculateNextCurrentTimeForEvent(currentTimeCode uint64, yearRule string, monthRule string, domRule string, dowRule string, nthRule string, doeRule string, hourRule string, minuteRule string, secondRule string) uint64 {
	// currentTimeCode has the time zone offset added before calling this function. This convention is different from other functions. I forget why. Maybe I should change it
	timeNumbers := timeCodeToTimeNumbers(currentTimeCode)

	var result timenumbers

	// forward clearing pass

	foundDiff := false
	newval, _ := applyRule(timeNumbers.year, yearRule, 2147483647, 0, 0)
	if newval > timeNumbers.year {
		timeNumbers.year = newval
		foundDiff = true
	}

	newval, _ = applyRule(timeNumbers.month, monthRule, 12, 1, 0)
	if foundDiff {
		timeNumbers.month = 1
	} else {
		if newval > timeNumbers.month {
			timeNumbers.month = newval
			foundDiff = true
		}
	}

	monthlengs := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if (timeNumbers.year % 4) == 0 {
		monthlengs[1] = 29
	}
	daysinmonth := monthlengs[timeNumbers.month-1]
	newval, _ = applyRule(timeNumbers.day, domRule, daysinmonth, 1, 0)
	if foundDiff {
		timeNumbers.day = 1
	} else {
		if newval > timeNumbers.day {
			timeNumbers.day = newval
			foundDiff = true
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
		if newval > timeNumbers.minute {
			timeNumbers.minute = newval
			foundDiff = true
		}
	}

	newval, _ = applyRule(timeNumbers.second, secondRule, 60, 0, 0)
	if foundDiff {
		timeNumbers.second = 0
	} else {
		if newval > timeNumbers.second {
			timeNumbers.second = newval
			foundDiff = true
		}
	}

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

	// move forward to weekday if weekday specificed
	if dowRule != "*" {
		// This section is kind of a hack -- doesn't use the neat applyRule() function that everything else gets to use
		dowNums := makeIntListFromRule(dowRule, 7)
		nthNums := makeIntListFromRule(nthRule, 6)
		dayTime := time.Date(result.year, time.Month(result.month), result.day, 12, 0, 0, 0, time.UTC)
		breakloop := 0
		matchDow := false
		for i := 0; i < len(dowNums); i++ {
			if dayTime.Weekday() == time.Weekday(dowNums[i]) {
				matchDow = true
			}
		}
		matchNth := false
		checkCount := countOfDaysInThisMonth(dayTime)
		for i := 0; i < len(nthNums); i++ {
			if checkCount == nthNums[i] {
				matchNth = true
			}
		}
		for (!matchDow) || (!matchNth) {
			breakloop++
			if breakloop == 100 {
				panic("Stuck in endless loop.")
			}
			beforeMonth := dayTime.Month()
			dayTime = dayTime.Add(86400000000000) // 1 day
			afterMonth := dayTime.Month()
			if beforeMonth != afterMonth {
				result.day = int(dayTime.Day())
				result.month = int(dayTime.Month())
				newval, yearCarry := applyRule(result.month, monthRule, 12, 1, monthCarry)
				result.month = newval
				newval, _ = applyRule(result.year, yearRule, 2147483647, 0, yearCarry)
				result.year = newval
				dayTime = time.Date(result.year, time.Month(result.month), result.day, 12, 0, 0, 0, time.UTC)
			}
			matchDow = false
			for i := 0; i < len(dowNums); i++ {
				if dayTime.Weekday() == time.Weekday(dowNums[i]) {
					matchDow = true
				}
			}
			matchNth = false
			checkCount := countOfDaysInThisMonth(dayTime)
			for i := 0; i < len(nthNums); i++ {
				if checkCount == nthNums[i] {
					matchNth = true
				}
			}
		}
		result.day = int(dayTime.Day())
		result.month = int(dayTime.Month())
	}
	// move forward to day of eternity if specified
	hasSlash := strings.Index(doeRule, "/")
	if hasSlash > 0 {
		TOETimeCode := timeNumbersToTimeCode(result)
		currentDayOfE := TOETimeCode / 86400
		num, dnom := breakSlashRuleIntoNumAndDnom(doeRule)
		addDays := 0
		for (currentDayOfE % uint64(dnom)) != uint64(num) {
			currentDayOfE++
			addDays++
		}
		TOETimeCode += uint64(addDays) * 86400
		result = timeCodeToTimeNumbers(TOETimeCode)
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
	}
	return timeZoneOffset
}

func showCalcronMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<p><a href="chimes">Chimes</a>
<a href="add">Add</a>
<a href="list">List</a>
&middot;`+htm(userName)+`
</p>
`)
}

func showEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
		doe         string
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
		ui.doe = "*"
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT title, description, year, month, dom, dow, nth, doe, hour, minute, second FROM calcron_entry WHERE (id_cal_ent = ?) AND (id_user = ?);"
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
				ui.doe = row.Str(7)
				ui.hour = row.Str(8)
				ui.minute = row.Str(9)
				ui.second = row.Str(10)
			}
		}
	}
	if method == "POST" {
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
		ui.doe = strings.Trim(postform["doe"][0], " \r\n\t")
		ui.hour = strings.Trim(postform["hour"][0], " \r\n\t")
		ui.minute = strings.Trim(postform["minute"][0], " \r\n\t")
		ui.second = strings.Trim(postform["second"][0], " \r\n\t")
		if ui.title == "" {
			errorList["title"] = "Please specify a title."
		}
		if checkTimeFieldSyntaxError(ui.year) {
			errorList["year"] = "Year is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.month) {
			errorList["month"] = "Month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.dom) {
			errorList["dom"] = "Day of month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.dow) {
			errorList["dow"] = "Day of week is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.nth) {
			errorList["nth"] = "Nth day of week in month is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.doe) {
			errorList["doe"] = "Day of eternity is invalid"
			errorOccurred = true
		}
		if ui.doe != "*" {
			hasSlash := strings.Index(ui.doe, "/")
			if hasSlash <= 0 {
				errorList["doe"] = "Day of eternity must be a slash rule"
				errorOccurred = true
			}
		}
		if checkTimeFieldSyntaxError(ui.hour) {
			errorList["hour"] = "Hour is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.minute) {
			errorList["minute"] = "Minute is invalid"
			errorOccurred = true
		}
		if checkTimeFieldSyntaxError(ui.second) {
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
		if ui.doe != "*" {
			if ui.dom != "*" {
				errorList["dom"] = "If you specify a day of eternity, you cannot specify a day of month."
				errorOccurred = true
			}
			if ui.dow != "*" {
				errorList["dow"] = "If you specify a day of eternity, you cannot specify a day of week."
				errorOccurred = true
			}
			if ui.nth != "*" {
				errorList["nth"] = "If you specify a day of eternity, you cannot specify an nth day of week in month."
				errorOccurred = true
			}
		}
		if errorOccurred {
			showform = true
		} else {
			// dbConnect!!
			db := accessdb.GetDbConnection()
			defer db.Close()
			var timeZoneOffset int64
			timeZoneOffset = getTimeZoneOffset(db, userid)
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
				doe         string
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
			save.doe = ui.doe
			save.hour = ui.hour
			save.minute = ui.minute
			save.second = ui.second
			// save.currenttime = uint64(time.Now().Unix())
			// save.currenttime = uint64(int64(calculateNextCurrentTimeForEvent(uint64(int64(save.starttime) + timeZoneOffset), save.year, save.month, save.dom, save.dow, save.nth, save.doe, save.hour, save.minute, save.second)) - timeZoneOffset)
			save.currenttime = calculateNextCurrentTimeForEvent(uint64(int64(save.starttime)+timeZoneOffset), save.year, save.month, save.dom, save.dow, save.nth, save.doe, save.hour, save.minute, save.second)
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
				stmt, err := db.Prepare("UPDATE calcron_entry SET title = ?, description = ?, year = ?, month = ?, dom = ?, dow = ?, nth = ?, doe = ?, hour = ?, minute = ?, second = ?, currenttime = ?, is_defunct = 0 WHERE (id_cal_ent = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.title, save.description, save.year, save.month, save.dom, save.dow, save.nth, save.doe, save.hour, save.minute, save.second, save.currenttime, entryid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO calcron_entry (id_user, title, description, starttime, year, month, dom, dow, nth, doe, hour, minute, second, currenttime, is_defunct) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idUser, save.title, save.description, save.starttime, save.year, save.month, save.dom, save.dow, save.nth, save.doe, save.hour, save.minute, save.second, save.currenttime)
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
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>CalCron Entry</title>
<link rel="stylesheet" type="text/css" href="/style.css">

<script>

/*jslint browser: true, devel: true */

// (C) 2013-2017 Wayne Radinsky

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
`)
		showCalcronMenuBar(w, userName)
		fmt.Fprint(w, `
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
<tr><td align="right"> Day of eternity </td><td> <input class="biginput" name="doe" id="doe" type="text" value="`+html.EscapeString(ui.doe)+`" /> </td></tr>
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

func determineLowestBitOf(num int) int {
	bitvals := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768}
	if num == 0 {
		return 0 // have to treat 0 as a special case because otherwise it'll cause an endless loop
	}
	i := 0
	for {
		v := bitvals[i]
		if (num & v) != 0 {
			if (num & (v - 1)) == 0 {
				return i
			}
		}
		i = i + 1
	}
}

func getPriorityTitleAndDescriptionForDate(db mysql.Conn, userid uint64, currenttime uint64) (string, string) {
	to := time.Unix(int64(currenttime), 0)
	secs := to.Unix()
	days := (secs / 86400) - 17307 // integer division // constant determines starting day
	bitNum := determineLowestBitOf(int(days))
	bitNum++
	sql := "SELECT title, description FROM calcron_priority WHERE (id_user = ?) AND (priority = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, bitNum)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var priority struct {
		title       string
		description string
	}
	for _, row := range rows {
		priority.title = row.Str(0)
		priority.description = row.Str(1)
	}
	// return intToStr(int(days)) + " " + priority.title + " yeah bit " + intToStr(bitNum), priority.description
	return priority.title, priority.description
}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	var entry struct {
		idCalEnt    uint64
		title       string
		description string
		year        string
		month       string
		dom         string
		dow         string
		nth         string
		doe         string
		hour        string
		minute      string
		second      string
		currenttime uint64
	}
	rightNow := uint64(time.Now().Unix())
	db := accessdb.GetDbConnection()
	defer db.Close()
	var timeZoneOffset int64
	timeZoneOffset = getTimeZoneOffset(db, userid)
	timeZoneClientSideAdjustNum := strconv.FormatInt((-timeZoneOffset)*1000, 10)
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Calcron Entries</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>

<script>

/*jslint browser: true, devel: true, passfail: true */

// have to pre-declare some globals to pass jslint
/*global jQuery */

`)
	showHumanIntervalJS(w)
	fmt.Fprint(w, `

function updateTimeRemainings() {
    "use strict";
    var i, timeCode, evttim, curdat, curtim, interval, seconds, intText;
    i = 0;
    while (document.getElementById("timerem_code_" + i)) {
        timeCode = document.getElementById("timerem_code_" + i).value;
        evttim = Date.parse(timeCode) + `+timeZoneClientSideAdjustNum+`;
        curdat = new Date();
        curtim = curdat.getTime();
        interval = evttim - curtim;
        seconds = Math.floor(interval / 1000);
        // intText = years + " years " + weeks + " weeks " + days + " days " + hours + " hours " + minutes + " min " + seconds + " sec";
        intText = humanInterval(seconds);
        if (intText.substring(0, 1) === "-") {
            document.getElementById("timerem_disp_" + i).innerHTML = '<font color="red">' + intText + '</font>';
        } else {
            document.getElementById("timerem_disp_" + i).innerHTML = intText;
        }
        i = i + 1;
    }
    window.setTimeout(updateTimeRemainings, 1000);
}

jQuery(function () {
    "use strict";
    updateTimeRemainings();
});

</script>
</head>
<body>
  <section>
`)
	showCalcronMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>List of Calcron Entries</h1>
`)
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	_, showLatest := getform["showlatest"]
	_, showFull := getform["full"]
	_, showDefunct := getform["defunct"]
	_, showAll := getform["all"]
	recalculateAllEvents(db, userid, showDefunct)
	isDefunct := 0 // MySQL needs a numerical 0 or 1
	if showDefunct {
		isDefunct = 1
	}
	if showLatest {
		sql = "SELECT id_cal_ent, title, description, year, month, dom, dow, nth, doe, hour, minute, second, currenttime FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?) ORDER BY id_cal_ent DESC"
	} else {
		sql = "SELECT id_cal_ent, title, description, year, month, dom, dow, nth, doe, hour, minute, second, currenttime FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?) ORDER BY currenttime"
	}
	if !showAll {
		sql = sql + " LIMIT 32;"
	}
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid, isDefunct)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	started := false
	crossedNow := false
	var lastTime uint64
	lastTime = 0
	var prevDay int
	prevDay = 0
	colSpan := "3"
	if showFull {
		colSpan = "11"
	}
	count := 0
	for _, row := range rows {
		entry.idCalEnt = row.Uint64(0)
		entry.title = row.Str(1)
		entry.description = row.Str(2)
		entry.year = row.Str(3)
		entry.month = row.Str(4)
		entry.dom = row.Str(5)
		entry.dow = row.Str(6)
		entry.nth = row.Str(7)
		entry.doe = row.Str(8)
		entry.hour = row.Str(9)
		entry.minute = row.Str(10)
		entry.second = row.Str(11)
		entry.currenttime = row.Uint64(12)

		if !started {
			fmt.Fprint(w, `<form><table border="0" cellpadding="6">
<tr>
<th> Next occurrence </th><th> Time Remaining </th><th> Title </th>
`)
			if showFull {
				fmt.Fprint(w, `<th> Year </th><th> Month </th><th> Day of Month </th><th> Day of Week </th><th> Nth Day of Week/Month </th><th> Day of eternity </th><th> Hour </th><th> Minute </th><th> Second </th>`)
			}
			fmt.Fprint(w, `</tr>
`)
			started = true
		}
		to := time.Unix(int64(uint64(int64(entry.currenttime)+timeZoneOffset)), 0)
		currentDay := to.YearDay()
		if crossedNow {
			if prevDay != currentDay {
				fmt.Fprint(w, `<tr><td colspan="`+colSpan+`" align="center"> <b>`+to.Weekday().String()+`</b> </td></tr>
`)
			}
		}
		prevDay = currentDay
		if entry.currenttime > rightNow {
			if !crossedNow {
				fmt.Fprint(w, `<tr><td colspan="`+colSpan+`" align="center" style="background-color: #ff8080;"> &lt; - - Now - - &gt; </td></tr>
`)
				crossedNow = true
			}
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		if entry.currenttime == lastTime {
			backgroundColor = " style=\"background-color: #FF8000;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+timeCodeToString(entry.currenttime, timeZoneOffset)+" </td>")
		// if crossedNow {
		fmt.Fprint(w, `<td align="right"> <input type="hidden" id="timerem_code_`+strconv.FormatInt(int64(count), 10)+`" value="`)
		timeNums := timeCodeToTimeNumbers(uint64(int64(entry.currenttime) + timeZoneOffset))
		fmt.Fprintf(w, "%d-%s-%sT%s:%s:%s.000Z", timeNums.year, twodigits(timeNums.month), twodigits(timeNums.day), twodigits(timeNums.hour), twodigits(timeNums.minute), twodigits(timeNums.second))
		fmt.Fprint(w, `" /><span id="timerem_disp_`+strconv.FormatInt(int64(count), 10)+`"></span> </td>`)
		count++
		// } else {
		// fmt.Fprint(w, "<td> &nbsp; </td>")
		// }
		theTitle := entry.title
		theDescription := entry.description
		if entry.title == "PRIORITY" {
			theTitle, theDescription = getPriorityTitleAndDescriptionForDate(db, userid, entry.currenttime)
		}
		fmt.Fprint(w, "<td> <a href=\"view?entry="+strconv.FormatUint(entry.idCalEnt, 10)+"\">"+html.EscapeString(theTitle)+"</a> </td>")
		if showFull {
			fmt.Fprint(w, "<td> "+html.EscapeString(entry.year)+" </td><td> "+html.EscapeString(entry.month)+" </td><td> "+html.EscapeString(entry.dom)+" </td><td> "+html.EscapeString(entry.dow)+" </td><td> "+html.EscapeString(entry.nth)+" </td><td> "+html.EscapeString(entry.doe)+" </td><td> "+html.EscapeString(entry.hour)+" </td><td> "+html.EscapeString(entry.minute)+" </td><td> "+html.EscapeString(entry.second)+" </td>")
		}
		fmt.Fprint(w, `</tr>
`)
		if showFull {
			fmt.Fprint(w, `<tr><td colspan="`+colSpan+`"> `+theDescription+`</td></tr>
`)
		}
		lastTime = entry.currenttime
	}
	if started {
		fmt.Fprint(w, "</table></form>")
	}

	fmt.Fprint(w, `
        <p><hr /></p>
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
		doe      string
		hour     string
		minute   string
		second   string
	}
	sql := "SELECT COUNT(*) FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?) AND (is_dismissed = 1);"
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

	sql = "SELECT id_cal_ent, year, month, dom, dow, nth, doe, hour, minute, second FROM calcron_entry WHERE (id_user = ?) AND (is_defunct = ?) AND (is_dismissed = 1) ORDER BY id_cal_ent;"
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
		entry.doe = row.Str(6)
		entry.hour = row.Str(7)
		entry.minute = row.Str(8)
		entry.second = row.Str(9)
		currenttime = calculateNextCurrentTimeForEvent(rightNowWithTimeZone, entry.year, entry.month, entry.dom, entry.dow, entry.nth, entry.doe, entry.hour, entry.minute, entry.second)
		currenttime = uint64(int64(currenttime) - timeZoneOffset)
		updateList[count] = updateItem{entryid: entry.idCalEnt, currenttime: currenttime}
		count++
	}
	var save struct {
		idCalEnt    uint64
		currenttime uint64
		isDefunct   int
		isDismissed int
	}
	cutoff := rightNowWithTimeZone - 259200 // 3 days
	rightNow := uint64(time.Now().Unix())
	for i := 0; i < count; i++ {
		stmt, err := db.Prepare("UPDATE calcron_entry SET currenttime = ?, is_defunct = ?, is_dismissed = ? WHERE id_cal_ent = ?;")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		save.idCalEnt = updateList[i].entryid
		save.currenttime = updateList[i].currenttime
		save.isDefunct = 0
		save.isDismissed = 1
		if save.currenttime < cutoff {
			save.isDefunct = 1
		}
		if save.currenttime >= rightNow {
			save.isDismissed = 0
		}
		stmt.Bind(save.currenttime, save.isDefunct, save.isDismissed, save.idCalEnt)
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

func showChimesPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	db := accessdb.GetDbConnection()
	recalculateAllEvents(db, userid, false)
	var entry struct {
		id          uint64
		title       string
		description string
		currenttime uint64
	}
	// rightNow := uint64(time.Now().Unix())
	// sql := "SELECT id_cal_ent, title, description, currenttime FROM calcron_entry WHERE currenttime >= ? ORDER BY currenttime LIMIT 1;"
	// sql := "SELECT id_cal_ent, title, description, currenttime FROM calcron_entry WHERE is_dismissed = 0 ORDER BY currenttime LIMIT 1;"
	sql := "SELECT id_cal_ent, title, description, currenttime FROM calcron_entry WHERE (id_user = ?) AND (is_dismissed = 0) ORDER BY currenttime LIMIT 1;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(userid) // rightNow not used any more
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	for _, row := range rows {
		entry.id = row.Uint64(0)
		entry.title = row.Str(1)
		entry.description = row.Str(2)
		entry.currenttime = row.Uint64(3)
	}
	theTitle := entry.title
	theDescription := entry.description
	if entry.title == "PRIORITY" {
		theTitle, theDescription = getPriorityTitleAndDescriptionForDate(db, userid, entry.currenttime)
	}
	timeZoneOffset := getTimeZoneOffset(db, userid)
	timeNums := timeCodeToTimeNumbers(uint64(int64(entry.currenttime) + timeZoneOffset))
	dateTimeString := timeCodeToString(entry.currenttime, timeZoneOffset)
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	timeZoneClientSideAdjustNum := strconv.FormatInt((-timeZoneOffset)*1000, 10)
	fmt.Fprint(w, `<title>Chimes: `+html.EscapeString(theTitle)+`</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>

<script>

// (C) 2017 Wayne Radinsky

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
        // gChimesData.timerGoing = false;
        amplitude = 0.9;
        frequency = Math.exp((gChimesData.centernoteLog + 9) * gChimesData.LOG2);
        startTime = 0;
        duration = 0.1;
        fromMoment = gChimesData.globalCtx.currentTime;
        chimesQueUpANote(fromMoment, frequency, startTime, duration, amplitude);
        xl = Math.log(Math.abs(interval)) / gChimesData.LOG2;
        xl = xl * 5;
        // document.getElementById("xl").value = xl;
        xl = Math.floor(xl);
    } else {
        xl = Math.log(interval) / gChimesData.LOG2;
        pitch = 9.0 - (xl / 4.7);
        xl = xl * 5;
        // document.getElementById("xl").value = xl;
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
    // thetime = Date.parse(datestr) + `+timeZoneClientSideAdjustNum+`;
    thetime = chimesConvertOurStringToTimeCode(datestr) + `+timeZoneClientSideAdjustNum+`;
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
    curtim.setTime(curtim.getTime() - `+timeZoneClientSideAdjustNum+`);
    document.getElementById("next_event").value = curtim.getTime();
    // document.getElementById("as_text").value = curtim.toISOString();
    document.getElementById("as_text").value = chimesConvertTimeCodeToOurString(curtim);
    document.getElementById("title").innerHTML = "";
    document.getElementById("description").innerHTML = "";
    document.getElementById("do_dis").style.display = "none";
}

// main

gChimesData = {SAMPLERATE: 44100, TAU: Math.PI * 2, LOG2: Math.log(2), globalCtx: chimesGetAudioContext(), cachedNotes: {}, timerGoing: false, lastXl: 0, lastMoment: 0, lastSecs: 0 };

jQuery(function () {
    "use strict";
    // var curtim;
    // curtim = new Date(); // year, month [, day, hour, minute, second, millisecond]);
    // curtim.setTime(curtim.getTime() - `+timeZoneClientSideAdjustNum+`);
    // document.getElementById("next_event").value = curtim.getTime();
    // document.getElementById("as_text").value = curtim.toISOString();
    chimesExecParseAndSet();
});

</script>

</head>
<body>
  <section>
`)
	showCalcronMenuBar(w, userName)
	fmt.Fprint(w, `
<form action="dismiss" method="get">

<p> Next event: `)
	fmt.Fprint(w, html.EscapeString(dateTimeString))
	fmt.Fprint(w, `</p>
    <h1 id="title"><b>`)
	fmt.Fprint(w, html.EscapeString(theTitle))
	fmt.Fprint(w, `</b></h1>
<p id="description">
`)
	fmt.Fprint(w, html.EscapeString(theDescription))
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
    <input type="textbox" id="interval" name="interval" value="" style="width:400px;" />
</p><p style="display:none;">
    <input type="textbox" id="xl" name="xl" value="" />
</p><p>
    <input type="hidden" id="dismiss" name="dismiss" value="`+strconv.FormatUint(entry.id, 10)+`" />
    <input type="submit" id="do_dis" name="do_dis" value="Dismiss" />
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

func showViewPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var ui struct {
		title       string
		description string
		year        string
		month       string
		dom         string
		dow         string
		nth         string
		doe         string
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
	ui.doe = "*"
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
		db := accessdb.GetDbConnection()
		defer db.Close()
		sql := "SELECT title, description, year, month, dom, dow, nth, doe, hour, minute, second FROM calcron_entry WHERE (id_cal_ent = ?) AND (id_user = ?);"
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
			ui.doe = row.Str(7)
			ui.hour = row.Str(8)
			ui.minute = row.Str(9)
			ui.second = row.Str(10)
		}
	}

	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	db := accessdb.GetDbConnection()
	defer db.Close()
	fmt.Fprint(w, `<title>CalCron Entry View</title>
<link rel="stylesheet" type="text/css" href="/style.css">

</head>
<body>
  <section>
`)
	showCalcronMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>CalCron Entry View</h1>

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
<tr><td align="right"> Day of eternity: </td><td> `+html.EscapeString(ui.doe)+` </td></tr>
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

func showRecalcPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	db := accessdb.GetDbConnection()
	recalculateAllEvents(db, userid, false)
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Chimes: recalc</title>
<link rel="stylesheet" type="text/css" href="/style.css">

</head>
<body>
  <section>
`)
	showCalcronMenuBar(w, userName)
	fmt.Fprint(w, `
<form>

    <h1 id="title">Recalc</h1>
    <p>Recalculated.</p>

  </section>

</body>
</html>`)
}

func doDismiss(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	db := accessdb.GetDbConnection()
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("parseform failed")
	}
	getform := r.Form
	_, dismissExists := getform["dismiss"]
	var dismissId uint64
	if dismissExists {
		dismissId, err = strconv.ParseUint(getform["dismiss"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
	}
	if dismissId > 0 {
		stmt, err := db.Prepare("UPDATE calcron_entry SET is_dismissed = 1 WHERE (id_cal_ent = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(dismissId, userid)
		_, _, err = stmt.Exec()
	}
	http.Redirect(w, r, "chimes", 302)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	testCheckTimeFieldSyntaxError()
	testApplyRule()
	switch {
	case op == "list":
		if userid != 0 {
			showListPage(w, r, op, userid, userName)
		}
	case op == "edit":
		if userid != 0 {
			showEditPage(w, r, op, userid, userName)
		}
	case op == "add":
		if userid != 0 {
			showEditPage(w, r, op, userid, userName)
		}
	case op == "chimes":
		if userid != 0 {
			showChimesPage(w, r, op, userid, userName)
		}
	case op == "view":
		if userid != 0 {
			showViewPage(w, r, op, userid, userName)
		}
	case op == "recalc":
		if userid != 0 {
			showRecalcPage(w, r, op, userid, userName)
		}
	case op == "dismiss":
		if userid != 0 {
			doDismiss(w, r, op, userid, userName)
		}
	default:
		// fmt.Fprintln(w, "Could not find page:", op)
		filename := "/home/ec2-user/wayneserver/staticappcontent/calcron/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
