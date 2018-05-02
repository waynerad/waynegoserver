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

func candeleteshowTopicPickListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Pick Topic</h1>
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
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th></tr>`)
			started = true
		}
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"practice?topic=`+uint64ToStr(topic.idTopic)+`">`+htmlize(topic.name)+"</a> </td>")
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

func candeleteshowPracticePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
	fmt.Fprint(w, getDoctype())
	db := accessdb.GetDbConnection()
	defer db.Close()
	fmt.Fprint(w, `<title>Practice</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Practice: `+htmlize(ui.name)+`</h1>
  </section>
</body>
</html>`)
}

func candeleteshowEditChapterPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
		fmt.Fprint(w, getDoctype())
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>Chapter Entry</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Chapter Entry</h1>
<form action="editchapter" method="post">
<input type="hidden" name="chapter" value="`+uint64ToStr(chapterid)+`" />
<input class="biginput" name="topic" id="topic" type="hidden" value="`+uint64ToStr(ui.idTopic)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<table border="0" cellpadding="4">
<tr><td align="right"> Name: </td><td> <input class="biginput" name="name" id="name" type="text" value="`+htmlize(ui.name)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func candeleteshowChapterListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Chapters of `+htmlize(topicName)+`</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Chapters of `+htmlize(topicName)+`</h1>
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
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> <a href="+`"editchapter?chapter=`+uint64ToStr(chapter.idChapter)+`">`+htmlize(chapter.name)+"</a> </td>")
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
</body>
</html>`)
}

func candeleteshowBulkEditQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
				currentSave.lnum = 5 // calculateLnum(currentSave.theFitbStr)
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
		fmt.Fprint(w, getDoctype())
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
		fmt.Fprint(w, `<title>List questions for topic: `+topicName+`</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>List questions for topic: `+topicName+`</h1>
<form action="bulkeditquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
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
<tr><td> <input type="hidden" name="id`+intToStr(count)+`" id="id" value="`+uint64ToStr(currentQuest.idQuestion)+`" /><input type="text" name="seq`+intToStr(count)+`" id="seq`+intToStr(count)+`" value="`+int64ToStr(currentQuest.seqNum)+`" /> </td><td> <select name="chapter`+intToStr(count)+`" id="chapter`+intToStr(count)+`">
`)
			for _, currentChapt := range allChapters {
				fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
				if currentChapt.idChapter == currentQuest.idChapter {
					fmt.Fprint(w, ` selected="selected"`)
				}
				fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
			}
			fmt.Fprint(w, `</select> </td><td> <input type="text" name="quest`+intToStr(count)+`" id="quest`+intToStr(count)+`" value="`+htmlize(currentQuest.theFitbStr)+`" /> </td><td align="right"> `+floatToStr(currentQuest.lnum)+` </td></tr>
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
<tr><td> <input type="hidden" name="id`+intToStr(count)+`" id="id" value="0" /><input type="text" name="seq`+intToStr(count)+`" id="seq`+intToStr(count)+`" value="`+int64ToStr(currentSeqNum)+`" /> </td><td> <select name="chapter`+intToStr(count)+`" id="chapter`+intToStr(count)+`">
`)
			for _, currentChapt := range allChapters {
				fmt.Fprint(w, `<option value="`+uint64ToStr(currentChapt.idChapter)+`"`)
				if currentChapt.idChapter == lstChapt {
					fmt.Fprint(w, ` selected="selected"`)
				}
				fmt.Fprint(w, `>`+htmlize(currentChapt.name)+`</option>`)
			}
			fmt.Fprint(w, `</select> </td><td> <input type="text" name="quest`+intToStr(count)+`" id="quest`+intToStr(count)+`" value="" /> </td><td> &nbsp; </td> </tr>
`)
		}
		fmt.Fprint(w, `
