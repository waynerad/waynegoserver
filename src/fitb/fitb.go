package fitb

import (
	"accessdb"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"io"
	"langtext"
	"math"
	"math/rand"
	"net/http"
	"static"
	"strconv"
	"strings"
	"time"
	"wutil"
)

func showFitbHeadHeader(w http.ResponseWriter, displayInfo map[string]string, defaultClose bool) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta http-equiv="X-UA-Compatible" content="ie=edge" />
        <title>`+displayInfo["hTitle"]+`</title>
        <style>

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
    width: 300px;
}

.biginput {
    font-size: 1.1em;
}

        </style>
`)
	if defaultClose {
		fmt.Fprint(w, `
    </head>
    <body>
`)
	}
}

func showFitbBodyHeader(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
        <div id="header">
            <p>
                <a href="picktopic">My Topics</a>
                &middot; `+displayInfo["hUserName"]+`
            </p>
        </div>
`)
}

func showFitbFooter(w http.ResponseWriter) {
	fmt.Fprint(w, `
        <div id="footer">
            <p>&nbsp;</p>
        </div>
`)
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

func showEditTopicPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		idTopic      uint64
		idUser       uint64
		name         string
		description  string
		introduction string
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_topic, name, description, introduction FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
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
				ui.description = row.Str(2)
				ui.introduction = row.Str(3)
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
			errorOccurred = true
		}
		ui.name = strings.Trim(postform["name"][0], " \r\n\t")
		if ui.name == "" {
			errorList["name"] = "Please specify a name."
			errorOccurred = true
		}
		ui.description = strings.Trim(postform["description"][0], " \r\n\t")
		if ui.description == "" {
			errorList["description"] = "Please specify a description."
			errorOccurred = true
		}
		ui.introduction = strings.Trim(postform["introduction"][0], " \r\n\t")
		if ui.introduction == "" {
			errorList["introduction"] = "Please write an introduction (HTML) ."
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db := accessdb.GetDbConnection()
			defer db.Close()
			var save struct {
				idTopic      uint64
				idUser       uint64
				name         string
				description  string
				introduction string
			}
			save.idTopic = topicid
			save.idUser = userid
			save.name = ui.name
			save.description = ui.description
			save.introduction = ui.introduction
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
				stmt, err := db.Prepare("UPDATE fitb_topic SET name = ?, description = ?, introduction = ? WHERE (id_topic = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.name, save.description, save.introduction, topicid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO fitb_topic (id_topic, id_user, name, description, introduction) VALUES (?, ?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idTopic, save.idUser, save.name, save.description, save.introduction)
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
		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "[Put Page Title Here (283)]"
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)
		fmt.Fprint(w, `
  <section>
    <h1>Topic Entry</h1>
<form action="edittopic" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		db := accessdb.GetDbConnection()
		defer db.Close()
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
<tr><td align="right"> Description: </td><td> <input class="biginput" name="description" id="description" type="text" value="`+html.EscapeString(ui.description)+`" /> </td></tr>
<tr><td align="right"> Introduction (HTML): </td><td> <input class="biginput" name="introduction" id="introduction" type="text" value="`+html.EscapeString(ui.introduction)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input class="infield" type="submit"> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showTopicListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	var topic struct {
		idTopic uint64
		name    string
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "List of Topics"
	showFitbHeadHeader(w, displayInfo, false)
	fmt.Fprint(w, `
<link rel="stylesheet" href="jquery-ui.css" />
<script src="jquery-1.9.1.js"></script>
<script src="jquery-ui.js"></script>
</head>
<body>
`)
	showFitbBodyHeader(w, displayInfo)
	fmt.Fprint(w, `
  <section>
    <h1>List of Topics</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
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
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Edit </th><th> Chapters </th><th> Questions </th><th> &nbsp; </th><th> Renumber </th></tr>`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		topicAsStr := uint64ToStr(topic.idTopic)
		// fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+html.EscapeString(topic.name)+" </td><td> "+`<a href="edittopic?topic=`+topicAsStr+`">Edit</a> </td> <td> <a href="listchapters?topic=`+topicAsStr+`">Chapters</a> </td><td> <a href="listquestions?topic=`+topicAsStr+`">List Questions</a> </td><td> <a href="bulkeditquestions?topic=`+topicAsStr+`">Attempt Bulk Edit</a> </td><td> <a href="bulkaddquestions?topic=`+topicAsStr+`">Bulk Add Questions</a> </td>`)
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+html.EscapeString(topic.name)+" </td><td> "+`<a href="edittopic?topic=`+topicAsStr+`">Edit</a> </td> <td> <a href="listchapters?topic=`+topicAsStr+`">Chapters</a> </td><td> <a href="listquestions?topic=`+topicAsStr+`">List Questions</a> </td><td> <a href="bulkaddquestions?topic=`+topicAsStr+`">Bulk Add Questions</a> </td><td> <a href="renumber?topic=`+uint64ToStr(topic.idTopic)+`">Renumber</a> </td><td> <a href="initialize?topic=`+topicAsStr+`">Go!</td>`)
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
        <p> <a href="edittopic?topic=0">Add Topic</a> </p>
  </section>
`)
	showFitbFooter(w)
}

func showTopicPickListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
	currentTime := uint64(time.Now().Unix())
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "My Topics"
	showFitbHeadHeader(w, displayInfo, true)
	showFitbBodyHeader(w, displayInfo)
	fmt.Fprint(w, `
  <section>
    <h1>My Topics</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// sql = "SELECT id_topic, name FROM fitb_topic WHERE (id_user = ?) ORDER BY id_topic ;"
	sql = "SELECT fitb_topic.id_topic, fitb_topic.name FROM fitb_user_topic_jct, fitb_topic WHERE (fitb_user_topic_jct.id_user = ?) AND (fitb_user_topic_jct.id_topic = fitb_topic.id_topic) ORDER BY fitb_topic.id_topic;"
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
		count++
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		progress := genProgressMessage(db, userid, topic.idTopic, currentTime)
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"initialize?topic=`+uint64ToStr(topic.idTopic)+`">`+html.EscapeString(topic.name)+"</a> </td><td> "+progress+" </td>")
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table border="1">
`)
	}
	fmt.Fprint(w, `
  </section>
`)
	showFitbFooter(w)
}

func showPracticePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
		db := accessdb.GetDbConnection()
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

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "[Put Page Title Here (509)]"
	showFitbHeadHeader(w, displayInfo, true)
	showFitbBodyHeader(w, displayInfo)

	db := accessdb.GetDbConnection()
	defer db.Close()
	fmt.Fprint(w, `
  <section>
    <h1>Practice: `+html.EscapeString(ui.name)+`</h1>
  </section>
`)
	showFitbFooter(w)
}

func showEditChapterPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		idChapter uint64
		idTopic   uint64
		idUser    uint64
		name      string
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
		_, topicExists := getform["topic"]
		if topicExists {
			ui.idTopic = strToUint64(getform["topic"][0])
		}
		_, chapterExists := getform["chapter"]
		if chapterExists {
			chapterid, err = strconv.ParseUint(getform["chapter"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db := accessdb.GetDbConnection()
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
			db := accessdb.GetDbConnection()
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
			errorList["topic"] = "Topic is missing or invalid(1)"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db := accessdb.GetDbConnection()
			defer db.Close()
			var save struct {
				idChapter uint64
				idTopic   uint64
				name      string
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
			http.Redirect(w, r, "listchapters?topic="+uint64ToStr(ui.idTopic), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "Chapter Entry"
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)

		db := accessdb.GetDbConnection()
		defer db.Close()

		fmt.Fprint(w, `
  <section>
    <h1>Chapter Entry</h1>
<form action="editchapter" method="post">
<input type="hidden" name="chapter" value="`+uint64ToStr(chapterid)+`" />
<input class="biginput" name="topic" id="topic" type="hidden" value="`+uint64ToStr(ui.idTopic)+`" />
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
<tr><td colspan="2" align="center"> <input class="infield" type="submit"> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showChapterListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	var chapter struct {
		idChapter uint64
		idTopic   uint64
		name      string
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
	db := accessdb.GetDbConnection()
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

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "Chapters of " + html.EscapeString(topicName)
	showFitbHeadHeader(w, displayInfo, true)
	showFitbBodyHeader(w, displayInfo)

	fmt.Fprint(w, `
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
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"editchapter?chapter=`+uint64ToStr(chapter.idChapter)+`">`+html.EscapeString(chapter.name)+"</a> </td>")
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	fmt.Fprint(w, `
  <p><a href="editchapter?topic=`+uint64ToStr(topicid)+`&chapter=0">New</a></p>
  </section>
`)
	showFitbFooter(w)
}

func calculateLnum(str string) float64 {
	// a	8.167%	->	1.555283
	// b	1.492%	->	8.513405
	// c	2.782%	->	4.56578
	// d	4.253%	->	2.986598
	// e	12.702%	->	1.0
	// f	2.228%	->	5.701077
	// g	2.015%	->	6.303722
	// h	6.094%	->	2.084345
	// i	6.966%	->	1.823428
	// j	0.153%	->	83.01961
	// k	0.772%	->	16.45337
	// l	4.025%	->	3.155776
	// m	2.406%	->	5.279302
	// n	6.749%	->	1.882057
	// o	7.507%	->	1.692021
	// p	1.929%	->	6.58759
	// q	0.095%	->	133.7053
	// r	5.987%	->	2.125197
	// s	6.327%	->	2.00785
	// t	9.056%	->	1.402606
	// u	2.758%	->	4.605511
	// v	0.978%	->	12.98733
	// w	2.361%	->	5.379924
	// x	0.150%	->	24.90588
	// y	1.974%	->	6.4365
	// z	0.074%	->	171.6486

	inAnswer := false
	var rv float64
	fmt.Println(str)
	for idx, char := range str {
		fmt.Println("    ", char) // uncomment this line to debug bad character issues.
		if char == rune("_"[0]) {
			inAnswer = !inAnswer
		} else {
			// calibrated to "e"
			var l float64
			l = 0.0
			// undoubtedly there is a smarter way to do this
			if char == rune("a"[0]) {
				l = 1.555283
			}
			if char == rune("b"[0]) {
				l = 8.513405
			}
			if char == rune("c"[0]) {
				l = 4.56578
			}
			if char == rune("d"[0]) {
				l = 2.986598
			}
			if char == rune("e"[0]) {
				l = 1.0
			}
			if char == rune("f"[0]) {
				l = 5.701077
			}
			if char == rune("g"[0]) {
				l = 6.303722
			}
			if char == rune("h"[0]) {
				l = 2.084345
			}
			if char == rune("i"[0]) {
				l = 1.823428
			}
			if char == rune("j"[0]) {
				l = 83.01961
			}
			if char == rune("k"[0]) {
				l = 16.45337
			}
			if char == rune("l"[0]) {
				l = 3.155776
			}
			if char == rune("m"[0]) {
				l = 5.279302
			}
			if char == rune("n"[0]) {
				l = 1.882057
			}
			if char == rune("o"[0]) {
				l = 1.692021
			}
			if char == rune("p"[0]) {
				l = 6.58759
			}
			if char == rune("q"[0]) {
				l = 133.7053
			}
			if char == rune("r"[0]) {
				l = 2.125197
			}
			if char == rune("s"[0]) {
				l = 2.00785
			}
			if char == rune("t"[0]) {
				l = 1.402606
			}
			if char == rune("u"[0]) {
				l = 4.605511
			}
			if char == rune("v"[0]) {
				l = 12.98733
			}
			if char == rune("w"[0]) {
				l = 5.379924
			}
			if char == rune("x"[0]) {
				l = 24.90588
			}
			if char == rune("y"[0]) {
				l = 6.4365
			}
			if char == rune("z"[0]) {
				l = 171.6486
			}
			if char == rune("A"[0]) {
				l = 1.555283 + 1.0
			}
			if char == rune("B"[0]) {
				l = 8.513405 + 1.0
			}
			if char == rune("C"[0]) {
				l = 4.56578 + 1.0
			}
			if char == rune("D"[0]) {
				l = 2.986598 + 1.0
			}
			if char == rune("E"[0]) {
				l = 1.0 + 1.0
			}
			if char == rune("F"[0]) {
				l = 5.701077 + 1.0
			}
			if char == rune("G"[0]) {
				l = 6.303722 + 1.0
			}
			if char == rune("H"[0]) {
				l = 2.084345 + 1.0
			}
			if char == rune("I"[0]) {
				l = 1.823428 + 1.0
			}
			if char == rune("J"[0]) {
				l = 83.01961 + 1.0
			}
			if char == rune("K"[0]) {
				l = 16.45337 + 1.0
			}
			if char == rune("L"[0]) {
				l = 3.155776 + 1.0
			}
			if char == rune("M"[0]) {
				l = 5.279302 + 1.0
			}
			if char == rune("N"[0]) {
				l = 1.882057 + 1.0
			}
			if char == rune("O"[0]) {
				l = 1.692021 + 1.0
			}
			if char == rune("P"[0]) {
				l = 6.58759 + 1.0
			}
			if char == rune("Q"[0]) {
				l = 133.7053 + 1.0
			}
			if char == rune("R"[0]) {
				l = 2.125197 + 1.0
			}
			if char == rune("S"[0]) {
				l = 2.00785 + 1.0
			}
			if char == rune("T"[0]) {
				l = 1.402606 + 1.0
			}
			if char == rune("U"[0]) {
				l = 4.605511 + 1.0
			}
			if char == rune("V"[0]) {
				l = 12.98733 + 1.0
			}
			if char == rune("W"[0]) {
				l = 5.379924 + 1.0
			}
			if char == rune("X"[0]) {
				l = 24.90588 + 1.0
			}
			if char == rune("Y"[0]) {
				l = 6.4365 + 1.0
			}
			if char == rune("Z"[0]) {
				l = 171.6486 + 1.0
			}
			if char == rune(","[0]) {
				l = 200
			}
			if char == rune("-"[0]) {
				l = 201
			}
			if char == rune("."[0]) {
				l = 202
			}
			if char == rune("?"[0]) {
				l = 203
			}
			if char == rune("!"[0]) {
				l = 204
			}
			if char == rune(":"[0]) {
				l = 205
			}
			if char == rune(";"[0]) {
				l = 206
			}
			if char == rune("'"[0]) {
				l = 207
			}
			if char == rune(`"`[0]) {
				l = 208
			}
			if char == rune("/"[0]) {
				l = 300
			}
			if char == rune("<"[0]) {
				l = 301
			}
			if char == rune(">"[0]) {
				l = 302
			}
			if char == rune("["[0]) {
				l = 303
			}
			if char == rune("("[0]) {
				l = 209
			}
			if char == rune(")"[0]) {
				l = 210
			}
			if char == rune("]"[0]) {
				l = 304
			}
			if char == rune("="[0]) {
				l = 305
			}
			if char == rune("+"[0]) {
				l = 306
			}
			if char == rune("{"[0]) {
				l = 307
			}
			if char == rune("}"[0]) {
				l = 308
			}
			if char == rune("$"[0]) {
				l = 310
			}
			if char == rune("~"[0]) {
				l = 600
			}
			if char == rune("1"[0]) {
				l = 211
			}
			if char == rune("2"[0]) {
				l = 211
			}
			if char == rune("3"[0]) {
				l = 211
			}
			if char == rune("4"[0]) {
				l = 211
			}
			if char == rune("5"[0]) {
				l = 211
			}
			if char == rune("6"[0]) {
				l = 211
			}
			if char == rune("7"[0]) {
				l = 211
			}
			if char == rune("8"[0]) {
				l = 211
			}
			if char == rune("9"[0]) {
				l = 211
			}
			if char == rune("0"[0]) {
				l = 211
			}
			if char == rune("|"[0]) {
				l = 211
			}
			if char == rune("&"[0]) {
				l = 211
			}
			if char == 92 { // backslash
				l = 500
			}
			if char == 181 { // greek character mu for "micro" for Metric (SI) prefix abbreviation
				l = 999
			}
			if char == rune(" "[0]) {
				l = 0.5 // works as long as the e add is more than 0.7
			}
			// Just for French
			if char == 232 {
				// e`
				l = 150
			}
			if char == 233 {
				// e'
				l = 149
			}
			if char == 244 {
				// o^
				l = 170
			}
			if char == 224 {
				// a`
				l = 151
			}
			if char == 231 {
				// ,c
				l = 155
			}
			if char == 226 {
				// a^
				l = 171
			}
			if char == 234 {
				// e^
				l = 171
			}
			if char == 171 {
				// << quote
				l = 172
			}
			if char == 187 {
				// >> quote
				l = 172
			}
			if char == 339 {
				// oe
				l = 156
			}
			if char == 192 {
				// capital A`
				l = 351 // lower case + 200 for the shift
			}
			if char == 252 {
				// "u  (umlauted u) (for Spanish)
				l = 204.605511 // lower case + 200 for the shift
			}
			if char == 243 {
				// 'o  (accented o) (for Spanish)
				l = 201.692021 // lower case + 200 for the shift
			}
			if char == 237 {
				// 'i  (accented i) (for Spanish)
				l = 201.823428 // lower case + 200 for the shift
			}
			if char == 239 {
				// "i  (umlauted i) (for French believe it or not)
				l = 201.823428 // lower case + 200 for the umlaut
			}
			if char == 249 {
				// `u  (back accented u) (for French)
				l = 204.605511 // lower case + 200 for accent
			}
			if char == 37 {
				// % symbol
				l = 300
			}
			if char == 251 {
				// ^u
				l = 204.605511 // lower case + 200 for caret
			}
			if char == 238 {
				// ^i
				l = 201.823428 // lower case + 200 for the caret
			}
			if char == 199 {
				// C//
				l = 204.56578 // upper case + 200 for the diatric
			}
			if inAnswer {
				if l == 0.0 {
					panic("Encountered uncountable character in input: character '" + string(char) + "' value " + intToStr(int(char)) + " position " + intToStr(idx))
				}
				rv = rv + math.Log(l) + 5.0 // makes e 5.0
			}
		}
	}
	return rv
}

func showBulkEditQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	type questionType struct {
		idQuestion uint64
		idChapter  uint64
		seqNum     int64
		theFitbStr string
		lnum       float64
	}
	var allQuests []questionType
	allQuests = make([]questionType, 0)
	if method == "GET" {
		// set defaults
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_question, id_chapter, sequence_num, the_fitb_str, lnum FROM fitb_question WHERE (id_topic = ?);"
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
			for _, row := range rows {
				var qst questionType
				qst.idQuestion = row.Uint64(0)
				qst.idChapter = row.Uint64(1)
				qst.seqNum = row.Int64(2)
				qst.theFitbStr = row.Str(3)
				qst.lnum = row.Float(4)
				allQuests = append(allQuests, qst)
			}
			sql = "SELECT name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err = db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err = sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				topicName = row.Str(0)
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
		topicid, err := strconv.ParseUint(postform["topic"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		db := accessdb.GetDbConnection()
		userOwnsTopic := false
		if topicid > 0 {
			defer db.Close()
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
		}
		if userOwnsTopic {
			type saveTyp struct {
				idQuestion  uint64
				idTopic     uint64
				idChapter   uint64
				sequenceNum int64
				theFitbStr  string
				lnum        float64
			}
			var allSaves []saveTyp
			allSaves = make([]saveTyp, 0)
			i := 0
			_, inptRowExists := postform["id"+intToStr(i)]
			for inptRowExists {
				var currentSave saveTyp
				currentSave.idQuestion = strToUint64(postform["id"+intToStr(i)][0])
				currentSave.idTopic = topicid
				currentSave.idChapter = strToUint64(postform["chapter"+intToStr(i)][0])
				currentSave.sequenceNum = strToInt64(postform["seq"+intToStr(i)][0])
				currentSave.theFitbStr = trim(postform["quest"+intToStr(i)][0])
				currentSave.lnum = calculateLnum(currentSave.theFitbStr)
				allSaves = append(allSaves, currentSave)
				i++
				_, inptRowExists = postform["id"+intToStr(i)]
			}
			for _, save := range allSaves {
				if save.idQuestion == 0 {
					if save.theFitbStr != "" {
						sql := "INSERT INTO fitb_question (id_question, id_topic, id_chapter, sequence_num, the_fitb_str, lnum) VALUES (?, ?, ?, ?, ?, ?);"
						stmt, err := db.Prepare(sql)
						if err != nil {
							fmt.Println(err)
							panic("Prepare failed")
						}
						// defer stmt.Close();
						stmt.Bind(save.idQuestion, save.idTopic, save.idChapter, save.sequenceNum, save.theFitbStr, save.lnum)
						_, _, err = stmt.Exec()
					}
				} else {
					sql := "UPDATE fitb_question SET id_chapter = ?, sequence_num = ?, the_fitb_str = ?, lnum = ?) WHERE id_question = ?;"
					stmt, err := db.Prepare(sql)
					if err != nil {
						fmt.Println(err)
						panic("Prepare failed")
					}
					// defer stmt.Close();
					stmt.Bind(save.idChapter, save.sequenceNum, save.theFitbStr, save.lnum, save.idQuestion)
					_, _, err = stmt.Exec()
				}
			}
		} else {
			errorList["topic"] = "Topic is missing or invalid(2)"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			http.Redirect(w, r, "listtopics?topic="+uint64ToStr(topicid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "[Put Page Title Here (1398)]"
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)

		db := accessdb.GetDbConnection()
		defer db.Close()
		type chapterType struct {
			idChapter uint64
			name      string
		}
		var allChapters []chapterType
		allChapters = make([]chapterType, 0)
		sql := "SELECT id_chapter, name FROM fitb_chapter WHERE id_topic = ?;"
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
		for _, row := range rows {
			var currentChap chapterType
			currentChap.idChapter = row.Uint64(0)
			currentChap.name = row.Str(1)
			allChapters = append(allChapters, currentChap)
		}
		fmt.Fprint(w, `
  <section>
    <h1>List questions for topic: `+topicName+`</h1>
<form action="bulkeditquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
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
<tr><th> Sequence number </th><th> Chapter <th> Question in FITB format </th><th> Size </th></tr>
`)
		last := 0
		var maxSeq int64
		maxSeq = 0
		var lstChapt uint64
		lstChapt = 0
		for count, currentQuest := range allQuests {
			fmt.Fprint(w, `
<tr><td> <input type="hidden" name="id`+intToStr(count)+`" id="id" value="`+uint64ToStr(currentQuest.idQuestion)+`" /><input class="infield" type="text" name="seq`+intToStr(count)+`" id="seq`+intToStr(count)+`" value="`+int64ToStr(currentQuest.seqNum)+`" /> </td><td> <select name="chapter`+intToStr(count)+`" id="chapter`+intToStr(count)+`">
`)
			for _, currentChapt := range allChapters {
				fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
				if currentChapt.idChapter == currentQuest.idChapter {
					fmt.Fprint(w, ` selected="selected"`)
				}
				fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
			}
			fmt.Fprint(w, `</select> </td><td> <input class="infield" type="text" name="quest`+intToStr(count)+`" id="quest`+intToStr(count)+`" value="`+htmlize(currentQuest.theFitbStr)+`" /> </td><td align="right"> `+floatToStr(currentQuest.lnum)+` </td></tr>
`)
			last = count // to get this var out of the loop
			if currentQuest.seqNum > maxSeq {
				maxSeq = currentQuest.seqNum
			}
			lstChapt = currentQuest.idChapter
		}
		// add 10 blanks
		for i := 0; i < 9; i++ {
			count := last + i
			currentSeqNum := maxSeq + (10 * int64(i+1))
			fmt.Fprint(w, `
<tr><td> <input type="hidden" name="id`+intToStr(count)+`" id="id" value="0" /><input class="infield" type="text" name="seq`+intToStr(count)+`" id="seq`+intToStr(count)+`" value="`+int64ToStr(currentSeqNum)+`" /> </td><td> <select name="chapter`+intToStr(count)+`" id="chapter`+intToStr(count)+`">
`)
			for _, currentChapt := range allChapters {
				fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
				if currentChapt.idChapter == lstChapt {
					fmt.Fprint(w, ` selected="selected"`)
				}
				fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
			}
			fmt.Fprint(w, `</select> </td><td> <input class="infield" type="text" name="quest`+intToStr(count)+`" id="quest`+intToStr(count)+`" value="" /> </td><td> &nbsp; </td> </tr>
`)
		}
		fmt.Fprint(w, `
<tr><td colspan="3" align="center"> <input class="infield" type="submit"> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showListQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	type questionType struct {
		idQuestion uint64
		idChapter  uint64
		seqNum     int64
		theFitbStr string
		lnum       float64
	}
	var quest questionType
	var allQuests []questionType
	allQuests = make([]questionType, 0)
	if method == "GET" {
		// set defaults
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
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
			topicExists = false
			for _, row := range rows {
				topicExists = true
				topicName = row.Str(0)
			}
			if topicExists {
				sql := "SELECT id_question, id_chapter, sequence_num, the_fitb_str, lnum FROM fitb_question WHERE (id_topic = ?) ORDER BY sequence_num;"
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
				for _, row := range rows {
					quest.idQuestion = row.Uint64(0)
					quest.idChapter = row.Uint64(1)
					quest.seqNum = row.Int64(2)
					quest.theFitbStr = row.Str(3)
					quest.lnum = row.Float(4)
					allQuests = append(allQuests, quest)
				}
			}
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "Edit questions for topic: " + htmlize(topicName)
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)

		db := accessdb.GetDbConnection()
		defer db.Close()
		type chapterType struct {
			idChapter uint64
			name      string
		}
		var allChapters map[uint64]string
		allChapters = make(map[uint64]string)
		allChapters[0] = "Unspecified"
		sql := "SELECT id_chapter, name FROM fitb_chapter WHERE id_topic = ?;"
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
		for _, row := range rows {
			var currentChap chapterType
			currentChap.idChapter = row.Uint64(0)
			currentChap.name = row.Str(1)
			allChapters[currentChap.idChapter] = currentChap.name
		}
		fmt.Fprintln(w, `
  <section>
    <h1>Edit questions for topic: `+topicName+`</h1>
<form action="listquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
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
<tr><th> Sequence number </th><th> Chapter <th> Question in FITB format point2 </th><th> Size </th></tr>
`)
		for _, currentQuest := range allQuests {
			fmt.Fprint(w, `
<tr><td> `+int64ToStr(currentQuest.seqNum)+` </td><td> `+allChapters[currentQuest.idChapter]+` </td> <td> `+htmlize(currentQuest.theFitbStr)+` </td><td align="right"> `+floatToStr(currentQuest.lnum)+` </td><td> <a href="editquestion?question=`+uint64ToStr(currentQuest.idQuestion)+`">Edit</a> <a href="deletequestion?question=`+uint64ToStr(currentQuest.idQuestion)+`">Delete</a> </td> </tr>
`)
		}
		fmt.Fprint(w, `
<tr><td> <a href="bulkaddquestions?topic=`+uint64ToStr(topicid)+`">Add Questions</a> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showAddBulkQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	type questionType struct {
		idQuestion uint64
		idChapter  uint64
		seqNum     int64
		theFitbStr string
		lnum       float64
	}
	if method == "GET" {
		// set defaults
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
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
			topicExists = false
			for _, row := range rows {
				topicExists = true
				topicName = row.Str(0)
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
			fmt.Println(err)
			panic("parseform failed")
		}
		chapterid, err := strconv.ParseUint(postform["chapter"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		db := accessdb.GetDbConnection()
		userOwnsTopic := false
		if topicid > 0 {
			defer db.Close()
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
		}
		if userOwnsTopic {
			var seqNum int64
			seqNum = 0
			sql := "SELECT MAX(sequence_num) FROM fitb_question WHERE (id_topic = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Exec() failed")
			}
			for _, row := range rows {
				seqNum = row.Int64(0)
			}
			seqNum += 10
			type saveTyp struct {
				idTopic     uint64
				idChapter   uint64
				sequenceNum int64
				theFitbStr  string
				lnum        float64
			}
			var allSaves []saveTyp
			allSaves = make([]saveTyp, 0)
			bulkAdditionStr := postform["bulkadditions"][0]
			bulkAddSlice := strings.Split(bulkAdditionStr, "\n")
			for _, entry := range bulkAddSlice {
				trEnt := trim(entry)
				if len(trEnt) > 0 {
					if trEnt[0:1] == "#" {
						trEnt = "" // if it's a comment, throw it away
					}
				}
				if trEnt != "" {
					var currentSave saveTyp
					currentSave.idTopic = topicid
					currentSave.idChapter = chapterid
					currentSave.sequenceNum = seqNum
					currentSave.theFitbStr = trEnt
					currentSave.lnum = calculateLnum(currentSave.theFitbStr)
					allSaves = append(allSaves, currentSave)
					seqNum += 10
				}
			}
			for _, save := range allSaves {
				if save.theFitbStr != "" {
					sql := "INSERT INTO fitb_question (id_topic, id_chapter, sequence_num, the_fitb_str, lnum) VALUES (?, ?, ?, ?, ?);"
					stmt, err := db.Prepare(sql)
					if err != nil {
						fmt.Println(err)
						panic("Prepare failed")
					}
					// defer stmt.Close();
					stmt.Bind(save.idTopic, save.idChapter, save.sequenceNum, save.theFitbStr, save.lnum)
					_, _, err = stmt.Exec()
				}
			}
		} else {
			errorList["topic"] = "Topic is missing or invalid(2)"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db := accessdb.GetDbConnection()
			defer db.Close()
			// var save struct {
			//	idChapter uint64
			//	idTopic uint64
			//	name    string
			// }
			// save.idChapter = chapterid
			// save.idTopic = ui.idTopic
			// save.name = ui.name
			http.Redirect(w, r, "listtopics?topic="+uint64ToStr(topicid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "Bulk add questions for topic: " + htmlize(topicName)
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)
		db := accessdb.GetDbConnection()
		defer db.Close()
		type chapterType struct {
			idChapter uint64
			name      string
		}
		var allChapters []chapterType
		allChapters = make([]chapterType, 0)
		var maxChapter uint64
		maxChapter = 0
		sql := "SELECT id_chapter, name FROM fitb_chapter WHERE id_topic = ?;"
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
		for _, row := range rows {
			var currentChap chapterType
			currentChap.idChapter = row.Uint64(0)
			currentChap.name = row.Str(1)
			allChapters = append(allChapters, currentChap)
			if currentChap.idChapter > maxChapter {
				maxChapter = currentChap.idChapter
			}
		}
		fmt.Fprint(w, `
  <section>
    <h1>Bulk add questions for topic: `+htmlize(topicName)+`</h1>
<form action="bulkaddquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<p>Chapter: <select name="chapter" id="chapter">
`)
		for _, currentChapt := range allChapters {
			fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
			if currentChapt.idChapter == maxChapter {
				fmt.Fprint(w, ` selected="selected"`)
			}
			fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
		}
		fmt.Fprint(w, `
</select></p>
<p><textarea name="bulkadditions" id="bulkadditions" rows="50" cols="80"></textarea></p>
<p><input class="infield" type="submit" /></p>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showEditQuestionPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	var questionid uint64
	questionid = 0
	// type questionType struct {
	//	idQuestion uint64
	//	idChapter  uint64
	//	seqNum     int64
	//	theFitbStr string
	//	lnum       float64
	// }
	var ui struct {
		idChapter  uint64
		theFitbStr string
	}
	if method == "GET" {
		// set defaults
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, questionExists := getform["question"]
		if questionExists {
			questionid, err = strconv.ParseUint(getform["question"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_topic, id_chapter, the_fitb_str FROM fitb_question WHERE (id_question = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(questionid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				topicid = row.Uint64(0)
				ui.idChapter = row.Uint64(1)
				ui.theFitbStr = row.Str(2)
			}
			sql = "SELECT name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err = db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err = sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				topicName = row.Str(0)
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
		questionid = strToUint64(postform["question"][0])
		topicid = strToUint64(postform["topic"][0])
		chapterid := strToUint64(postform["chapter"][0])
		ui.idChapter = chapterid
		fitbStr := postform["fitb"][0]
		ui.theFitbStr = fitbStr
		db := accessdb.GetDbConnection()
		userOwnsTopic := false
		if topicid > 0 {
			defer db.Close()
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
		}
		if userOwnsTopic {
			var save struct {
				idQuestion uint64
				idTopic    uint64
				idChapter  uint64
				theFitbStr string
				lnum       float64
			}
			save.idQuestion = questionid
			save.idTopic = topicid
			save.idChapter = chapterid
			save.theFitbStr = fitbStr
			save.lnum = calculateLnum(save.theFitbStr)
			if save.idQuestion != 0 {
				if save.theFitbStr != "" {
					sql := "UPDATE fitb_question SET id_chapter = ?, the_fitb_str = ?, lnum = ? WHERE id_question = ?;"
					stmt, err := db.Prepare(sql)
					if err != nil {
						fmt.Println(err)
						panic("Prepare failed")
					}
					// defer stmt.Close();
					stmt.Bind(save.idChapter, save.theFitbStr, save.lnum, save.idQuestion)
					_, _, err = stmt.Exec()
				}
			}
		} else {
			errorList["topic"] = "Topic is missing or invalid(2)"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db := accessdb.GetDbConnection()
			defer db.Close()
			// var save struct {
			//	idChapter uint64
			//	idTopic uint64
			//	name    string
			// }
			// save.idChapter = chapterid
			// save.idTopic = ui.idTopic
			// save.name = ui.name
			http.Redirect(w, r, "listquestions?topic="+uint64ToStr(topicid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "[Put Page Title Here (2046)]"
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)

		db := accessdb.GetDbConnection()
		defer db.Close()
		type chapterType struct {
			idChapter uint64
			name      string
		}
		var allChapters []chapterType
		allChapters = make([]chapterType, 0)
		sql := "SELECT id_chapter, name FROM fitb_chapter WHERE id_topic = ?;"
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
		for _, row := range rows {
			var currentChap chapterType
			currentChap.idChapter = row.Uint64(0)
			currentChap.name = row.Str(1)
			allChapters = append(allChapters, currentChap)
		}
		fmt.Fprint(w, `
  <section>
    <h1>Edit questions for topic: `+topicName+`</h1>
<form action="editquestion" method="post">
<input type="hidden" name="question" value="`+uint64ToStr(questionid)+`" />
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
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
<tr><td> Chapter </td><td> <select name="chapter" id="chapter">
`)
		for _, currentChapt := range allChapters {
			fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
			if currentChapt.idChapter == ui.idChapter {
				fmt.Fprint(w, ` selected="selected"`)
			}
			fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
		}
		fmt.Fprint(w, `</select> </td></tr>
<tr><td colspan="2"> Question in FITB format <br /><textarea name="fitb" id="fitb" rows="8" cols="80">`+htmlize(ui.theFitbStr)+`</textarea></td></tr>
</table>
`)
		fmt.Fprint(w, `
<tr><td colspan="2" align="center"> <input class="infield" type="submit"> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showDeleteQuestionPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	theFitbStr := ""
	var questionid uint64
	questionid = 0
	if method == "GET" {
		// set defaults
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, questionExists := getform["question"]
		if questionExists {
			questionid, err = strconv.ParseUint(getform["question"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_topic, the_fitb_str FROM fitb_question WHERE (id_question = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(questionid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				topicid = row.Uint64(0)
				theFitbStr = row.Str(1)
			}
			sql = "SELECT name FROM fitb_topic WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err = db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err = sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				topicName = row.Str(0)
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
		questionid = strToUint64(postform["question"][0])
		topicid = strToUint64(postform["topic"][0])
		db := accessdb.GetDbConnection()
		userOwnsTopic := false
		if topicid > 0 {
			defer db.Close()
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
		}
		if userOwnsTopic {
			// var del struct {
			// 	idQuestion uint64
			// 	idUser  uint64
			// }
			// del.idQuestion = questionid
			// del.idUser = userid
			// del.theFitbStr = fitbStr
			// del.lnum = calculateLnum(del.theFitbStr)
			if questionid != 0 {
				sql := "DELETE FROM fitb_question WHERE id_question = ?;"
				stmt, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(questionid)
				_, _, err = stmt.Exec()
				sql = "DELETE FROM fitb_user_question_jct WHERE id_question = ?;"
				stmt, err = db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(questionid)
				_, _, err = stmt.Exec()
			}
		} else {
			errorList["topic"] = "Topic is missing or invalid(2)"
			errorOccurred = true
		}
		if errorOccurred {
			showform = true
		} else {
			db := accessdb.GetDbConnection()
			defer db.Close()
			http.Redirect(w, r, "listquestions?topic="+uint64ToStr(topicid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = "[Put Page Title Here (2255)]"
		showFitbHeadHeader(w, displayInfo, true)
		showFitbBodyHeader(w, displayInfo)

		db := accessdb.GetDbConnection()
		defer db.Close()
		type chapterType struct {
			idChapter uint64
			name      string
		}
		var allChapters []chapterType
		allChapters = make([]chapterType, 0)
		sql := "SELECT id_chapter, name FROM fitb_chapter WHERE id_topic = ?;"
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
		for _, row := range rows {
			var currentChap chapterType
			currentChap.idChapter = row.Uint64(0)
			currentChap.name = row.Str(1)
			allChapters = append(allChapters, currentChap)
		}
		fmt.Fprint(w, `
  <section>
    <h1>Delete question for topic: `+topicName+`</h1>
<form action="deletequestion" method="post">
<input type="hidden" name="question" value="`+uint64ToStr(questionid)+`" />
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
	<p>Are you sure you want to delete:</p>
	<p>`+htmlize(theFitbStr)+`</p>
`)
		fmt.Fprint(w, `
<tr><td colspan="2" align="center"> <input class="infield" type="submit" value="Delete"> </td></tr>
</table>
</form>
  </section>
`)
		showFitbFooter(w)
	}
}

func showRenumberPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	method := r.Method
	var topicid uint64
	topicid = 0
	if method == "GET" {
		// set defaults
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("Parseform failed")
		}
		getform := r.Form
		_, topicExists := getform["topic"]
		if topicExists {
			topicid = strToUint64(getform["topic"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			type updateTyp struct {
				questionid uint64
				lnum       float64
			}
			var updateList []updateTyp
			updateList = make([]updateTyp, 0)
			sql := "SELECT id_question, the_fitb_str FROM fitb_question WHERE id_topic = ? ORDER BY id_question;"
			sel1, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel1.Bind(topicid)
			rows, _, err := sel1.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				var x updateTyp
				x.questionid = row.Uint64(0)
				fitbStr := row.Str(1)
				x.lnum = calculateLnum(fitbStr)
				updateList = append(updateList, x)
			}
			sql = "UPDATE fitb_question SET lnum = ? WHERE id_question = ?"
			stmt1, err := db.Prepare(sql)
			if err != nil {
				panic("Prepare failed")
			}
			for _, updateInfo := range updateList {
				stmt1.Bind(updateInfo.lnum, updateInfo.questionid)
				_, _, err = stmt1.Exec()
			}
			var questionList []uint64
			questionList = make([]uint64, 0)
			// sql = "SELECT id_question FROM fitb_question WHERE id_topic = ? ORDER BY id_chapter, lnum;"
			sql = "SELECT fitb_question.id_question FROM fitb_chapter, fitb_question WHERE (fitb_chapter.id_chapter = fitb_question.id_chapter) AND (fitb_question.id_topic = ?) ORDER BY fitb_chapter.sequence_num, fitb_question.lnum;"
			sel2, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel2.Bind(topicid)
			rows, _, err = sel2.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				questionid := row.Uint64(0)
				questionList = append(questionList, questionid)
			}
			var seqNum int64
			seqNum = 10
			sql = "UPDATE fitb_question SET sequence_num = ? WHERE id_question = ?"
			stmt2, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			for _, questionid := range questionList {
				seqNum = seqNum + 10
				stmt2.Bind(seqNum, questionid)
				_, _, err = stmt2.Exec()
			}
		}
	}
	if method == "POST" {
		panic("POST request received on renumber page.")
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "[Put Page Title Here (2405)]"
	showFitbHeadHeader(w, displayInfo, true)
	showFitbBodyHeader(w, displayInfo)

	db := accessdb.GetDbConnection()
	defer db.Close()
	fmt.Fprint(w, `
  <section>
    <h1>Renumber Topic</h1>
<form action="renumbertopic" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
	fmt.Fprint(w, `
<table border="0" cellpadding="4">
</table>
</form>
<p>Done</p>
  </section>
`)
	showFitbFooter(w)
}

func initialize(db mysql.Conn, userid uint64, topicid uint64) {
	// set up user userid so they are ready to go on topic topic id -- i.e. set up the junction table fitb_user_question_jct
	var questionid uint64
	jctMap := make(map[uint64]bool)
	sql := "SELECT id_question FROM fitb_user_question_jct WHERE (id_topic = ?) AND (id_user = ?);"

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
	for _, row := range rows {
		questionid = row.Uint64(0)
		jctMap[questionid] = true
	}

	questionList := make([]uint64, 0)
	seqNumMap := make(map[uint64]int64)
	sql = "SELECT id_question, sequence_num FROM fitb_question WHERE (id_topic = ?) ORDER BY id_question;"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(topicid)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, row := range rows {
		questionid = row.Uint64(0)
		sequenceNum := row.Int64(1)
		_, there := jctMap[questionid]
		if !there {
			questionList = append(questionList, questionid)
			seqNumMap[questionid] = sequenceNum
		}
	}

	sql = "INSERT INTO fitb_user_question_jct (id_user, id_question, id_topic, ask_time_gmt, time_interval, factorup, factordown) VALUES (?, ?, ?, ?, ?, ?, ?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	// defer stmt.Close();

	for _, questionid := range questionList {

		sequenceNum := seqNumMap[questionid]
		askTime := 1 + sequenceNum                                    // This is so the sequence of first introduction is the same as the sequence_num bers in the database
		stmt.Bind(userid, questionid, topicid, askTime, 10, 1.0, 1.0) // INITIAL REPETITIVENESS is set here
		_, _, err = stmt.Exec()
	}
}

func showInitializePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	topicIntroduction := ""
	count := 0
	if method == "GET" {
		// set defaults
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("Parseform failed")
		}
		getform := r.Form
		_, topicExists := getform["topic"]
		if topicExists {
			topicid = strToUint64(getform["topic"][0])
			db := accessdb.GetDbConnection()
			defer db.Close()
			// get the name & introduction text
			sql := "SELECT name, introduction FROM fitb_topic WHERE id_topic = ?;"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Exec() failed")
			}
			for _, row := range rows {
				topicName = row.Str(0)
				topicIntroduction = row.Str(1)
			}
			sql = "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_topic = ?) AND (id_user = ?);"
			sel, err = db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(topicid, userid)
			rows, _, err = sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Exec() failed")
			}
			for _, row := range rows {
				count = row.Int(0)
			}
			initialize(db, userid, topicid)
		}
	}
	if method == "POST" {
		panic("POST request received on initialize page.")
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")

	displayInfo := make(map[string]string)
	displayInfo["hUserName"] = userName
	displayInfo["hTitle"] = "[Put Page Title Here (2550)]"
	showFitbHeadHeader(w, displayInfo, true)
	showFitbBodyHeader(w, displayInfo)

	db := accessdb.GetDbConnection()
	defer db.Close()
	var title string
	var buttonFace string
	if count == 0 {
		title = "Begin: " + htmlize(topicName)
		buttonFace = "Begin!"
	} else {
		title = "Resume: " + htmlize(topicName)
		buttonFace = "Resume!"
	}
	fmt.Fprint(w, `
  <section>
    <h1>`+title+`</h1>
<form action="quiz" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
<input type="hidden" name="questionjct" value="0" />
<input type="hidden" name="response0" value="" />
<p>`+topicIntroduction+`</p>
<p><input class="infield" type="submit" value="`+buttonFace+`" />
</form>
  </section>
`)
	showFitbFooter(w)
}

func followJctToQuestionInfo(db mysql.Conn, questionjctid uint64, userid uint64) (uint64, uint64, uint64, uint64, float64, float64, string, string) {
	var questionid uint64
	var topicid uint64
	var askTimeGmt uint64
	var timeInterval uint64
	var factorup float64
	var factordown float64
	var theFitbStr string
	var topicName string
	sql := "SELECT id_question, id_topic, ask_time_gmt, time_interval, factorup, factordown FROM fitb_user_question_jct WHERE (id_uq_jct = ?) AND (id_user = ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(questionjctid, userid)
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, row := range rows {
		questionid = row.Uint64(0)
		topicid = row.Uint64(1)
		askTimeGmt = row.Uint64(2)
		timeInterval = row.Uint64(3)
		factorup = row.Float(4)
		factordown = row.Float(5)
	}
	sql = "SELECT the_fitb_str FROM fitb_question WHERE id_question = ?;"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(questionid)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, row := range rows {
		theFitbStr = row.Str(0) //  + ", ponky = _ponky_."
	}
	sql = "SELECT name FROM fitb_topic WHERE id_topic = ?;"
	sel, err = db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(topicid)
	rows, _, err = sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	for _, row := range rows {
		topicName = row.Str(0)
	}
	return questionid, topicid, askTimeGmt, timeInterval, factorup, factordown, theFitbStr, topicName
}

func updateInterval(db mysql.Conn, questionjctid uint64, askTime uint64, timeInterval uint64, factorup float64, factordown float64, userid uint64) {
	sql := "UPDATE fitb_user_question_jct SET ask_time_gmt = ?, time_interval = ?, factorup = ?, factordown = ? WHERE (id_uq_jct = ?) AND (id_user = ?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	stmt.Bind(askTime, timeInterval, factorup, factordown, questionjctid, userid)
	_, _, err = stmt.Exec()
}

func getCountFromSQLForUserTopicAndTime(db mysql.Conn, sql string, userid uint64, topicid uint64, currentTime uint64, bindCurrentTime bool) int {
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	// this is kinda werd, having the number of bind variables be variable, but allows us to do this one function for all the counts
	if bindCurrentTime {
		sel.Bind(userid, topicid, currentTime)
	} else {
		sel.Bind(userid, topicid)
	}
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Exec() failed")
	}
	count := -1 // we should always get a value back
	for _, row := range rows {
		count = row.Int(0)
	}
	if count < 0 {
		fmt.Println("Count query returned nothing")
		panic("Count query returned nothing")
	}
	return count
}

func genProgressMessage(db mysql.Conn, userid uint64, topicid uint64, currentTime uint64) string {
	original := false
	if original {
		sql := "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt < 1474838209);" // the time code here is the moment the program first went live
		notStarted := getCountFromSQLForUserTopicAndTime(db, sql, userid, topicid, currentTime, false)
		sql = "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt > 1474838209) AND (ask_time_gmt < ?);"
		needToBeReviewed := getCountFromSQLForUserTopicAndTime(db, sql, userid, topicid, currentTime, true)
		sql = "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt > ?);"
		learned := getCountFromSQLForUserTopicAndTime(db, sql, userid, topicid, currentTime, true)
		return intToStr(learned) + " learned, " + intToStr(needToBeReviewed) + " need to be reviewed, " + intToStr(notStarted) + " not started.<br />" + floatToStr(float64(learned*100)/float64(learned+needToBeReviewed)) + "% of started entries learned, " + floatToStr(float64(learned*100)/float64(learned+needToBeReviewed+notStarted)) + "% of total entries learned."
	} else {
		sql := "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt < ?);"
		notLearned := getCountFromSQLForUserTopicAndTime(db, sql, userid, topicid, currentTime, true)
		sql = "SELECT COUNT(*) FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt > ?);"
		learned := getCountFromSQLForUserTopicAndTime(db, sql, userid, topicid, currentTime, true)
		if (learned + notLearned) == 0 {
			return ""
		}
		return floatToStr(float64(learned*100)/float64(learned+notLearned)) + "%"
	}
}

func calcFullAnswer(theFitbStr string) string {
	fullAnswer := ""
	fitbList := strings.Split(theFitbStr, "_")
	for _, entry := range fitbList {
		fullAnswer = fullAnswer + entry
	}
	return fullAnswer
}

func showAskQuestionPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var questionjctid uint64
	var questionid uint64
	var topicid uint64
	var askTimeGmt uint64
	var timeInterval uint64
	var factorup float64
	var factordown float64
	var theFitbStr string
	var topicName string
	var alreadywrong int
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		panic("Parseform failed")
	}
	questionjctid = 0
	theform := r.Form
	_, questionJctExists := theform["questionjct"]
	if questionJctExists {
		questionjctid = strToUint64(theform["questionjct"][0])
	}
	db := accessdb.GetDbConnection()
	defer db.Close()
	// initialization
	responseMap := make(map[int]string)
	correctMessage := ""
	fullAnswer := ""
	alreadywrong = 0
	takeABreakMode := false
	gotRight := false
	doBreakRedirect := false
	increment := 0.6931471805599453 // log of 2
	decrement := 2.0                // divide by this instead of subtract
	currentTime := uint64(time.Now().Unix())
	_, responseExists := theform["response1"]
	if responseExists {
		alreadywrong = strToInt(theform["alreadywrong"][0])
		allCorrect := true
		questionid, topicid, askTimeGmt, timeInterval, factorup, factordown, theFitbStr, topicName = followJctToQuestionInfo(db, questionjctid, userid)
		fitbList := strings.Split(theFitbStr, "_")
		inBlank := false
		for idx, entry := range fitbList {
			fullAnswer = fullAnswer + entry
			if inBlank {
				answer := theform["response"+intToStr(idx)][0]
				answer = langtext.ProcessInternationalLanguageCharacterConversion(answer)
				responseMap[idx] = answer
				if trim(answer) != entry {
					allCorrect = false
				}
			}
			inBlank = !inBlank
		}

		seedTime := time.Now()
		seedUnix := seedTime.UnixNano()
		rand.Seed(seedUnix)
		if allCorrect {
			// Next question
			correctMessage = `<font color="green">Correct!</font> ` + fullAnswer + `</font><p><p>`
			if alreadywrong == 0 {
				// if they got it correct, we use the UP factor
				// and increase it and set the DOWN factor to
				// the new UP factor so it will be used if they
				// go down next time
				if factorup < 1.0 {
					factorup = 1.0
				}
				factorup = factorup + increment
				if timeInterval < 1 {
					timeInterval = 1
				}
				factorup = factorup * (1.0 + ((rand.Float64() - 0.5) / 25.0)) // allows 2% variation
				factordown = factorup
				timeInterval = uint64(float64(timeInterval) * factorup)
				newTime := currentTime + timeInterval
				updateInterval(db, questionjctid, newTime, timeInterval, factorup, factordown, userid)
			}
			questionjctid = 0                  // will prompt code below to retrieve a new question
			alreadywrong = 0                   // clear this flag for the next question
			responseMap = make(map[int]string) // throw away repeat of answers
			gotRight = true
		} else {
			correctMessage = `<font color="red">Wrong!</font> The answer is: ` + fullAnswer
			if alreadywrong == 0 {
				// If they got it wrong, we 1) chop the time
				// interval by the current DOWN factor first,
				// before changing it, 2) change the DOWN
				// factor (making sure it doesn't go too low),
				// and divide the UP factor by 2 making it
				// (normally) go down much faster, but we're
				// going to continue using the DOWN factor as
				// long as we're going down and only switch
				// back to the UP factor (and set the DOWN
				// factor equal to the UP factor) when we start
				// going back up.
				timeInterval = uint64(float64(timeInterval) / factordown)
				d1 := factordown - increment
				d2 := factordown / decrement
				if d2 > d1 {
					factordown = d2
				} else {
					factordown = d1
				}
				if factordown < 1.0 {
					factordown = 1.0
				}
				factordown = factordown * (1.0 + ((rand.Float64() - 0.5) / 25.0)) // allows 2% variation
				//////
				if timeInterval < 1 {
					timeInterval = 1
				}
				factorup = factorup / decrement
				if factorup < 1.0 {
					factorup = 1.0
				}
				newTime := currentTime + timeInterval
				updateInterval(db, questionjctid, newTime, timeInterval, factorup, factordown, userid)
			}
			alreadywrong = 1
		}
	}
	if questionjctid == 0 {
		// Ok! So, if there was no junction ID provided, then the use is to advance to the NEXT question!
		// If there was a topic ID from the last question, we stick with it
		// If not, we assume we are starting a new topic, and the topic ID will be provided
		if topicid == 0 {
			topicid = strToUint64(theform["topic"][0])
		}
		// Ok, first we see if there are any questions already started that are ready to be asked
		currentTime := uint64(time.Now().Unix())

		sql := "SELECT fitb_user_question_jct.id_uq_jct, fitb_question.the_fitb_str FROM fitb_user_question_jct, fitb_question WHERE (fitb_user_question_jct.id_user = ?) AND (fitb_user_question_jct.id_topic = ?) AND (fitb_user_question_jct.ask_time_gmt > 1474838209) AND (fitb_user_question_jct.ask_time_gmt <= ?) AND (fitb_user_question_jct.id_question = fitb_question.id_question) ORDER BY fitb_user_question_jct.ask_time_gmt LIMIT ?, 1;" // the time code here is the moment the program first went live

		selAlreadyAsked, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}

		keepGoing := true
		skipOffset := 0
		contemplatingFitbStr := ""
		for keepGoing {
			keepGoing = false

			selAlreadyAsked.Bind(userid, topicid, currentTime, skipOffset)
			rows, _, err := selAlreadyAsked.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Exec() failed")
			}
			questionjctid = 0 // we have to explicitly set to zero so we can detect running off the end
			for _, row := range rows {
				questionjctid = row.Uint64(0)
				contemplatingFitbStr = row.Str(1)

			}

			if calcFullAnswer(contemplatingFitbStr) == fullAnswer {
				// if we're using the same punch-out sentence, skip and find something different
				if questionjctid != 0 { // make sure we don't go into an endless loop when we go off the end
					skipOffset++
					keepGoing = true
				}
			}

		}

		// Now, if we didn't get anything from the last query, it means we are out of questions already asked, so we should retrieve a new question that has never been asked
		if questionjctid == 0 {

			// sql := "SELECT id_uq_jct FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt < 1474838209) ORDER BY ask_time_gmt LIMIT ?, 1;" // the time code here is the moment the program first went live
			sql := "SELECT fitb_user_question_jct.id_uq_jct, fitb_question.the_fitb_str FROM fitb_user_question_jct, fitb_question WHERE (fitb_user_question_jct.id_user = ?) AND (fitb_user_question_jct.id_topic = ?) AND (fitb_user_question_jct.ask_time_gmt <= 1474838209) AND (fitb_user_question_jct.id_question = fitb_question.id_question) ORDER BY fitb_user_question_jct.ask_time_gmt LIMIT ?, 1;" // the time code here is the moment the program first went live
			selNeverBeenAsked, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}

			keepGoing = true
			skipOffset = 0
			for keepGoing {
				keepGoing = false
				selNeverBeenAsked.Bind(userid, topicid, skipOffset)
				rows, _, err := selNeverBeenAsked.Exec()
				if err != nil {
					fmt.Println(err)
					panic("Exec() failed")
				}
				questionjctid = 0 // we have to explicitly set to zero so we can detect running off the end
				for _, row := range rows {
					questionjctid = row.Uint64(0)
					contemplatingFitbStr = row.Str(1)
				}

				if calcFullAnswer(contemplatingFitbStr) == fullAnswer {
					// if we're using the same punch-out sentence, skip and find something different
					if questionjctid != 0 { // make sure we don't go into an endless loop when we go off the end
						skipOffset++
						keepGoing = true
					}
				}
			}
		}
	}
	if questionjctid == 0 {
		// Still zero?? Then we have to give the user a message telling them there is nothing to learn!
		sql := "SELECT id_uq_jct FROM fitb_user_question_jct WHERE (id_user = ?) AND (id_topic = ?) AND (ask_time_gmt > 1474838209) ORDER BY ask_time_gmt LIMIT 1;" // the time code here is the moment the program first went live
		sel, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		sel.Bind(userid, topicid)
		rows, _, err := sel.Exec()
		if err != nil {
			fmt.Println(err)
			panic("Exec() failed")
		}
		var wouldBeQuestionjctid uint64
		for _, row := range rows {
			wouldBeQuestionjctid = row.Uint64(0)
		}
		questionid, topicid, askTimeGmt, timeInterval, factorup, factordown, theFitbStr, topicName = followJctToQuestionInfo(db, wouldBeQuestionjctid, userid)
		takeABreakMode = true
		if gotRight {
			doBreakRedirect = true
		}
	} else {
		questionid, topicid, askTimeGmt, timeInterval, factorup, factordown, theFitbStr, topicName = followJctToQuestionInfo(db, questionjctid, userid)
	}
	progressMessage := genProgressMessage(db, userid, topicid, currentTime)
	if doBreakRedirect {
		http.Redirect(w, r, "quiz?topic="+uint64ToStr(topicid), 302)
	} else {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		title := topicName

		displayInfo := make(map[string]string)
		displayInfo["hUserName"] = userName
		displayInfo["hTitle"] = htmlize(title)
		showFitbHeadHeader(w, displayInfo, false)

		if takeABreakMode {
			timeInterval := int(askTimeGmt - currentTime)
			inEnglish := wutil.TimeIntervalSecondsToEnglish(timeInterval)
			showFitbBodyHeader(w, displayInfo)
			fmt.Fprint(w, `
</head>
<body>
`)
			fmt.Fprint(w, `

  <section>
    <p>`+progressMessage+`</p>
    <h1>`+title+`</h1>
    <p>Time to take a break!</p>
    <p>Come back in `+inEnglish+`.</p>
<form action="quiz" name="frmQuiz" id="frmQuiz" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
<input class="infield" type="submit" value="Resume" />
</form>
  </section>
`)
			showFitbFooter(w)
		} else {
			fitbList := strings.Split(theFitbStr, "_")
			retElemNumber := 0
			inBlank := false
			for idx, _ := range fitbList {
				if inBlank {
					retElemNumber = idx
				}
				inBlank = !inBlank
			}
			fmt.Fprint(w, `
<script>

function advanceOnReturn(ev, num) {
    if (ev.keyCode == 13) {
        ev.stopPropagation();
        if (num == `+intToStr(retElemNumber)+`) {
            document.forms.frmQuiz.submit();
        } else {
            nxt = num + 2
            document.getElementById("response" + nxt).focus();
        }
        return false;
    } else {
        return false;
    }
}

</script>
</head>
<body onload="document.getElementById('response1').focus();">
`)
			showFitbBodyHeader(w, displayInfo)
			fmt.Fprint(w, `
  <section>
    <p>`+progressMessage+`</p>
    <h1>`+title+`</h1>
    <p>`+correctMessage+`</p>
<form action="quiz" name="frmQuiz" id="frmQuiz" method="post" onsubmit="return false;">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
<input type="hidden" name="questionjct" value="`+uint64ToStr(questionjctid)+`" />
<input type="hidden" name="_questionid" value="`+uint64ToStr(questionid)+`" />
<input type="hidden" name="alreadywrong" value="`+intToStr(alreadywrong)+`" />
`)
			inClear := true
			for idx, entry := range fitbList {
				if inClear {
					// fmt.Fprint(w, entry)
					io.WriteString(w, entry)
				} else {
					_, respEntryExists := responseMap[idx]
					if !respEntryExists {
						responseMap[idx] = ""
					}
					ixAsStr := intToStr(idx)
					io.WriteString(w, `<input class="infield" name="response`+ixAsStr+`" id="response`+ixAsStr+`" value="`+responseMap[idx]+`" onkeypress="advanceOnReturn(event, `+ixAsStr+`);" />`)
				}
				inClear = !inClear
			}
			fmt.Fprint(w, `
</form>
  </section>
`)
			showFitbFooter(w)
		}
	}
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	// fmt.Println("In FITB handler, running op:", op)
	switch {
	case op == "listtopics":
		if userid == 1 {
			showTopicListPage(w, r, op, userid, userName)
		}
	case op == "edittopic":
		if userid == 1 {
			showEditTopicPage(w, r, op, userid, userName)
		}
	case op == "picktopic":
		if userid == 1 {
			showTopicPickListPage(w, r, op, userid, userName)
		}
	case op == "listchapters":
		if userid == 1 {
			showChapterListPage(w, r, op, userid, userName)
		}
	case op == "editchapter":
		if userid == 1 {
			showEditChapterPage(w, r, op, userid, userName)
		}
	case op == "bulkeditquestions":
		if userid == 1 {
			showBulkEditQuestionsPage(w, r, op, userid, userName)
		}
	case op == "listquestions":
		if userid == 1 {
			showListQuestionsPage(w, r, op, userid, userName)
		}
	case op == "bulkaddquestions":
		if userid == 1 {
			showAddBulkQuestionsPage(w, r, op, userid, userName)
		}
	case op == "editquestion":
		if userid == 1 {
			showEditQuestionPage(w, r, op, userid, userName)
		}
	case op == "deletequestion":
		if userid == 1 {
			showDeleteQuestionPage(w, r, op, userid, userName)
		}
	case op == "renumber":
		if userid == 1 {
			showRenumberPage(w, r, op, userid, userName)
		}
	case op == "initialize":
		if userid == 1 {
			showInitializePage(w, r, op, userid, userName)
		}
	case op == "quiz":
		if userid == 1 {
			showAskQuestionPage(w, r, op, userid, userName)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/fitb/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
