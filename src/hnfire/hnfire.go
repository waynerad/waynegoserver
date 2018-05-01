package hnfire

import (
	"accessdb"
	"bufio"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"io"
	"net/http"
	"os"
	"static"
	"strconv"
	"strings"
	"time"
)

// "math"

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

func intToStr(z int) string {
	return strconv.FormatInt(int64(z), 10)
}

func htm(z string) string {
	return html.EscapeString(z)
}

func int64ToStr(ii int64) string {
	return strconv.FormatInt(ii, 10)
}

func strToInt(str string) int {
	var rv int64
	rv, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fmt.Println("ParseInt failed")
		fmt.Println(err)
		panic("ParseInt failed")
	}
	return int(rv)
}

func readline(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	// if err != nil {
	// panic("could not read line")
	// }
	return string(ln), err
}

func fetchfile(remoteurl string, localfile string) {
	out, err := os.Create(localfile)
	if err != nil {
		fmt.Println("os.Create failed")
		fmt.Println(err)
		panic("os.Create failed")
	}
	defer out.Close()
	resp, err := http.Get(remoteurl)
	if err != nil {
		fmt.Println("http.Get failed")
		fmt.Println(err)
		panic("http.Get failed")
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("io.Copy failed")
		fmt.Println(err)
		panic("io.Copy failed")
	}
}

func debracket(numstr string) string {
	i := strings.Index(numstr, "[")
	j := strings.Index(numstr, "]")
	if i >= 0 {
		return numstr[i+1:]
	} else {
		if j >= 0 {
			return numstr[0:j]
		} else {
			return numstr
		}
	}
}

func itemExistsInDb(db mysql.Conn, itemNum int) bool {
	sql := "SELECT id_story FROM hnfire_story WHERE item_num = ?;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(uint64(itemNum))
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var storyId uint64
	for _, row := range rows {
		storyId = row.Uint64(0)
	}
	if storyId > 0 {
		return true
	}
	return false
}

func isNumeric(ch byte) bool {
	numchars := "0123456789"
	i := strings.Index(numchars, string(ch))
	return i >= 0
}

func parseStoryFile(storyfile string, itemNum int) map[string]string {
	result := make(map[string]string)
	fhOriginal, err := os.Open(storyfile)
	if err != nil {
		fmt.Printf("Not found: %s", storyfile)
		panic("Could not open story file")
		return result
	}
	defer fhOriginal.Close()
	fhIn := bufio.NewReader(fhOriginal)
	for err == nil {
		var inLine string
		inLine, err = readline(fhIn)
		lL := len(inLine)
		var inBracket bool
		var jstr string
		var nameStr string
		for i := 0; i < lL; i++ {
			if inLine[i] == '{' {
				inBracket = true
			}
			if inLine[i] == '}' {
				result[nameStr] = jstr
				inBracket = false
			}
			if inLine[i] == '"' {
				if inBracket {
					j := strings.Index(inLine[i+1:], `"`) + i
					jstr = inLine[i+1 : j+1]
					i = j + 1
				}
			}
			if isNumeric(inLine[i]) {
				j := i + 1
				for isNumeric(inLine[j]) {
					j++
				}
				jstr = inLine[i:j]
				i = j
			}
			if inLine[i] == ':' {
				nameStr = jstr
			}
			if inLine[i] == ',' {
				result[nameStr] = jstr
			}
		}
	}
	return result
}

func storeStory(db mysql.Conn, itemNum int, storyAttributes map[string]string) {
	var rec struct {
		storyId uint64
		itemNum uint64
		timeNum uint64
		url     string
		title   string
	}
	// fmt.Println("------storyAttributes")
	// for y, z := range storyAttributes {
	//	fmt.Println(y, z)
	// }
	// fmt.Println("------")
	// return
	rec.itemNum = uint64(itemNum)
	sql := "SELECT id_story FROM hnfire_story WHERE item_num = ?;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(uint64(itemNum))
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var storyId uint64
	for _, row := range rows {
		storyId = row.Uint64(0)
	}
	if storyId > 0 {
		return // already stored, don't store again
	}
	rec.storyId = storyId
	rec.timeNum = uint64(strToInt(storyAttributes["time"]))
	rec.url = storyAttributes["url"]
	rec.title = storyAttributes["title"]
	sql = "INSERT INTO hnfire_story (item_num, item_time, url, title) VALUES(?, ?, ?, ?);"
	stmt, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	// defer stmt.Close();
	stmt.Bind(rec.itemNum, rec.timeNum, rec.url, rec.title)
	_, _, err = stmt.Exec()
}

