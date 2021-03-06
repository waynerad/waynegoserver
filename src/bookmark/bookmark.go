package bookmark

import (
	"accessdb"
	"fmt"
	"html"
	"net/http"
	"static"
	"strconv"
	"strings"
	"time"
)

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<meta http-equiv="X-UA-Compatible" content="ie=edge" />
`
}

func getStyle() string {
	return `<style>
body {
    font-size: 1.1em;
    font-family: helvetica;
}
#header {
    background-color: #FFEFE0;
}
#footer {
    background-color: #FFEFE0;
}

h1 {
    color: #550000;
}

.infield {
    font-size: 1.1em;
}

</style>

`
}

func uintToStr(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func intToStr(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

func strToUint(s string) uint64 {
	num, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("ParseUint failed")
	}
	return num
}

func strToInt(s string) int {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("ParseUint failed")
	}
	return int(num)
}

func htm(str string) string {
	return html.EscapeString(str)
}

func showBookmrkMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<div id="header">
<p><a href="list">List</a>
<a href="list?edit=1">Edit</a>
<a href="add">Add</a>
&middot; `+htm(userName)+`
</p>
</div>
`)
}

func showEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		title  string
		target string
	}
	var bookmarkid uint64
	bookmarkid = 0
	if method == "GET" {
		// set defaults
		ui.title = ""
		ui.target = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, ok := getform["bookmark"]
		if ok {
			bookmarkid = strToUint(getform["bookmark"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT title, target FROM bookmark_link WHERE (id_bookmark = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(bookmarkid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.title = row.Str(0)
				ui.target = row.Str(1)
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
		bookmarkid = strToUint(postform["bookmark"][0])
		ui.title = strings.Trim(postform["title"][0], " \r\n\t")
		ui.target = strings.Trim(postform["target"][0], " \r\n\t")
		if ui.title == "" {
			errorList["title"] = "Name is blank"
			errorOccurred = true
		}
		if ui.target == "" {
			errorList["target"] = "Target is blank"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			// dbConnect!!
			db := accessdb.GetDbConnection()
			defer db.Close()
			var save struct {
				idBookmark   uint64
				idUser       uint64
				title        string
				target       string
				lastAccessed uint64
			}
			save.idBookmark = bookmarkid
			save.idUser = userid
			save.title = ui.title
			save.target = ui.target
			save.lastAccessed = uint64(time.Now().Unix())
			// query, if there, update, if not, create new
			alreadyExists := false
			if bookmarkid > 0 {
				sql := "SELECT id_bookmark FROM bookmark_link WHERE (id_bookmark = ?) AND (id_user = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(bookmarkid, userid)
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
				stmt, err := db.Prepare("UPDATE bookmark_link SET title = ?, target = ?, last_accessed = ? WHERE (id_bookmark = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.title, save.target, save.lastAccessed, bookmarkid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO bookmark_link (id_user, title, target, last_accessed) VALUES (?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idUser, save.title, save.target, save.lastAccessed)
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
		fmt.Fprint(w, getDoctype()+getStyle())
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>Bookmark Entry</title>
</head>
<body onload="document.getElementById('target').focus();">
  <section>
`)
		showBookmrkMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Bookmark Entry</h1>
<form action="edit" method="post">
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
<tr><td align="right"> Target: </td><td> <input type="hidden" name="bookmark" value="`+uintToStr(bookmarkid)+`"/><input size="40" name="target" id="target" class="infield" type="text" value="`+html.EscapeString(ui.target)+`" /> </td></tr>
<tr><td align="right"> Title: </td><td> <input size="40" name="title" id="title" class="infield" type="text" value="`+html.EscapeString(ui.title)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input class="infield" type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	var entry struct {
		idBookmark uint64
		title      string
		target     string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	// <link rel="stylesheet" href="jquery-ui.css" />
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>List of Bookmarks</title>
</head>
<body>
  <section>
`)
	showBookmrkMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>List of Bookmarks</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	sql = "SELECT id_bookmark, title, target FROM bookmark_link WHERE id_user = ? ORDER BY id_user, last_accessed DESC LIMIT 0, ?;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	number := 200
	getform := r.Form
	_, ok := getform["number"]
	if ok {
		number = strToInt(getform["number"][0])
	}
	showtargets := false
	_, showtargets = getform["target"]
	showedits := false
	_, showedits = getform["edit"]
	sel.Bind(userid, number)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	started := false
	count := 0
	for _, row := range rows {
		entry.idBookmark = row.Uint64(0)
		entry.title = row.Str(1)
		entry.target = row.Str(2)
		if !started {
			fmt.Fprint(w, `<form><table border="0" cellpadding="2">
<tr>
`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		count++
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+">")
		fmt.Fprint(w, `<td> <b><a href="evoke?bookmark=`+uintToStr(entry.idBookmark)+`" title="`+htm(entry.target)+`">`+html.EscapeString(entry.title)+`</a></b>`)
		if showedits {
			fmt.Fprint(w, ` &middot; <font size="1"><a href="edit?bookmark=`+uintToStr(entry.idBookmark)+`">Edit</a> &middot; <a href="delete?bookmark=`+uintToStr(entry.idBookmark)+`">Delete</a>`)
		}
		if showtargets {
			fmt.Fprint(w, ` &middot; `+htm(entry.target))
		}
		fmt.Fprint(w, `</font> </td>`)
		// fmt.Fprint(w, "<td> "+html.EscapeString(entry.title)+" </td><td> "+html.EscapeString(entry.target)+" </td>")
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, "</table></form>")
	}
	if count == number {
		number *= 10
		fmt.Fprint(w, `<p><a href="list?number=`+intToStr(number)+`">Show More</a></p>`)
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func evoke(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var bookmarkid uint64
	var target string
	bookmarkid = 0
	method := r.Method
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, ok := getform["bookmark"]
		if ok {
			bookmarkid = strToUint(getform["bookmark"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT target FROM bookmark_link WHERE (id_bookmark = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(bookmarkid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				target = row.Str(0)
			}
			var lastAccessed uint64
			lastAccessed = uint64(time.Now().Unix())
			sql = "UPDATE bookmark_link SET last_accessed = ? WHERE (id_bookmark = ?) AND (id_user = ?);"
			sel, err = db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(lastAccessed, bookmarkid, userid)
			rows, _, err = sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			http.Redirect(w, r, target, 302)
		}
	}
}

func showDeletePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		title  string
		target string
	}
	var bookmarkid uint64
	bookmarkid = 0
	if method == "GET" {
		// set defaults
		ui.title = ""
		ui.target = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, ok := getform["bookmark"]
		if ok {
			bookmarkid = strToUint(getform["bookmark"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT title, target FROM bookmark_link WHERE (id_bookmark = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(bookmarkid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.title = row.Str(0)
				ui.target = row.Str(1)
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
		bookmarkid = strToUint(postform["bookmark"][0])

		// dbConnect!!
		db := accessdb.GetDbConnection()
		defer db.Close()
		// query, if there, update, if not, create new
		stmt, err := db.Prepare("DELETE FROM bookmark_link WHERE (id_bookmark = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(bookmarkid, userid)
		_, _, err = stmt.Exec()

		if err != nil {
			fmt.Println(err)
			panic("Exec failed")
		}
		http.Redirect(w, r, "list", 302)

	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype()+getStyle())
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>Delete Bookmark</title>
</head>
<body>
  <section>
`)
		showBookmrkMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Delete Bookmark</h1>
<form action="delete" method="post">
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<p>Do you want to delete:</p>
<p><b>`+htm(ui.title)+`</b><p>
<p>Target: `+htm(ui.target)+`</p>
<p><input type="hidden" name="bookmark" value="`+uintToStr(bookmarkid)+`"/>
<input type="submit" value="Delete">
</p>
</form>
  </section>
</body>
</html>`)
	}
}

func redirectToLoginPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "../login/login", 302)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	switch {
	case op == "list":
		if userid != 0 {
			showListPage(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	case op == "evoke":
		if userid != 0 {
			evoke(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	case op == "edit":
		if userid != 0 {
			showEditPage(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	case op == "add":
		if userid != 0 {
			showEditPage(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	case op == "delete":
		if userid != 0 {
			showDeletePage(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	default:
		// fmt.Fprintln(w, "Could not find page:", op)
		filename := "/home/ec2-user/wayneserver/staticappcontent/bookmark/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
