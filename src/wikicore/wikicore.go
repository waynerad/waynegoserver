package wikicore

// CREATE TABLE wiki_page (
//   id_page BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
//   id_user BIGINT UNSIGNED NOT NULL DEFAULT 0,
//   title VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
//   content TEXT COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
// PRIMARY KEY(id_page),
// KEY (title)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

import (
	"fmt"
	"forms"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"login"
	"net/http"
	"net/url"
	"static"
	"strconv"
	"strings"
	"wikitypes"
	"wikiui"
	"wutil"
)

func htmlize(stg string) string {
	return html.EscapeString(stg)
}

func uint64ToStr(ii uint64) string {
	return strconv.FormatUint(ii, 10)
}

func intToStr(ii int) string {
	return strconv.FormatInt(int64(ii), 10)
}

func strToUint64(z string) uint64 {
	rv, err := strconv.ParseUint(z, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("Parse Uint failed")
	}
	return rv
}

func trim(z string) string {
	return strings.Trim(z, " \r\n\t")
}

////////////////////////////////////////////////////////////////
// New Form object system
////////////////////////////////////////////////////////////////

// ----------------------------------------------------------------
// Wiki Page List page
// ----------------------------------------------------------------

func getWikiPageListDBData(db mysql.Conn, userInfo *login.UserInformationRecord) wikitypes.WikiPageListData {
	var currentEntry wikitypes.WikiPageDisplayData
	theList := make(wikitypes.WikiPageListData, 0)
	sql := "SELECT id_page, title, content FROM wiki_page WHERE (id_user = ?) ORDER BY title;"
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
		currentEntry.IdPage = row.Uint64(0)
		currentEntry.Title = row.Str(1)
		currentEntry.Content = row.Str(2)
		theList = append(theList, currentEntry)
	}
	return theList
}

func showWikiPageListPage(w http.ResponseWriter, userInfo *login.UserInformationRecord, userInput map[string]string, dbDataList wikitypes.WikiPageListData) {
	_, editmode := userInput["edit"]
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Wiki Page List"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	wikiui.ShowWikiHeadHeader(w, displayInfo)
	wikiui.ShowWikiBodyHeader(w, displayInfo)
	wikiui.ShowWikiPageList(w, dbDataList, editmode)
	wikiui.ShowWikiFooter(w, displayInfo)
}

type wikiPageListForm struct {
	objectName string
}

func (self *wikiPageListForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	_, editmode := userInput["edit"]
	if editmode {
		rv := make(map[string]string)
		rv["edit"] = "1"
		return rv
	}
	return nil
}

func (self *wikiPageListForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	dbDataList := getWikiPageListDBData(db, userInfo)
	showWikiPageListPage(w, userInfo, userInput, dbDataList)
}

func (self *wikiPageListForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	return nil, nil
}

func (self *wikiPageListForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	return nil
}

// ----------------------------------------------------------------
// End of wiki page list page
// ----------------------------------------------------------------

type wikiPageEntryData struct {
	idPage   uint64
	title    string
	contents string
}

func pageIdFromTitle(db mysql.Conn, userInfo *login.UserInformationRecord, title string) uint64 {
	sql := "SELECT id_page FROM wiki_page WHERE (title = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(title, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var idPage uint64
	idPage = 0
	for _, row := range rows {
		idPage = row.Uint64(0)
	}
	return idPage
}

func replaceTLookup(db mysql.Conn, userInfo *login.UserInformationRecord, content string, idx int) (string, int) {
	eocmd := wutil.FindInString(content, "]]", idx) + 2
	cmd := trim(content[idx:eocmd])
	splt := strings.Split(cmd[:len(cmd)-2], ":")
	var target string
	var linkTitle string
	var displayTitle string
	var replacement string
	if len(splt) == 2 {
		linkTitle = splt[1]
		displayTitle = splt[1]
	} else {
		linkTitle = splt[1]
		displayTitle = splt[2]
	}
	idPage := pageIdFromTitle(db, userInfo, linkTitle)
	if idPage == 0 {
		target = "wikipageedit?page=" + uint64ToStr(idPage) + "&prefil=" + url.QueryEscape(linkTitle)
		replacement = `<a href="` + target + `" style="color: red;">` + htmlize(displayTitle) + "</a>"
	} else {
		target = "wikipageshow?page=" + uint64ToStr(idPage)
		replacement = `<a href="` + target + `" style="color: green;">` + htmlize(displayTitle) + "</a>"
	}
	additional := len(replacement) - len(cmd)
	return content[:idx] + replacement + content[eocmd:], eocmd + additional
}