// exported so runnable as a cron
func Pull(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	dopull := true
	outputToBrowser := false
	if op != "cron" {
		outputToBrowser = true
	}
	if outputToBrowser {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		title := "Pulling Firebase"
		fmt.Fprint(w, `<title>`+title+`</title>
`+getStyle()+`
</head>
<body>
<section>
<h1>`+title+`</h1>
<p>Pulling Firebase</p>
`)
	} else {
		current := time.Now()
		fmt.Println("Firebase pull at: ", current)
	}
	newstoriesURL := "https://hacker-news.firebaseio.com/v0/newstories.json"
	newstoriesFile := "/home/ec2-user/hnnewstories.json"
	if dopull {
		fetchfile(newstoriesURL, newstoriesFile)
	}
	count := 0
	db := accessdb.GetDbConnection()
	defer db.Close()
	fhOriginal, err := os.Open(newstoriesFile)
	if err != nil {
		fmt.Printf("Not found: %s", newstoriesFile)
		return
	}
	defer fhOriginal.Close()
	fhIn := bufio.NewReader(fhOriginal)
	for err == nil {
		var inLine string
		inLine, err = readline(fhIn)
		stuff := strings.Split(inLine, ",")
		for _, itemnumber := range stuff {
			// hacky parser, assumes source is well-formed
			itemnumber = debracket(itemnumber)
			if outputToBrowser {
				fmt.Fprint(w, itemnumber+"<br />")
			} else {
				fmt.Println(itemnumber)
			}
			if itemnumber != "" {
				itemNum := strToInt(itemnumber)
				if !itemExistsInDb(db, itemNum) {
					storyurl := "https://hacker-news.firebaseio.com/v0/item/" + itemnumber + ".json"
					storyfile := "/home/ec2-user/hnstories/story" + itemnumber + ".json"
					if dopull {
						fetchfile(storyurl, storyfile)
					}
					storyAttributes := parseStoryFile(storyfile, itemNum)
					_, ok := storyAttributes["time"]
					if !ok {
						storyAttributes["time"] = "0"
						storyAttributes["url"] = ""
						storyAttributes["title"] = ""
					}
					storeStory(db, itemNum, storyAttributes)
					count++
				}
			}
		}
	}
	if outputToBrowser {
		fmt.Fprint(w, `
<p> Added <b>`+intToStr(count)+`</b> stories.</p>
</section>
</body>
</html>`)
	} else {
		fmt.Println("Added " + intToStr(count) + " stories.")
	}
}

func getLastId(db mysql.Conn) int {
	sql := "SELECT MAX(id_story) FROM hnfire_story WHERE 1;"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	// Nothing to bind here
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var storyId uint64
	for _, row := range rows {
		storyId = row.Uint64(0)
	}
	return int(storyId)
}

func showMatchingArticles(db mysql.Conn, searchTerm string, fromEntry int, w http.ResponseWriter) {
	matchCase := false
	if searchTerm == strings.ToUpper(searchTerm) {
		matchCase = true
	}
	sql := "SELECT url, title FROM hnfire_story WHERE (id_story > ?) AND (title LIKE ?);"
	sel, err := db.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		panic("Prepare failed")
	}
	sel.Bind(uint64(fromEntry), "%"+searchTerm+"%")
	rows, _, err := sel.Exec()
	if err != nil {
		fmt.Println(err)
		panic("Bind/Exec failed")
	}
	var rec struct {
		url   string
		title string
	}
	for _, row := range rows {
		rec.url = row.Str(0)
		rec.title = row.Str(1)
		if matchCase {
			i := strings.Index(rec.title, searchTerm)
			if i >= 0 {
				fmt.Fprint(w, "<p>"+htm(rec.title)+"<br />"+`<a href="`+rec.url+`">`+htm(rec.url)+"</a></p>")
			}
		} else {
			fmt.Fprint(w, "<p>"+htm(rec.title)+"<br />"+`<a href="`+rec.url+`">`+htm(rec.url)+"</a></p>")
		}
	}
}

func search(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	title := "Searching Firebase"
	fmt.Fprint(w, `<title>`+title+`</title>
`+getStyle()+`
</head>
<body>
<section>
<h1>`+title+`</h1>
`)
	db := accessdb.GetDbConnection()
	lastId := getLastId(db)
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("parseform failed")
	}
	postform := r.Form
	fromEntry := lastId - 500
	_, ok := postform["from"]
	if ok {
		fromEntry = strToInt(postform["from"][0])
	}
	fmt.Fprint(w, "<p>Last ID: "+intToStr(lastId)+"</p>")
	fmt.Fprint(w, "<p>Displaying from: "+intToStr(fromEntry)+"</p>")
	terms := [69]string{"emotion", "tensorflow", "pytorch", "deepmind", "deep learn", "neura", "neuro", "robo", "autonomous", "automate", "artificial", "algorithm", "machine learn", "reinforcement learn", "AI", "brain", "intelligence", "IQ", "data science", "nanometer", "moore", "google", "microsoft", "amazon", "facebook", "apple", "nvidia", "futur", "china", "chines", "snapchat", "evolution", "genetic", "dna", "crypt", "bill gates", "carmack", "elon musk", "golang", "python", "video game", "music", "porn", "police", "kill", "computer science", "startup", "silicon valley", "math", "emotion", "fermi", "female", "digital", "salary", "hire", "hiring", "dark", "paul graham", "jessica livingston", "sam altman", "damore", "privilege", "VR", "AR", "fast", "speed", "medical", "youtube", "Go"}
	for idx := 0; idx < 63; idx++ {
		// fmt.Fprint(w, intToStr(idx)+": "+terms[idx]+"<br />")
		fmt.Fprint(w, "<h1>"+terms[idx]+"</h1>")
		showMatchingArticles(db, terms[idx], fromEntry, w)
	}
	fmt.Fprint(w, `
<p>Done</p>
</section>
</body>
</html>`)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	// fmt.Println("in hnfire handler, op is", op)
	switch {
	case op == "pull":
		if userid == 1 {
			Pull(w, r, op, userid, userName)
		}
	case op == "search":
		if userid == 1 {
			search(w, r, op, userid, userName)
		}

	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/hnfire/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