<tr><td colspan="3" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func candeleteshowListQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
		fmt.Fprint(w, getDoctype())
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
		fmt.Fprint(w, `<title>Edit questions for topic: `+topicName+`</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Edit questions for topic: `+topicName+`</h1>
<form action="listquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<table border="0" cellpadding="4">
<tr><th> Sequence number </th><th> Chapter <th> Question in FITB format </th><th> Size </th></tr>
`)
		for _, currentQuest := range allQuests {
			fmt.Fprint(w, `
<tr><td> `+int64ToStr(currentQuest.seqNum)+` </td><td> `+allChapters[currentQuest.idChapter]+` </td> <td> `+htmlize(currentQuest.theFitbStr)+` </td><td align="right"> `+floatToStr(currentQuest.lnum)+` </td><td> <a href="editquestion?question=`+uint64ToStr(currentQuest.idQuestion)+`">Edit</a> </td> </tr>
`)
		}
		fmt.Fprint(w, `
<tr><td> <a href="bulkaddquestions?topic=`+uint64ToStr(topicid)+`">Add Questions</a> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func candeleteshowAddBulkQuestionsPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
				if trEnt != "" {
					var currentSave saveTyp
					currentSave.idTopic = topicid
					currentSave.idChapter = chapterid
					currentSave.sequenceNum = seqNum
					currentSave.theFitbStr = trEnt
					currentSave.lnum = 5 // calculateLnum(currentSave.theFitbStr)
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
		fmt.Fprint(w, getDoctype())
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
		fmt.Fprint(w, `<title>Bulk add questions for topic: `+topicName+`</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Bulk add questions for topic: `+topicName+`</h1>
<form action="bulkaddquestions" method="post">
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
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
<p><input type="submit" /></p>
</form>
  </section>
</body>
</html>`)
	}
}

func candeleteshowEditQuestionPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var topicid uint64
	topicid = 0
	topicName := ""
	var questionid uint64
	questionid = 0
	type questionType struct {
		idQuestion uint64
		idChapter  uint64
		seqNum     int64
		theFitbStr string
		lnum       float64
	}
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
			save.lnum = 5 // calculateLnum(save.theFitbStr)
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
			http.Redirect(w, r, "listtopics?topic="+uint64ToStr(topicid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
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
		fmt.Fprint(w, `<title>Edit questions for topic: `+topicName+`</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Edit questions for topic: `+topicName+`</h1>
<form action="editquestion" method="post">
<input type="hidden" name="question" value="`+uint64ToStr(questionid)+`" />
<input type="hidden" name="topic" value="`+uint64ToStr(topicid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
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
<tr><td> Question in FITB format <br /><textarea name="fitb" id="fitb" rows="8" cols="80">`+htmlize(ui.theFitbStr)+`</textarea></td></tr>
</table>
`)
		fmt.Fprint(w, `
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
}

func candeleteinitialize(db mysql.Conn, userid uint64, topicid uint64) {
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
	for _, questionid := range questionList {
		sql = "INSERT INTO fitb_user_question_jct (id_user, id_question, id_topic, ask_time_gmt, time_interval, factorup, factordown) VALUES (?, ?, ?, ?, ?, ?, ?);"
		stmt, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		sequenceNum := seqNumMap[questionid]
		askTime := 1 + sequenceNum                                    // This is so the sequence of first introduction is the same as the sequence_num bers in the database
		stmt.Bind(userid, questionid, topicid, askTime, 60, 1.0, 1.0) // INITIAL REPETITIVENESS is set here
		_, _, err = stmt.Exec()
	}
}

