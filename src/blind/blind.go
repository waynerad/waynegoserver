package blind
// set ID is hardcoded -- need to fix

import (
	"accessdb"
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
)

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
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
		laplaceNum := (numerator + (3.0 * popFactor)) / (float64(number) + popFactor)
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

func showBlindMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<p><a href="scrape">Scrape</a> &middot; `+htm(userName)+`</p>
`)
}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
	writeaccess := false
	if userid == 1 {
		writeaccess = true
	}
	if writeaccess {
		showBlindMenuBar(w, userName)
	}
	fmt.Fprint(w, `
    <h1>List of Books</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
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
	var ord uint64
	ord = 1
	_, setExists = getform["ord"]
	if setExists {
		ord, err = strconv.ParseUint(getform["ord"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
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
	sql = ""
	if ord == 1 {
		sql = "SELECT id_book, title, authors, stars5, stars4, stars3, stars2, stars1, laplace FROM blind_book WHERE (id_user = ?) AND (id_set = ?) ORDER BY laplace DESC;"
	}
	if ord == 2 {
		sql = "SELECT id_book, title, authors, stars5, stars4, stars3, stars2, stars1, laplace FROM blind_book WHERE (id_user = ?) AND (id_set = ?) ORDER BY id_book DESC;"
	}
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
`)
			if writeaccess {
				fmt.Fprint(w, `<th> Actions </th>`)
			}
			fmt.Fprint(w, `
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

		fmt.Fprint(w, `<td align="right">`+strconv.FormatUint(book.stars5, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars4, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars3, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars2, 10)+` </td><td align="right">`+strconv.FormatUint(book.stars1, 10)+` </td><td align="right">`+floattostr(average)+` </td><td align="right">`+floattostr(book.laplace)+` </td>`)
		if writeaccess {
			fmt.Fprint(w, `<td> <a href="delete?book=`+uint64tostr(book.idBook)+`">Delete</a> </td>`)
		}
		fmt.Fprint(w, `
</tr>
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
	fmt.Println("stars to parse:", stars[idx1+10:idx2])
	pct, err := strconv.ParseFloat(stars[idx1+10:idx2], 64)
	if err != nil {
		return 0.0, err
	}
	return pct, nil
}

func round(x float64) int {
	// fmt.Println("round: ", x)
	// fmt.Println("    floor: ", math.Floor(x+0.5))
	// fmt.Println("    to int: ", int(math.Floor(x+0.5)))
	return int(math.Floor(x + 0.5))
}

func showScrapePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
			db := accessdb.GetDbConnection()
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
		byIdx := strings.Index(ui.titleauthors, "by ")
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
			anotherIdx = strings.Index(ui.titleauthors, "Mass Market")
			if anotherIdx != -1 {
				if anotherIdx < ttlIdx {
					ttlIdx = anotherIdx
				}
			}
			title = strings.Trim(ui.titleauthors[:ttlIdx], " \r\n\t")
			authors = strings.Trim(ui.titleauthors[byIdx+3:], " \r\n\t")
			fmt.Println("authors", authors)
			outOfAuthorsIdx := strings.Index(authors, "out of")
			fmt.Println("outOfAuthorsIdx", outOfAuthorsIdx)
			var reviewsCount int64
			if outOfAuthorsIdx == -1 {
				errorList["titleauthors"] = "Could not find out of in authors field."
				errorOccurred = true
			} else {
				customerIdx := strings.Index(authors, "customer")
				fmt.Println("customerIdx", customerIdx)
				fmt.Println("string to convert", strings.Trim(authors[outOfAuthorsIdx+14:customerIdx-1], " \r\n\t"))
				reviewsCount, err = strconv.ParseInt(strings.Trim(authors[outOfAuthorsIdx+14:customerIdx-1], " \r\n\t"), 10, 64)
				if err != nil {
					errorList["stars"] = "The reviews count was found but the found count failed to parse."
					errorOccurred = true
				}
				fmt.Println("reviewsCount is", reviewsCount)
				authors = strings.Trim(authors[0:outOfAuthorsIdx-4], " \r\n\t")
				fmt.Println("authors (final?)", authors)
			}
			outOfStarsIdx := strings.Index(ui.stars, "out of")
			if outOfStarsIdx == -1 {
				errorList["stars"] = "Could not find out of in stars field."
				errorOccurred = true
			} else {
				starsAfterOutOf := ui.stars[outOfStarsIdx+14:]
				st5pct, pct5err := parseStarsPct(starsAfterOutOf, "5 star")
				st4pct, pct4err := parseStarsPct(starsAfterOutOf, "4 star")
				st3pct, pct3err := parseStarsPct(starsAfterOutOf, "3 star")
				st2pct, pct2err := parseStarsPct(starsAfterOutOf, "2 star")
				st1pct, pct1err := parseStarsPct(starsAfterOutOf, "1 star")
				if (pct5err != nil) || (pct4err != nil) || (pct3err != nil) || (pct2err != nil) || (pct1err != nil) {
					errorList["stars"] = "One of the stars percentages failed to parse."
					errorOccurred = true
				} else {
					// 	seeAllIdx := strings.Index(ui.stars, "See all")
					// 	reviewsIdx := strings.Index(ui.stars, "reviews")
					// 	if (seeAllIdx == -1) || (reviewsIdx == -1) {
					// 		errorList["stars"] = "The reviews count failed to parse."
					// 		errorOccurred = true
					// 	} else {
					// 		reviewsCount, err := strconv.ParseInt(ui.stars[seeAllIdx+8:reviewsIdx-1], 10, 64)
					// 		if err != nil {
					// 			errorList["stars"] = "The reviews count was found but the found count failed to parse."
					// 			errorOccurred = true
					// 		}
					rc := float64(reviewsCount)
					stars5 = round((st5pct / 100.0) * rc)
					fmt.Println("stars5", stars5)
					stars4 = round((st4pct / 100.0) * rc)
					fmt.Println("stars4", stars4)
					stars3 = round((st3pct / 100.0) * rc)
					fmt.Println("stars3", stars3)
					stars2 = round((st2pct / 100.0) * rc)
					fmt.Println("stars2", stars2)
					stars1 = round((st1pct / 100.0) * rc)
					fmt.Println("stars1", stars1)
					// 	}
				}
			}
		}
		if errorOccurred {
			showform = true
		} else {
			// dbConnect!!
			db := accessdb.GetDbConnection()
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
			save.idSet = 12 // hardcoded for now
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
			setid := 12 // BUGBUG this is hardcoded but shouldn't be
			http.Redirect(w, r, "list?set="+inttostr(setid), 302)
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		db := accessdb.GetDbConnection()
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

func showDeletePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	if userid != 1 {
		panic("delete page reached without access")
	}
	showform := false
	errorList := make(map[string]string)
	errorOccurred := false
	method := r.Method
	var bookid uint64
	bookid = 0
	var title string
	var authors string
	if method == "GET" {
		// set defaults
		showform = true
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			panic("parseform failed")
		}
		getform := r.Form
		_, bookspecified := getform["book"]
		if !bookspecified {
			fmt.Println("no book specified on book delete form")
			panic("parseform failed")
		}
		bookid, err = strconv.ParseUint(getform["book"][0], 10, 64)
		if err != nil {
			fmt.Println(err)
			panic("ParseUint failed")
		}
		db := accessdb.GetDbConnection()
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
			title = row.Str(0)
			authors = row.Str(1)
		}
	}
	if method == "POST" {
		fmt.Println("point 830")
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
			fmt.Println(err)
			panic("bookid parse failed")
		}
		db := accessdb.GetDbConnection()
		defer db.Close()
		// query, if there, update, if not, create new
		fmt.Println("id_book", bookid)
		fmt.Println("id_user", userid)
		stmt, err := db.Prepare("DELETE FROM blind_book WHERE (id_book = ?) AND (id_user = ?);")
		if err != nil {
			fmt.Println(err)
			panic("Prepare failed")
		}
		stmt.Bind(bookid, userid)
		_, _, err = stmt.Exec()
		setid := 12 // BUGBUG this is hardcoded but shouldn't be
		http.Redirect(w, r, "list?set="+inttostr(setid), 302)
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		db := accessdb.GetDbConnection()
		defer db.Close()
		fmt.Fprint(w, `<title>Delete Blind Book</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
`)
		showBlindMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Delete Blind Book</h1>
`)
		fmt.Fprint(w, `
<form action="delete" method="post">
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
<p>
Do you want to delete `+html.EscapeString(title)+` by `+html.EscapeString(authors)+`
</p><p>
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
	fmt.Println("We are in the blind handler, and op is", op)
	switch {
	case op == "scrape":
		if userid != 0 {
			showScrapePage(w, r, op, userid, userName)
		} else {
			redirectToLoginPage(w, r)
		}
	case op == "list":
		if userid != 0 {
			showListPage(w, r, op, userid, userName)
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
		filename := "/home/ec2-user/wayneserver/staticappcontent/blind/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
