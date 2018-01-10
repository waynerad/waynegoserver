package streak

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
<meta charset=utf-8 />
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

func showStreakMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<p><a href="list">List</a>
<a href="list?edit=1">Edit</a>
<a href="add">Add</a>
&middot; `+htm(userName)+`
</p>
`)
}

func showEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		name   string
		title string
	}
	var taskid uint64
	taskid = 0
	if method == "GET" {
		// set defaults
		ui.name = ""
		ui.title = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, ok := getform["task"]
		if ok {
			taskid = strToUint(getform["task"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT name, title FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(taskid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.name = row.Str(0)
				ui.title = row.Str(1)
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
		taskid = strToUint(postform["task"][0])
		ui.name = strings.Trim(postform["name"][0], " \r\n\t")
		ui.title = strings.Trim(postform["title"][0], " \r\n\t")
		if ui.name == "" {
			errorList["name"] = "Name is blank"
			errorOccurred = true
		}
		if ui.title == "" {
			errorList["title"] = "Target is blank"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			// dbConnect!!
			db := accessdb.GetDbConnection()
			defer db.Close()
			var save struct {
				idTask   uint64
				idUser       uint64
				title       string
			}
			save.idTask = taskid
			save.idUser = userid
			save.title = ui.title
			save.lastAccessed = uint64(time.Now().Unix())
			// query, if there, update, if not, create new
			alreadyExists := false
			if taskid > 0 {
				sql := "SELECT id_task FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
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
				for _, _ = range rows {
					alreadyExists = true
				}
			}
			if alreadyExists {
				stmt, err := db.Prepare("UPDATE streak_task SET title = ? WHERE (id_task = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.title, save.taskid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO streak_task (id_user, title) VALUES (?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idUser, save.title)
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
		fmt.Fprint(w, `<title>Streak Task</title>
<link rel="stylesheet" type="text/css" href="/bookmark/style.css">
</head>
<body onload="document.getElementById('name').focus();">
  <section>
`)
		showStreakMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Streak Task</h1>
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
<tr><td align="right"> Name: </td><td> <input type="hidden" name="task" value="`+uintToStr(taskid)+`"/><input size="40" name="name" id="name" type="text" value="`+html.EscapeString(ui.name)+`" /> </td></tr>
<tr><td align="right"> Target: </td><td> <input size="40" name="title" id="title" type="text" value="`+html.EscapeString(ui.title)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
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
		idTask uint64
		title     string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Streak Tasks</title>
<link rel="stylesheet" type="text/css" href="/bookmark/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
</head>
<body>
  <section>
`)
	showStreakMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>List of Streak Tasks</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	sql = "SELECT id_task, title FROM streak_task WHERE id_user = ? ORDER BY id_user, last_accessed DESC LIMIT 0, ?;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	number := 1000
	getform := r.Form
	_, ok := getform["number"]
	if ok {
		number = strToInt(getform["number"][0])
	}
	showtitles := false
	_, showtitles = getform["titles"]
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
		entry.idTask = row.Uint64(0)
		entry.titles = row.Str(2)
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
		fmt.Fprint(w, `<td> <b><a href="evoke?bookmark=`+uintToStr(entry.idTask)+`" title="`+htm(entry.title)+`">`+html.EscapeString(entry.name)+`</a></b>`)
		if showedits {
			fmt.Fprint(w, ` &middot; <font size="1"><a href="edit?bookmark=`+uintToStr(entry.idTask)+`">Edit</a> &middot; <a href="delete?bookmark=`+uintToStr(entry.idTask)+`">Delete</a>`)
		}
		if showtitles {
			fmt.Fprint(w, ` &middot; `+htm(entry.title))
		}
		fmt.Fprint(w, `</font> </td>`)
		// fmt.Fprint(w, "<td> "+html.EscapeString(entry.name)+" </td><td> "+html.EscapeString(entry.title)+" </td>")
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, "</table></form>")
	}
	if count == number {
		number *= 10
		fmt.Fprint(w, `<p><a href="list?number=`+intToStr(number)+`">More</a></p>`)
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func showDeletePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		title string
	}
	var taskid uint64
	taskid = 0
	if method == "GET" {
		// set defaults
		ui.title = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, ok := getform["task"]
		if ok {
			taskid = strToUint(getform["task"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT title FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(taskid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.title = row.Str(1)
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
		taskid = strToUint(postform["task"][0])

		// dbConnect!!
		db := accessdb.GetDbConnection()
		defer db.Close()
		// query, if there, update, if not, create new
		stmt, err := db.Prepare("DELETE FROM streak_task WHERE (id_task = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(taskid, userid)
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
		fmt.Fprint(w, getDoctype())
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>Delete Streak Task</title>
<link rel="stylesheet" type="text/css" href="/bookmark/style.css">
</head>
<body>
  <section>
`)
		showStreakMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Delete Streak Task</h1>
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
<p>Title: `+htm(ui.title)+`</p>
<p><input type="hidden" name="task" value="`+uintToStr(taskid)+`"/>
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
		filename := "/home/ec2-user/wayneserver/staticappcontent/streak/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