func candeletefollowJctToQuestionInfo(db mysql.Conn, questionjctid uint64, userid uint64) (uint64, uint64, uint64, uint64, float64, float64, string, string) {
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

func candeleteupdateInterval(db mysql.Conn, questionjctid uint64, askTime uint64, timeInterval uint64, factorup float64, factordown float64, userid uint64) {
	sql := "UPDATE fitb_user_question_jct SET ask_time_gmt = ?, time_interval = ?, factorup = ?, factordown = ? WHERE (id_uq_jct = ?) AND (id_user = ?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	stmt.Bind(askTime, timeInterval, factorup, factordown, questionjctid, userid)
	_, _, err = stmt.Exec()
}

// show time interval, e.g. "5 sec", "14 min", etc
func candeletetimeIntervalSecondsToEnglish(inv int) string {
	if inv < 0 {
		return "0 seconds"
	}
	increments := [5]int{60, 60, 24, 7, 52}
	singulars := [6]string{"second", "minute", "hour", "day", "week", "year"}
	plurals := [6]string{"seconds", "minutes", "hours", "days", "weeks", "years"}
	var breakdown [7]int
	var num int
	for i, x := range increments {
		num = inv / x // has to be integer division
		rem := inv - (num * x)
		breakdown[i] = rem
		inv = num
	}
	breakdown[6] = num
	toUse := 2
	isLeading := true
	rv := ""
	for i := 5; i >= 0; i-- {
		if isLeading {
			if breakdown[i] != 0 {
				isLeading = false
			}
		}
		if !isLeading {
			if toUse != 0 {
				if breakdown[i] == 1 {
					rv = rv + ", 1 " + singulars[i]
				} else {
					if breakdown[i] != 0 {
						rv = rv + ", " + intToStr(breakdown[i]) + " " + plurals[i]
					}
				}
				toUse--
			}
		}
	}
	return rv[1:]
}

func candeletegetCountFromSQLForUserTopicAndTime(db mysql.Conn, sql string, userid uint64, topicid uint64, currentTime uint64, bindCurrentTime bool) int {
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
	title       string
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
		fmt.Fprint(w, "<tr "+backgroundColor+"><td> "+htmlize(streakTaskXYZ.title)+" </td><td> "+`<a href="edittask?task=`+streakAsStr+`">Edit</a> </td>
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
	// | title   | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry taskEntryData
	theList := make(taskListData, 0)
	sql = "SELECT id_task, title, description, cycle_days FROM streak_task WHERE (id_user = ?) ORDER BY id_task;"
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
		currentEntry.title = row.Str(1)
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
	// | title   | varchar(255)        | NO   |     |         |                |
	// +---------+---------------------+------+-----+---------+----------------+
	var sql string
	var currentEntry taskEntryData
	rv := make(map[string]string)
	sql = "SELECT id_task, title, description, cycle_days FROM streak_task WHERE (id_task = ?) AND (id_user = ?);"
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
		currentEntry.title = row.Str(1)
		currentEntry.description = row.Str(2)
		currentEntry.cycleDays = row.Int(3)
		rv["task"] = uint64ToStr(taskid)
		rv["title"] = currentEntry.title
		rv["description"] = currentEntry.description
		rv["cycle_days"] = intToStr(currentEntry.cycleDays)
		found = true
	}
	if !found {
		rv["task"] = uint64ToStr(taskid)
		rv["title"] = ""
		rv["description"] = ""
		rv["cycle_days"] = "1"
	}
	return rv
}

func (self *taskEditForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	// getStreakTaskEditDBData(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) void
	showTaskEditForm(w, r, op, userInfo, errorList, userInput)
}

func (self *taskEditForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
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
	var streakTaskRecord struct {
		taskId      uint64
		title       string
		description string
		cycleDays   int
	}
	streakTaskRecord.taskId = strToUint64(userInput["task"])
	streakTaskRecord.title = userInput["task"]
	streakTaskRecord.description = userInput["task"]
	streakTaskRecord.cycleDays = strToInt(userInput["task"])

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
	if alreadyExists {
		stmt, err := db.Prepare("UPDATE streak_task SET title = ?, description = ?, cycle_days = ?  WHERE (id_task = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(streakTaskRecord.title, streakTaskRecord.description, streakTaskRecord.cycleDays, streakTaskRecord.taskId, userInfo.UserId)
		_, _, err = stmt.Exec()
	} else {
		stmt, err := db.Prepare("INSERT INTO streak_task (id_user, title, description, cycle_days) VALUES (?, ?, ?, ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		// defer stmt.Close();
		stmt.Bind(userInfo.UserId, streakTaskRecord.title, streakTaskRecord.description, streakTaskRecord.cycleDays)
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

func candeletegetUserInput(r *http.Request) map[string]string {
	rv := make(map[string]string)
	theForm := r.Form
	for a, b := range theForm {
		if len(b) == 1 {
			rv[a] = b[0]
		} else {
			lb := len(b)
			for i := 0; i < lb; i++ {
				rv[a+"_"+intToStr(i)] = b[i]
			}
		}
	}
	return rv
}

func candeleteHandleForm(formObject forms.SimpleWebForm, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
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
		userInput = candeletegetUserInput(r)
		userInput = formObject.GetDefaults(dbConn, userInfo, userInput)
	}
	if method == "POST" {
		if errorOccurred {
			showform = true
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

func showTaskListPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject taskListForm
	formObject.objectName = "Task List Form"
	forms.HandleForm(&formObject, w, r, op, userInfo)
}

func showTaskEditPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject taskEditForm
	formObject.objectName = "Task Edit Form"
	forms.HandleForm(&formObject, w, r, op, userInfo)
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

func candeleteAgainshowTaskEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var ui struct {
		idStreak    uint64
		idUser      uint64
		name        string
		description string
	}
	var streakid uint64
	streakid = 0
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
		_, streakExists := getform["streak"]
		if streakExists {
			streakid, err = strconv.ParseUint(getform["streak"][0], 10, 64)
			if err != nil {
				fmt.Println(err)
				panic("ParseUint failed")
			}
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_streak, name, description FROM streak_streak WHERE (id_streak = ?) AND (id_user = ?);"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind(streakid, userid)
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			for _, row := range rows {
				ui.idStreak = row.Uint64(0)
				ui.name = row.Str(1)
				ui.description = row.Str(2)
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
		streakid, err = strconv.ParseUint(postform["streak"][0], 10, 64)
		if err != nil {
			streakid = 0
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
		if errorOccurred {
			showform = true
		} else {

			db := accessdb.GetDbConnection()
			defer db.Close()
			var save struct {
				idStreak    uint64
				idUser      uint64
				name        string
				description string
			}
			save.idStreak = streakid
			save.idUser = userid
			save.name = ui.name
			save.description = ui.description
			alreadyExists := false
			if streakid > 0 {
				sql := "SELECT id_streak FROM streak_streak WHERE (id_streak = ?) AND (id_user = ?);"
				sel, err := db.Prepare(sql)
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				sel.Bind(streakid, userid)
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
				stmt, err := db.Prepare("UPDATE streak_streak SET name = ?, description = ?  WHERE (id_streak = ?) AND (id_user = ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				stmt.Bind(save.name, save.description, streakid, userid)
				_, _, err = stmt.Exec()
			} else {
				stmt, err := db.Prepare("INSERT INTO streak_streak (id_streak, id_user, name, description) VALUES (?, ?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
					panic("Prepare failed")
				}
				// defer stmt.Close();
				stmt.Bind(save.idStreak, save.idUser, save.name, save.description)
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
		fmt.Fprint(w, `<title>Streak Entry</title>
`+getStyle()+`
</head>
<body>
  <section>
    <h1>Streak Entry</h1>
<form action="editstreak" method="post">
<input type="hidden" name="streak" value="`+uint64ToStr(streakid)+`" />
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmlize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `
<table border="0" cellpadding="4">
<tr><td align="right"> Name: </td><td> <input class="biginput" name="name" id="name" type="text" value="`+htmlize(ui.name)+`" /> </td></tr>
<tr><td align="right"> Description: </td><td> <input class="biginput" name="description" id="description" type="text" value="`+htmlize(ui.description)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input type="submit"> </td></tr>
</table>
</form>
  </section>
</body>
</html>`)
	}
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