func replaceHLookup(db mysql.Conn, userInfo *login.UserInformationRecord, content string, idx int) (string, int) {
	eocmd := wutil.FindInString(content, "]]", idx) + 2
	cmd := trim(content[idx:eocmd])
	splt := strings.Split(cmd[:len(cmd)-2], ":")
	var target string
	var displayTitle string
	var strPage string
	if len(splt) == 2 {
		strPage = splt[1]
		displayTitle = ""
	} else {
		strPage = splt[1]
		displayTitle = splt[2]
	}
	var idPage uint64
	idPage, err := strconv.ParseUint(strPage, 10, 64)
	if err != nil {
		idPage = 0
	}
	if displayTitle == "" {
		sql := "SELECT id_page, title FROM wiki_page WHERE (id_page = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			panic("Prepare failed")
		}
		sel.Bind(idPage, userInfo.UserId)
		rows, _, err := sel.Exec()
		if err != nil {
			panic("Bind/Exec failed")
		}
		displayTitle = "[mising]"
		for _, row := range rows {
			displayTitle = row.Str(1)
		}
	}
	target = "wikipageshow?page=" + uint64ToStr(idPage)
	replacement := `<a href="` + target + `">` + htmlize(displayTitle) + "</a>"
	additional := len(replacement) - len(cmd) // should be negative
	return content[:idx] + replacement + content[eocmd:], eocmd + additional
}

func commandT(db mysql.Conn, userInfo *login.UserInformationRecord, content string) string {
	idx := wutil.FindInString(content, "[[T:", -1)
	for idx >= 0 {
		content, idx = replaceTLookup(db, userInfo, content, idx)
		idx = wutil.FindInString(content, "[[T:", idx)
	}
	return content
}

func commandH(db mysql.Conn, userInfo *login.UserInformationRecord, content string) string {
	idx := wutil.FindInString(content, "[[H:", -1)
	for idx >= 0 {
		content, idx = replaceHLookup(db, userInfo, content, idx)
		idx = wutil.FindInString(content, "[[H:", idx)
	}
	return content
}

func commandize(db mysql.Conn, userInfo *login.UserInformationRecord, content string) string {
	idx := strings.Index(content, "[[T:")
	if idx >= 0 {
		content = commandT(db, userInfo, content)
	}
	idx = strings.Index(content, "[[H:")
	if idx >= 0 {
		content = commandH(db, userInfo, content)
	}
	return content
}

func replaceTWithHByLookup(db mysql.Conn, userInfo *login.UserInformationRecord, content string, idx int) (string, int) {
	eocmd := wutil.FindInString(content, "]]", idx) + 2
	cmd := trim(content[idx:eocmd])
	splt := strings.Split(cmd[:len(cmd)-2], ":")
	// var target string
	var linkTitle string
	var displayTitle string
	var replacement string
	if len(splt) == 2 {
		linkTitle = splt[1]
		displayTitle = ""
	} else {
		linkTitle = splt[1]
		displayTitle = splt[2]
	}
	idPage := pageIdFromTitle(db, userInfo, linkTitle)
	if idPage == 0 {
		return content, eocmd
	} else {
		if displayTitle == "" {
			replacement = "[[H:" + uint64ToStr(idPage) + "]]"
		} else {
			replacement = "[[H:" + uint64ToStr(idPage) + ":" + displayTitle + "]]"
		}
	}
	additional := len(replacement) - len(cmd)
	return content[:idx] + replacement + content[eocmd:], eocmd + additional
}

func replaceCommandTWithCommandH(db mysql.Conn, userInfo *login.UserInformationRecord, content string) string {
	idx := wutil.FindInString(content, "[[T:", -1)
	for idx >= 0 {
		content, idx = replaceTWithHByLookup(db, userInfo, content, idx)
		idx = wutil.FindInString(content, "[[T:", idx)
	}
	return content
}

// ----------------------------------------------------------------
// Wiki Page Edit Page
// ----------------------------------------------------------------

type wikiPageEditForm struct {
	objectName string
}

