package fitb

import (
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"net/http"
	"static"
	"strconv"
	"strings"
)

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

func showEditTopicPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		idTopic uint64
		idUser  uint64
		name    string
	}
	var topicid uint64
	topicid = 0
	if method == "GET" {
		// set defaults
		ui.name = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, topicExists := getform["topic"]
		if topicExists {
			topicid, err = strconv.ParseUint(getform["topic"][0], 10, 64)
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
			sql := "SELECT id_topic, name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.idTopic = row.Uint64(0)
				ui.name = row.Str(1)
			}
		}
	}
	if method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		postform := r.Form
		// error checking
		topicid, err = strconv.ParseUint(postform["topic"][0], 10, 64)
		if err != nil {
			topicid = 0
		}
		ui.name = strings.Trim(postform["name"][0], " \r\n\t")
		if ui.name == "" {
			errorList["name"] = "Please specify a name."
		}
		if errorOccurred {
			showform = true
		} else {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			var save struct {
				idTopic uint64
				idUser  uint64
				name    string
			}
			save.idTopic = topicid
			save.idUser = userid
			save.name = ui.name
			alreadyExists := false
			if topicid > 0 {
				sql := "SELECT id_topic FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(topicid, userid)
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
				stmt, err := db.Prepare("UPDATE fitb_topic SET name = ? WHERE (id_topic = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.name, topicid, userid)
				_, _, err = stmt.Exec()
			} else {
				// mysql> DESCRIBE fitb_topic;
				// +----------+------------------+------+-----+---------+----------------+
				// | Field    | Type             | Null | Key | Default | Extra          |
				// +----------+------------------+------+-----+---------+----------------+
				// | id_topic | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
				// | name     | varchar(255)     | NO   |     |         |                |
				// +----------+------------------+------+-----+---------+----------------+
				// 2 rows in set (0.00 sec)
				stmt, err := db.Prepare("INSERT INTO fitb_topic (id_topic, id_user, name) VALUES (?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idTopic, save.idUser, save.name)
				_, _, err = stmt.Exec()
			}
			if err != nil {
				fmt.Println(err)
				panic("Exec failed")
			}
			http.Redirect(w, r, "listtopics", 302)
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
		fmt.Fprint(w, `<title>Topic Entry</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
    <h1>Topic Entry</h1>
<form action="edittopic" method="post">
<input type="hidden" name="topic" value="`+strconv.FormatUint(topicid, 10)+`" />
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
<tr><td align="right"> Name: </td><td> <input class="biginput" name="name" id="name" type="text" value="`+html.EscapeString(ui.name)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func showTopicListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var sql string
	var topic struct {
		idTopic uint64
		name    string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of Topics</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>
</head>
<body>
  <section>
    <h1>List of Topics</h1>
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
	sql = "SELECT id_topic, name FROM fitb_topic WHERE (id_user = ?) ORDER BY id_topic ;"
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
		topic.idTopic = row.Uint64(0)
		topic.name = row.Str(1)
		if !started {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Edit </th><th> Chapters </th><th> Questions </th></tr>`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		topicAsStr := strconv.FormatUint(topic.idTopic, 10)
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+html.EscapeString(topic.name)+" </td><td> " + `<a href="edittopic?topic=`+topicAsStr+`">Edit</a> </td> <td> <a href="listchapters?topic=`+topicAsStr+`">Chapters</a> </td><td> <a href="editquestions?topic=`+topicAsStr+`">Questions</a> </td>`)
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func showTopicPickListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var sql string
	// mysql> DESCRIBE fitb_topic;
	// +----------+------------------+------+-----+---------+----------------+
	// | Field    | Type             | Null | Key | Default | Extra          |
	// +----------+------------------+------+-----+---------+----------------+
	// | id_topic | int(10) unsigned | NO   | PRI | NULL    | auto_increment |
	// | name     | varchar(255)     | NO   |     |         |                |
	// +----------+------------------+------+-----+---------+----------------+
	// 2 rows in set (0.00 sec)
	var topic struct {
		idTopic uint64
		name    string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Pick Topic</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>
</head>
<body>
  <section>
    <h1>Pick Topic</h1>
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
	sql = "SELECT id_topic, name FROM fitb_topic WHERE (id_user = ?) ORDER BY id_topic ;"
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
		topic.idTopic = row.Uint64(0)
		topic.name = row.Str(1)
		if !started {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th></tr>`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"practice?topic=`+strconv.FormatUint(topic.idTopic, 10)+`">`+html.EscapeString(topic.name)+"</a> </td>")
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func showPracticePage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var ui struct {
		idTopic uint64
		idUser  uint64
		name    string
	}
	var topicid uint64
	topicid = 0
	// set defaults
	ui.name = ""
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("parseform failed")
	}
	getform := r.Form
	_, topicExists := getform["topic"]
	if topicExists {
		topicid, err = strconv.ParseUint(getform["topic"][0], 10, 64)
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
		sql := "SELECT id_topic, name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(topicid, userid)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Bind/Exec failed")
		}
		for _, row := range rows {
			ui.idTopic = row.Uint64(0)
			ui.name = row.Str(1)
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
	fmt.Fprint(w, `<title>Practice</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
    <h1>Practice: `+html.EscapeString(ui.name)+`</h1>
  </section>
</body>
</html>`)
}

func showEditChapterPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		idChapter uint64
		idTopic uint64
		idUser  uint64
		name    string
	}
	var chapterid uint64
	chapterid = 0
	if method == "GET" {
		// set defaults
		ui.name = ""
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, chapterExists := getform["chapter"]
		if chapterExists {
			chapterid, err = strconv.ParseUint(getform["chapter"][0], 10, 64)
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
			sql := "SELECT id_chapter, id_topic, name FROM fitb_chapter WHERE (id_chapter = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(chapterid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.idChapter = row.Uint64(0)
				ui.idTopic = row.Uint64(1)
				ui.name = row.Str(2)
			}
		}
	}
	if method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		postform := r.Form
		// error checking
		chapterid, err = strconv.ParseUint(postform["chapter"][0], 10, 64)
		if err != nil {
			chapterid = 0
		}
		ui.name = strings.Trim(postform["name"][0], " \r\n\t")
		if ui.name == "" {
			errorList["name"] = "Please specify a name."
			errorOccurred = true
		}
		ui.idTopic, err = strconv.ParseUint(postform["topic"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		userOwnsTopic := false
		if ui.idTopic > 0 {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			topicid := ui.idTopic
			sql := "SELECT id_topic FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Exec() failed")
			}
			for _, _ = range rows {
				userOwnsTopic = true
			}
			if userOwnsTopic {
				if chapterid > 0 {
					var currentTopic uint64
					currentTopic = 0
					userOwnsTopic = false
					sql := "SELECT id_topic FROM fitb_chapter WHERE id_chapter = ?;"
					sel, err := db.Prepare(sql)
					if err != nil {
						fmt.Println(err)
						panic("Prepare failed")
					}
					sel.Bind(chapterid)
					rows, _, err := sel.Exec()
					if err != nil {
						fmt.Println(err)
						panic("Exec() failed")
					}
					for _, row := range rows {
						currentTopic = row.Uint64(0)
					}
					sql = "SELECT id_topic FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
					sel, err = db.Prepare(sql)
					if err != nil {
						fmt.Println(err)
						panic("Prepare failed")
					}
					sel.Bind(currentTopic, userid)
					rows, _, err = sel.Exec()
					if err != nil {
						fmt.Println(err)
						panic("Exec() failed")
					}
					for _, _ = range rows {
						userOwnsTopic = true
					}
				}
			}
		}
		if (!userOwnsTopic) || (ui.idTopic == 0) {
			errorList["topic"] = "Topic is missing or invalid"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			var save struct {
				idChapter uint64
				idTopic uint64
				name    string
			}
			save.idChapter = chapterid
			save.idTopic = ui.idTopic
			save.name = ui.name
			alreadyExists := false
			if chapterid > 0 {
				sql := "SELECT id_chapter FROM fitb_chapter WHERE (id_chapter = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(chapterid)
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
				stmt, err := db.Prepare("UPDATE fitb_chapter SET id_topic = ?, name = ? WHERE (id_chapter = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.idTopic, save.name, chapterid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO fitb_chapter (id_topic, name) VALUES (?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idTopic, save.name)
				_, _, err = stmt.Exec()
			}
			if err != nil {
				fmt.Println(err)
				panic("Exec failed")
			}
			http.Redirect(w, r, "listchapters?topic="+ strconv.FormatUint(ui.idTopic, 10), 302)
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
		fmt.Fprint(w, `<title>Chapter Entry</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
    <h1>Chapter Entry</h1>
<form action="editchapter" method="post">
<input type="hidden" name="chapter" value="`+strconv.FormatUint(chapterid, 10)+`" />
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
<tr><td align="right"> Topic: </td><td> <input class="biginput" name="topic" id="topic" type="text" value="`+strconv.FormatUint(ui.idTopic, 10)+`" /> </td></tr>
<tr><td align="right"> Name: </td><td> <input class="biginput" name="name" id="name" type="text" value="`+html.EscapeString(ui.name)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func showChapterListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	var sql string
	var chapter struct {
		idChapter uint64
		idTopic uint64
		name    string
	}
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	topicid, err := strconv.ParseUint(getform["topic"][0], 10, 64)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer db.Close()
	sql = "SELECT name FROM fitb_topic WHERE (id_topic = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(topicid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	topicName := "what?"
	for _, row := range rows {
		topicName = row.Str(0)
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Chapters of `+html.EscapeString(topicName)+`</title>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>
</head>
<body>
  <section>
    <h1>Chapters of `+html.EscapeString(topicName)+`</h1>
`)

	sql = "SELECT id_chapter, name FROM fitb_chapter WHERE (id_topic = ?) ORDER BY sequence_num;"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(topicid)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	count := 0
	started := false
	for _, row := range rows {
		chapter.idChapter = row.Uint64(0)
		chapter.idTopic = topicid
		chapter.name = row.Str(1)
		if !started {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th></tr>`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"editchapter?chapter=`+strconv.FormatUint(chapter.idChapter, 10)+`">`+html.EscapeString(chapter.name)+"</a> </td>")
		fmt.Fprint(w, `</tr>
`)
	}

	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
  <p><a href="editchapter?topic=`+strconv.FormatUint(chapter.idTopic, 10)+`&chapter=0">New</a></p>
  </section>
</body>
</html>`)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	switch {
	case op == "listtopics":
		if userid != 0 {
			showTopicListPage(w, r, op, userid)
		}
	case op == "edittopic":
		if userid != 0 {
			showEditTopicPage(w, r, op, userid)
		}
	case op == "picktopic":
		if userid != 0 {
			showTopicPickListPage(w, r, op, userid)
		}
	case op == "listchapters":
		if userid != 0 {
			showChapterListPage(w, r, op, userid)
		}
	case op == "editchapter":
		if userid != 0 {
			showEditChapterPage(w, r, op, userid)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/blind/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