func (self *wikiPageEditForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	var currentEntry wikiPageEntryData
	rv := make(map[string]string)
	sql := "SELECT id_page, title, content FROM wiki_page WHERE (id_page = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	wikiPageId := strToUint64(userInput["page"])
	sel.Bind(wikiPageId, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	found := false
	for _, row := range rows {
		currentEntry.idPage = row.Uint64(0)
		currentEntry.title = row.Str(1)
		currentEntry.contents = row.Str(2)
		rv["page"] = uint64ToStr(currentEntry.idPage)
		rv["title"] = currentEntry.title
		rv["content"] = currentEntry.contents
		found = true
	}
	if !found {
		rv["page"] = uint64ToStr(wikiPageId)
		rv["title"] = userInput["prefil"]
		rv["content"] = ""
	}
	return rv
}

func (self *wikiPageEditForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Edit Wiki Page"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	wikiui.ShowWikiHeadHeader(w, displayInfo)
	wikiui.ShowWikiBodyHeader(w, displayInfo)
	wikiui.ShowWikiPageEditForm(w, errorList, userInput, displayInfo)
	wikiui.ShowWikiFooter(w, displayInfo)
}

func (self *wikiPageEditForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	errorList := make(map[string]string)
	if userInput["title"] == "" {
		errorList["title"] = "Title cannot be left blank."
	}
	if userInput["content"] == "" {
		errorList["content"] = "Content cannot be left blank."
	}
	return errorList, nil
}

func (self *wikiPageEditForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	var wikiPageSaveRecord struct {
		pageId  uint64
		title   string
		content string
	}
	wikiPageSaveRecord.pageId = strToUint64(userInput["page"])
	wikiPageSaveRecord.title = trim(userInput["title"])
	wikiPageSaveRecord.content = replaceCommandTWithCommandH(db, userInfo, trim(userInput["content"]))

	var err error
	alreadyExists := false
	if wikiPageSaveRecord.pageId > 0 {
		sql := "SELECT id_page FROM wiki_page WHERE (id_page = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(wikiPageSaveRecord.pageId, userInfo.UserId)
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
		stmt, err := db.Prepare("UPDATE wiki_page SET title = ?, content = ?  WHERE (id_page = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(wikiPageSaveRecord.title, wikiPageSaveRecord.content, wikiPageSaveRecord.pageId, userInfo.UserId)
		_, _, err = stmt.Exec()
	} else {
		stmt, err := db.Prepare("INSERT INTO wiki_page (id_user, title, content) VALUES (?, ?, ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		stmt.Bind(userInfo.UserId, wikiPageSaveRecord.title, wikiPageSaveRecord.content)
		_, _, err = stmt.Exec()
	}
	if err != nil {
		fmt.Println(err)
		panic("Exec failed")
	}
	return nil
}

// ----------------------------------------------------------------
// End wiki page edit page
// ----------------------------------------------------------------

func getDisplayDataForWikiPageShowPage(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	var currentEntry wikiPageEntryData
	rv := make(map[string]string)
	sql := "SELECT id_page, title, content FROM wiki_page WHERE (id_page = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	wikiPageId := strToUint64(userInput["page"])
	sel.Bind(wikiPageId, userInfo.UserId)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	found := false
	for _, row := range rows {
		currentEntry.idPage = row.Uint64(0)
		currentEntry.title = row.Str(1)
		currentEntry.contents = row.Str(2)
		rv["page"] = uint64ToStr(currentEntry.idPage)
		rv["title"] = currentEntry.title
		rv["content"] = commandize(db, userInfo, wutil.Markdownize(htmlize(currentEntry.contents)))
		found = true
	}
	if !found {
		rv["page"] = uint64ToStr(wikiPageId)
		rv["title"] = ""
		rv["content"] = ""
	}
	return rv
}

// ----------------------------------------------------------------
// Wiki Page Show Page
// ----------------------------------------------------------------

type wikiPageShowForm struct {
	objectName string
}

func (self *wikiPageShowForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	rv := make(map[string]string)
	wikiPageId := strToUint64(userInput["page"])
	rv["page"] = uint64ToStr(wikiPageId)
	return rv
}

func (self *wikiPageShowForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	displayInfo := getDisplayDataForWikiPageShowPage(db, userInfo, userInput)
	// make(map[string]string)
	displayInfo["hTitle"] = "Wiki Page:"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	wikiui.ShowWikiHeadHeader(w, displayInfo)
	wikiui.ShowWikiBodyHeader(w, displayInfo)
	wikiui.ShowWikiPageShowForm(w, errorList, userInput, displayInfo)
	wikiui.ShowWikiFooter(w, displayInfo)
}

func (self *wikiPageShowForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	return nil, nil
}

func (self *wikiPageShowForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	return nil
}

// ----------------------------------------------------------------
// End wiki page show page
// ----------------------------------------------------------------

func doWikiPageListPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject wikiPageListForm
	formObject.objectName = "Wiki Page List Form"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "http://www.yahoo.com/")
}

func doWikiPageEditPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject wikiPageEditForm
	formObject.objectName = "Wiki Page Edit Form"
	// we have to parse here to get the page ID because we need to know where to redirect to afterwards already
	// form parsing is idempotent so it's no harm to be done twice
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("parseform failed")
	}
	pageStr := r.Form["page"][0]
	pageId := strToUint64(pageStr)
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "wikipageshow?page="+uint64ToStr(pageId))
}

func doWikiPageShowPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject wikiPageShowForm
	formObject.objectName = "Wiki Page Display"
	forms.HandleStandaloneForm(&formObject, w, r, op, userInfo, "http://www.yahoo.com/")
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	fmt.Println("op is", op)
	switch {
	case op == "wikipagelist":
		if userInfo.UserId != 0 {
			doWikiPageListPage(w, r, op, userInfo)
		}
	case op == "wikipageedit":
		if userInfo.UserId != 0 {
			doWikiPageEditPage(w, r, op, userInfo)
		}
	case op == "wikipageshow":
		if userInfo.UserId != 0 {
			doWikiPageShowPage(w, r, op, userInfo)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/wiki/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
