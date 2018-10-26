package links

import (
	"accessdb"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"net/http"
	"os"
	"static"
	"strconv"
	"strings"
	"time"
)

type lnksqlfields struct {
	linkId      uint64
	createdGMT  uint64
	targetURL   string
	imageURL    string
	description string
	myComment   string
	title       string
}

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

.bigtextarea {
    font-size: 1.1em;
}

</style>

`
}

func htmize(str string) string {
	return html.EscapeString(str)
}

func isNumeric(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return true
	}
	return false
}

func strToFloat(str string) float64 {
	flt, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fmt.Println(err)
		panic("strToFloat failed")
	}
	return flt
}

func floatToStr(flt float64) string {
	str := strconv.FormatFloat(flt, 'G', -1, 64)
	return str
}

func uintToStr(ii uint64) string {
	str := strconv.FormatUint(ii, 10)
	return str
}

func intToStr(ii int) string {
	str := strconv.FormatInt(int64(ii), 10)
	return str
}

func strToInt(s string) int {
	ii, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
		panic("strToInt failed")
	}
	return int(ii)
}

func showLinksMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<div id="header">
<p><a href="add">Add</a>
<a href="list">List</a> &middot;
<font color="grey">`+htmize(userName)+`</font>
</p>
</div>
`)
}

func showExposition(w http.ResponseWriter, db mysql.Conn, targetUrl string, imageUrl string, description string, myComment string, title string, createTimeGMT uint64) {
	fmt.Fprint(w, `
      <p> (full post) Google+ / <a href="https://www.facebook.com/">Facebook</a> / <a href="http://www.linkedin.com/">LinkedIn</a> / <a href="https://www.tumblr.com/blog/waynerad/new/link">Tumblr</a> / <a href="https://mewe.com/myworld">MeWe</a> / <a href="https://www.minds.com/therealwaynerad">Minds</a> / <a href="https://vk.com/feed">VKontakte</a> / (Twitter-style) <a href="http://www.twitter.com/">Twitter</a> / <a href="https://gab.com/home">Gab</a> / <a href="https://mastodon.social/web/web">Mastodon</a> / (special case) <a href="https://pluspora.com/stream">Pluspora</a> </p>
      <p><textarea class="infield" cols="80" rows="20">`)
	fmt.Fprint(w, description)
	fmt.Fprint(w, `
`)
	fmt.Fprint(w, targetUrl)
	fmt.Fprint(w, `</textarea> <textarea class="infield" cols="80" rows="20">`)
	fmt.Fprint(w, myComment)
	fmt.Fprint(w, `</textarea> <textarea class="infield" cols="80" rows="4">`)
	fmt.Fprint(w, title)
	fmt.Fprint(w, ` `)
	fmt.Fprint(w, targetUrl)
	fmt.Fprint(w, `</textarea> </p>
<p> <img src="`+imageUrl+`" alt="Thumbnail" /> </p>
`)
	showTrentExposition(w, db, targetUrl, imageUrl, description, myComment, title, createTimeGMT)
	fmt.Fprint(w, `
  </section>
`)
}

func showAddPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var showform bool
	var errorList map[string]string
	var uiGrabbedUrl string
	var uiOriginalText string
	var uiMyComment string
	var uiImageUrl string
	var uiTitle string
	var uiBpm string

	errorOccurred := false
	method := r.Method
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("parseform failed")
		}
		postform := r.Form
		if (postform["grabbed_url"][0] == "") || (postform["original_text"][0] == "") || (postform["title"][0] == "") {
			showform = true
			errorOccurred = true
			errorList = make(map[string]string)
			if postform["grabbed_url"][0] == "" {
				errorList["grabbed_url"] = "Grabbed URL is empty"
			}
			if postform["original_text"][0] == "" {
				errorList["original_text"] = "Original text is empty"
			}
			if postform["title"][0] == "" {
				errorList["title"] = "Title is empty"
			}
		}
		if postform["bpm"][0] != "" {
			if !isNumeric(postform["bpm"][0]) {
				showform = true
				errorOccurred = true
				errorList["bpm"] = "BPM is not a number"
			}
		}
		if errorOccurred == false {
			db := accessdb.GetDbConnection()
			defer db.Close()
			stmt, err := db.Prepare("INSERT INTO link_link (created_gmt, target_url, image_url, description, my_comment, title, is_email, is_public, is_video, is_pdf, bpm) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// defer stmt.Close();
			saveCreatedGmt := uint64(time.Now().Unix())
			saveTargetUrl := postform["grabbed_url"][0]
			saveImageUrl := postform["image_url"][0]
			saveDescription := postform["original_text"][0]
			saveDescription = strings.Trim(saveDescription, " \r\n\t")
			saveMyComment := postform["my_comment"][0]
			saveMyComment = strings.Trim(saveMyComment, " \r\n\t")
			saveTitle := postform["title"][0]
			saveTitle = strings.Trim(saveTitle, " \r\n\t")
			saveIsEmail := 0
			saveIsPdf := 0
			_, ok := postform["email"]
			if ok {
				saveIsEmail = 1
			}
			saveIsPublic := saveIsEmail
			saveIsVideo := 0
			_, ok = postform["video"]
			if ok {
				saveIsVideo = 1
			}
			_, ok = postform["pdf"]
			if ok {
				saveIsPdf = 1
			}
			var saveBpm float64
			saveBpm = 0.0
			if postform["bpm"][0] != "" {
				saveBpm = strToFloat(postform["bpm"][0])
			}
			stmt.Bind(saveCreatedGmt, saveTargetUrl, saveImageUrl, saveDescription, saveMyComment, saveTitle, saveIsEmail, saveIsPublic, saveIsVideo, saveIsPdf, saveBpm)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			header := w.Header()
			header.Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, getDoctype()+getStyle())
			fmt.Fprint(w, `<title>URL saver</title>
</head>
<body>
  <section>
`)
			showLinksMenuBar(w, userName)
			fmt.Fprint(w, `
    <h1>URL Saver</h1>
	<a href="add">Next</a>
	<a href="list">List</a>
`)
			showExposition(w, db, saveTargetUrl, saveImageUrl, saveDescription, saveMyComment, saveTitle, saveCreatedGmt)
			fmt.Fprint(w, `
</section></body></html>`)
		}
		uiGrabbedUrl = postform["grabbed_url"][0]
		uiImageUrl = postform["image_url"][0]
		uiOriginalText = postform["original_text"][0]
		uiTitle = postform["title"][0]
		uiBpm = postform["bpm"][0]
	}
	if method == "GET" {
		showform = true
		uiGrabbedUrl = ""
		uiImageUrl = ""
		uiOriginalText = ""
		uiTitle = ""
		uiBpm = ""
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype()+getStyle())
		db := accessdb.GetDbConnection()
		defer db.Close()
		sql := "SELECT COUNT(*) FROM link_link WHERE is_email = 1;"
		res, err := db.Start(sql)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// defer res.Close();
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		emailCount := row.Uint64(0)
		// <link rel="stylesheet" type="text/css" href="/style.css">
		fmt.Fprint(w, `<title>URL saver</title>

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

function minposition(i, j) {
    "use strict";
    if (i === -1) {
        return j;
    }
    if (j === -1) {
        return i;
    }
    if (i < j) {
        return i;
    }
    return j;
}

function execUrlGrab() {
    "use strict";
    var txt, i, j, k, link, setvideo, setpdf;
    txt = document.getElementById("original_text").value;
    i = txt.indexOf("http://");
    if (i === -1) {
        i = txt.indexOf("https://");
        if (i === -1) {
            return;
        }
    }
    j = txt.indexOf(" ", i);
    k = txt.indexOf("\n", i);
    j = minposition(j, k);
    if (j === -1) {
        link = txt.substring(i);
        txt = txt.substring(0, i);
    } else {
        link = txt.substring(i, j);
        txt = txt.substring(0, i) + txt.substring(j);
    }
    document.getElementById("grabbed_url").value = link.trim();
    document.getElementById("original_text").value = txt.trim();
    window.clearInterval(timeridUrl);
    setvideo = false;
    i = link.indexOf("youtube");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("vimeo");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("thedailyshow");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("colbertnation");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("video.pbs");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("collegehumor");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("ted.com");
    if (i !== -1) {
        setvideo = true;
    }
    i = link.indexOf("kickstarter");
    if (i !== -1) {
        setvideo = true;
    }
    if (setvideo) {
        document.getElementById("video").checked = true;
    }
    i = link.indexOf("pdf");
    setpdf = false;
    if (i !== -1) {
        setpdf = true;
    }
    if (setpdf) {
        document.getElementById("pdf").checked = true;
    }
    timeridImage = window.setInterval(execImageGrab, 100);
}

function execImageGrab() {
    "use strict";
    var txt, i, j, k, link;
    txt = document.getElementById("original_text").value;
    i = txt.indexOf("http://");
    if (i === -1) {
        i = txt.indexOf("https://");
        if (i === -1) {
            return;
        }
    }
    window.clearInterval(timeridImage);
    j = txt.indexOf(" ", i);
    k = txt.indexOf("\n", i);
    j = minposition(j, k);
    if (j === -1) {
        link = txt.substring(i);
        txt = txt.substring(0, i);
    } else {
        link = txt.substring(i, j);
        txt = txt.substring(0, i) + txt.substring(j);
    }
    document.getElementById("image_url").value = link.trim();
    document.getElementById("original_text").value = txt.trim();
}

function execDoTheLcase() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    str = str.toLowerCase();
    document.getElementById("original_text").value = str;
}

function execDoTheResize() {
    "use strict";
    var cols;
    cols = document.getElementById("original_text").cols;
    if (cols === 80) {
        cols = 53;
    } else {
        if (cols === 53) {
            cols = 113;
        } else {
            if (cols === 113) {
                cols = 257;
            } else {
                cols = 80;
            }
        }
    }
    document.getElementById("original_text").cols = cols;
}

function strReplaceAll(str, from, to) {
    "use strict";
    var rv;
    rv = str.replace(from, to);
    while (rv !== str) {
        str = rv;
        rv = str.replace(from, to);
    }
    return rv;
}

function execDoFlipQuotes() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    str = strReplaceAll(str, "'", "gwikkagakkageek");
    str = strReplaceAll(str, '"', "'");
    str = strReplaceAll(str, "gwikkagakkageek", '"');
    document.getElementById("original_text").value = str;
}

function execDoUnderscores() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    str = strReplaceAll(str, " ", "_");
    document.getElementById("original_text").value = str;
}

function execDoTheAnalyze() {
    "use strict";
    var str, result, i;
    str = document.getElementById("original_text").value;
    result = "";
    for (i = 0; i < str.length; i = i + 1) {
        result = result + " " + str.charAt(i) + " " + str.charCodeAt(i);
    }
    document.getElementById("analyze_result").value = result;
}

function replaceQuotesWithHyphens(str) {
    "use strict";
    var result, i, a;
    result = "";
    for (i = 0; i < str.length; i = i + 1) {
        a = str.charCodeAt(i);
        switch (a) {
        case 8211:
        case 8212:
            result = result + " -- ";
            break;
        case 8230:
            result = result + "...";
            break;
        case 8216:
            result = result + "'";
            break;
        case 8217:
            result = result + "'";
            break;
        case 8220:
            result = result + '"';
            break;
        case 8221:
            result = result + '"';
            break;
        default:
            result = result + str.charAt(i);
            break;
        }
    }
    result = strReplaceAll(result, "  ", "gwikkagakkageek");
    result = strReplaceAll(result, "gwikkagakkageek", " ");
    return result;
}

function execDoTheQuotesHyphens() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    document.getElementById("original_text").value = replaceQuotesWithHyphens(str);
    str = document.getElementById("my_comment").value;
    document.getElementById("my_comment").value = replaceQuotesWithHyphens(str);
}

function execDoTheSingleQuotes() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    str = strReplaceAll(str, '"', "'");
    document.getElementById("original_text").value = str;
    str = document.getElementById("my_comment").value;
    str = strReplaceAll(str, '"', "'");
    document.getElementById("my_comment").value = str;

}

function execDoBothQuotesHyphensAndSingle() {
    "use strict";
    execDoTheQuotesHyphens();
    execDoTheSingleQuotes();
}

// main

var timeridUrl, timeridImage;
timeridUrl = window.setInterval(execUrlGrab, 100);

</script>

</head>
<body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>URL Saver</h1>

<p>Email count: `+htmize(strconv.FormatUint(emailCount, 10))+`

<form action="add" method="post">
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for errOn, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmize(errOn)+": "+htmize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		if emailCount < 80 {
			fmt.Fprint(w, `

<p>
<input class="infield" name="grab_url" id="grab_url" type="button" value="Grab URL" />
URL: <input class="infield" class="biginput" name="grabbed_url" id="grabbed_url" type="text" value="`+htmize(uiGrabbedUrl)+`" />
</p>
<p>
<input class="infield" name="do_resize" id="do_resize" type="button" value="resize" />
<input class="infield" name="submit" id="submit" type="submit" />
<input class="infield" type="checkbox" checked="checked" id="email" name="email"> Email, 
<input class="infield" type="checkbox" id="video" name="video"> Video
<input class="infield" type="checkbox" id="pdf" name="pdf"> PDF
</p>
<p>
<input class="infield" name="grab_image" id="grab_image" type="button" value="Grab URL" />
Image URL: <input class="infield" class="biginput" name="image_url" id="image_url" type="text" value="`+htmize(uiImageUrl)+`" /> </p>
<p> Title: <input class="infield" class="biginput" name="title" id="title" type="text" value="`+htmize(uiTitle)+`" /> </p>
<p><textarea class="bigtextarea" name="original_text" id="original_text" cols="80" rows="20">`+htmize(uiOriginalText)+`</textarea></p>
<p><input class="infield" name="do_lcase" id="do_lcase" type="button" value="lcase" />
 <input class="infield" name="do_underscores" id="do_underscores" type="button" value="underscores" />
 <input class="infield" name="do_flip_quotes" id="do_flip_quotes" type="button" value="flip quotes" />
 <input class="infield" name="do_analyze" id="do_analyze" type="button" value="analyze ascii" />
 <input class="infield" name="do_single_quotes" id="do_singlequotes" type="button" value="single quotes" />
 <input class="infield" name="do_quotes_hyphens" id="do_quoteshyphens" type="button" value="quotes hyphens" />
 <input class="infield" name="do_both" id="do_both" type="button" value="both" />
</p>
<p><textarea class="bigtextarea" name="my_comment" id="my_comment" cols="80" rows="20">`+htmize(uiMyComment)+`</textarea></p>
<p>Analysis<br /><textarea class="bigtextarea" name="analyze_result" id="analyze_result" cols="80" rows="10"></textarea></p>
<p>
BPM: <input class="infield" class="biginput" name="bpm" id="bpm" type="text" value="`+htmize(uiBpm)+`" />
</p>

</form>

  </section>
<script>

var startElement = document.getElementById('grab_url');
startElement.addEventListener('click', execUrlGrab, true);

startElement = document.getElementById('grab_image');
startElement.addEventListener('click', execImageGrab, true);

startElement = document.getElementById('do_lcase');
startElement.addEventListener('click', execDoTheLcase, true);

startElement = document.getElementById('do_flip_quotes');
startElement.addEventListener('click', execDoFlipQuotes, true);

startElement = document.getElementById('do_underscores');
startElement.addEventListener('click', execDoUnderscores, true);

startElement = document.getElementById('do_resize');
startElement.addEventListener('click', execDoTheResize, true);

startElement = document.getElementById('do_analyze');
startElement.addEventListener('click', execDoTheAnalyze, true);

startElement = document.getElementById('do_singlequotes');
startElement.addEventListener('click', execDoTheSingleQuotes, true);

startElement = document.getElementById('do_quoteshyphens');
startElement.addEventListener('click', execDoTheQuotesHyphens, true);

startElement = document.getElementById('do_both');
startElement.addEventListener('click', execDoBothQuotesHyphensAndSingle, true);

</script>
`)
		}
		fmt.Fprint(w, `
</body>
</html>`)
		return
	}

}

func showListPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>List of URLs</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>List of URLs</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	_, showAll := getform["showall"]
	if showAll {
		sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE 1 ORDER BY id_lnk DESC;"
	} else {
		_, showEmail := getform["emailonly"]
		if showEmail {
			// emailCutoff
			sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE (is_email = 1) AND (is_pdf = 0) ORDER BY id_lnk DESC;"
			_, exists := getform["cutoff"]
			if exists {
				cutstr := getform["cutoff"][0]
				cutoff, err := strconv.ParseUint(cutstr, 10, 64)
				if err != nil {
					fmt.Fprintln(w, err)
					return
				}
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE (is_email = 1) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;"
			}
		} else {
			_, videoOnly := getform["videoonly"]
			if videoOnly {
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE (is_video = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
			} else {
				_, pdf_only := getform["pdfonly"]
				if pdf_only {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE (is_pdf = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
				} else {

					_, doSearch := getform["search"]
					if doSearch {
						sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE (description LIKE '%3D print%') OR (description LIKE '%3D-print%') ORDER BY id_lnk;"
					} else {
						sql = "SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE 1 ORDER BY id_lnk DESC LIMIT 0, 50;"
					}
				}
			}
		}
	}
	res, err := db.Start(sql)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			linkid := row.Uint64(0)
			// createdGmt = row.Uint64(1)
			targetUrl := row.Str(2)
			// imageUrl = row.Str(3)
			description := row.Str(4)
			myComment := row.Str(5)
			fmt.Fprint(w, `<p>`)
			fmt.Fprint(w, htmize(description))
			fmt.Fprint(w, `<br />`)
			if myComment != "" {
				fmt.Fprint(w, htmize(myComment))
				fmt.Fprint(w, `<br />`)
			}
			fmt.Fprint(w, "<a href="+targetUrl+">"+htmize(targetUrl)+"</a>")
			fmt.Fprint(w, ` &middot; <a href="exposit?link=`+strconv.FormatUint(linkid, 10)+`">exposit</a> &middot; <a href="edit?link=`+strconv.FormatUint(linkid, 10)+`">edit</a> &middot; <a href="delete?link=`+strconv.FormatUint(linkid, 10)+`">delete</a>
</p>`)
		}
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func showExpositPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	postform := r.Form
	linkstr := postform["link"][0]
	linkid, err := strconv.ParseUint(linkstr, 10, 64)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>URL exposition</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Exposition</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment, title FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			// linkid = row.Uint64(0)
			createdGmt := row.Uint64(1)
			targetUrl := row.Str(2)
			imageUrl := row.Str(3)
			description := row.Str(4)
			myComment := row.Str(5)
			title := row.Str(6)
			showExposition(w, db, targetUrl, imageUrl, description, myComment, title, createdGmt)
		}
	}
}

func checkedStr(isSet bool) string {
	if isSet {
		return `checked="checked"`
	}
	return ""
}

func showEditPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var errorOccurred bool
	var errorList map[string]string
	var linkid uint64
	var linkstr string
	var uiIsEmail bool
	var uiIsPublic bool
	var uiIsVideo bool
	var uiIsPdf bool
	var uiCreated string
	var uiTargetUrl string
	var uiImageUrl string
	var uiDescription string
	var uiMyComment string
	var uiTitle string
	var uiBpm string
	method := r.Method
	errorList = make(map[string]string)
	showform := true
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		postform := r.Form
		_, ok := postform["link"]
		if !ok {
			fmt.Fprintln(w, "link id is missing or invalid.")
			return
		}
		linkstr = postform["link"][0]
		linkid, err = strconv.ParseUint(linkstr, 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		db := accessdb.GetDbConnection()
		res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment, title, is_email, is_public, is_video, is_pdf, bpm FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// defer res.Close();
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			fmt.Fprintln(w, "What? row == nil. Where is it?")
			return
		} else {
			// linkid = row.Uint64(0)
			createdGmt := row.Uint64(1)
			targetUrl := row.Str(2)
			imageUrl := row.Str(3)
			description := row.Str(4)
			myComment := row.Str(5)
			title := row.Str(6)
			timeObject := time.Unix(int64(createdGmt), 0)
			// timeObject.Format("Mon Jan 2 15:04:05 MST 2006  (MST is GMT-0700)
			createstr := timeObject.Format("2006-01-02 15:04:05")
			uiCreated = createstr
			uiTargetUrl = targetUrl
			uiImageUrl = imageUrl
			uiDescription = description
			uiMyComment = myComment
			uiTitle = title
			uiIsEmail = row.Bool(7)
			uiIsPublic = row.Bool(8)
			uiIsVideo = row.Bool(9)
			uiIsPdf = row.Bool(10)
			uiBpm = floatToStr(row.Float(11))
		}
	}
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		postform := r.Form
		uiCreated = postform["created"][0]
		uiTargetUrl = postform["target_url"][0]
		uiImageUrl = postform["image_url"][0]
		uiDescription = postform["description"][0]
		uiMyComment = postform["my_comment"][0]
		uiTitle = postform["title"][0]
		_, uiIsEmail = postform["email"]
		_, uiIsPublic = postform["public"]
		_, uiIsVideo = postform["video"]
		_, uiIsPdf = postform["pdf"]
		uiBpm = postform["bpm"][0]
		if (postform["target_url"][0] == "") || (postform["description"][0] == "") || (postform["title"][0] == "") {
			showform = true
			errorOccurred = true
			if postform["target_url"][0] == "" {
				errorList["target_url"] = "Target URL is empty"
			}
			if postform["description"][0] == "" {
				errorList["description"] = "Description is empty"
			}
			if postform["title"][0] == "" {
				errorList["title"] = "Title is empty"
			}
		}
		if postform["bpm"][0] != "" {
			if postform["bpm"][0] != "" {
				if !isNumeric(postform["bpm"][0]) {
					errorList["bpm"] = "BPM is not a number"
					showform = true
					errorOccurred = true
				}
			}
		}
		linkstr = postform["link"][0]
		linkid, err = strconv.ParseUint(linkstr, 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if errorOccurred == false {
			db := accessdb.GetDbConnection()
			defer db.Close()
			stmt, err := db.Prepare("UPDATE link_link SET target_url = ?, image_url = ?, description = ?, my_comment = ?, title = ?, is_email = ?, is_public = ?, is_video = ?, is_pdf = ?, bpm = ? WHERE id_lnk = ?;")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// defer stmt.Close();
			// updateCreatedGmt = uint64(time.Now().Unix())
			updateTargetUrl := postform["target_url"][0]
			updateImageUrl := postform["image_url"][0]
			updateDescription := postform["description"][0]
			updateDescription = strings.Trim(updateDescription, " \r\n\t")
			updateMyComment := postform["my_comment"][0]
			updateMyComment = strings.Trim(updateMyComment, " \r\n\t")
			updateTitle := postform["title"][0]
			updateTitle = strings.Trim(updateTitle, " \r\n\t")
			updateIdLnk := linkid
			updateIsEmail := 0
			_, ok := postform["email"]
			if ok {
				updateIsEmail = 1
			}
			_, ok = postform["public"]
			updateIsPublic := 0
			if ok {
				updateIsPublic = 1
			}
			_, ok = postform["video"]
			updateIsVideo := 0
			if ok {
				updateIsVideo = 1
			}
			_, ok = postform["pdf"]
			updateIsPdf := 0
			if ok {
				updateIsPdf = 1
			}
			var updateBpm float64
			updateBpm = 0.0
			if postform["bpm"][0] != "" {
				updateBpm = strToFloat(postform["bpm"][0])
			}
			stmt.Bind(updateTargetUrl, updateImageUrl, updateDescription, updateMyComment, updateTitle, updateIsEmail, updateIsPublic, updateIsVideo, updateIsPdf, updateBpm, updateIdLnk)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			http.Redirect(w, r, "exposit?link="+linkstr, 302)
			return
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for errOn, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmize(errOn)+": "+htmize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, getDoctype()+getStyle())
		fmt.Fprint(w, `<title>URL saver</title>
</head><body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Edit URL</h1>

<form action="edit" method="post">

<table border="0" cellpadding="4">
<tr><td> Created </td><td> <input class="infield" name="link" id="link" value="`+strconv.FormatUint(linkid, 10)+`" type="hidden" /> <input class="infield" name="created" id="created" type="text" value="`+htmize(uiCreated)+`" readonly="readonly" /> </td></tr>
<tr><td> Target URL </td><td> <input class="infield" name="target_url" id="target_url" type="text" value="`+htmize(uiTargetUrl)+`" style="width:400px;" /> </td></tr>
<tr><td> Image URL </td><td> <input class="infield" name="image_url" id="image_url" type="text" value="`+htmize(uiImageUrl)+`" style="width:400px;" /> </td></tr>
<tr><td> Description </td><td> <textarea class="infield" name="description" id="description" rows="20" cols="80">`+htmize(uiDescription)+`</textarea> </td></tr>
<tr><td> My Comment </td><td> <textarea class="infield" name="my_comment" id="my_comment" rows="20" cols="80">`+htmize(uiMyComment)+`</textarea> </td></tr>
<tr><td> Title </td><td> <input class="infield" name="title" id="title" type="text" value="`+htmize(uiTitle)+`" style="width:400px;" /> </td></tr>
<tr><td> Email </td><td> <input class="infield" type="checkbox" name="email" id="email" `+checkedStr(uiIsEmail)+`> Email </td></tr>
<tr><td> Public </td><td> <input class="infield" type="checkbox" name="public" id="public" `+checkedStr(uiIsPublic)+`> Public </td></tr>
<tr><td> Video </td><td> <input class="infield" type="checkbox" name="video" id="video" `+checkedStr(uiIsVideo)+`> Video </td></tr>
<tr><td> PDF </td><td> <input class="infield" type="checkbox" name="pdf" id="video" `+checkedStr(uiIsPdf)+`> PDF </td></tr>
<tr><td> BPM </td><td> <input class="infield" name="bpm" id="bpm" type="text" value="`+htmize(uiBpm)+`" style="width:400px;" /> </td></tr>
</table>

<p><input class="infield" name="submit" id="submit" type="submit" />

</form>

</body></html>`)
	}
	return
}

func showDeletePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var linkid uint64
	var uiTargetUrl string
	var uiDescription string
	var uiMyComment string
	method := r.Method
	showform := true
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		postform := r.Form
		linkstr := postform["link"][0]
		linkid, err = strconv.ParseUint(linkstr, 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		db := accessdb.GetDbConnection()
		res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// defer res.Close();
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			fmt.Fprintln(w, "What? row == nil. Where is it?")
			return
		} else {
			// linkid = row.Uint64(0)
			// createdGmt = row.Uint64(1)
			targetUrl := row.Str(2)
			// imageUrl = row.Str(3)
			description := row.Str(4)
			myComment := row.Str(5)
			// timeObject = time.Unix(int64(createdGmt), 0)
			// timeObject.Format("Mon Jan 2 15:04:05 MST 2006  (MST is GMT-0700)
			// createstr = timeObject.Format("2006-01-02 15:04:05")
			// uiCreated = createstr
			uiTargetUrl = targetUrl
			// uiImageUrl = imageUrl
			uiDescription = description
			uiMyComment = myComment
		}
	}
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		postform := r.Form
		db := accessdb.GetDbConnection()
		defer db.Close()
		stmt, err := db.Prepare("DELETE FROM link_link WHERE id_lnk=?;")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// defer stmt.Close();
		linkstr := postform["link"][0]
		linkid, err = strconv.ParseUint(linkstr, 10, 64)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		stmt.Bind(linkid)
		_, _, err = stmt.Exec()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		http.Redirect(w, r, "list", 302)
		return
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype()+getStyle())
		fmt.Fprint(w, `<title>URL saver</title>
</head><body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Delete URL</h1>

<form action="delete" method="post">

<p> Delete link to: <a href="`+htmize(uiTargetUrl)+`">`+htmize(uiTargetUrl)+`</a>? </p>

<p> `+htmize(uiDescription)+`</p>
<p> `+htmize(uiMyComment)+`</p>

<p><input class="infield" type="submit" id="submit" name="submit" value="Delete" />
<input class="infield" name="link" id="link" value="`+strconv.FormatUint(linkid, 10)+`" type="hidden" />
</p>

</form>

</body></html>`)
	}
	return
}

func showEmailPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var cutoff uint64
	doubleLinefeeds := true
	method := r.Method
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		getform := r.Form
		_, nodouble := getform["nodouble"]
		doubleLinefeeds = !nodouble
		_, exists := getform["cutoff"]
		if exists {
			cutstr := getform["cutoff"][0]
			cutoff, err = strconv.ParseUint(cutstr, 10, 64)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
		} else {
			cutoff = 1610612736
		}
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>List of URLs</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>For Emails</h1>
<textarea class="infield" rows="100" cols="80">
=== News bits ===

(In reverse chronological order, except videos which are in their own section at the bottom)


`)
	db := accessdb.GetDbConnection()
	defer db.Close()

	count := 0

	// emailCutoff
	res, err := db.Start("SELECT target_url, description, my_comment FROM link_link WHERE (is_email = 1) AND (is_video = 0) AND (is_pdf = 0) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			targetUrl := row.Str(0)
			description := row.Str(1)
			myComment := row.Str(2)
			fmt.Fprint(w, "\n")
			if doubleLinefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, htmize(strings.Replace(strings.Replace(description, "\r", "", -1), "\n", " ", -1)))
			fmt.Fprint(w, "\n")
			if doubleLinefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, htmize(targetUrl)+"\n")
			if myComment != "" {
				fmt.Fprint(w, "\n")
				fmt.Fprint(w, htmize(strings.Replace(strings.Replace(myComment, "\r", "", -1), "\n", " ", -1)))
				fmt.Fprint(w, "\n")
			}
			count++
		}
	}
	fmt.Fprint(w, `



=== Video ===


`)
	// emailCutoff
	res, err = db.Start("SELECT target_url, description, my_comment FROM link_link WHERE (is_email = 1) AND (is_video = 1) AND (is_pdf = 0) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			targetUrl := row.Str(0)
			description := row.Str(1)
			myComment := row.Str(2)
			fmt.Fprint(w, "\n")
			if doubleLinefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, htmize(strings.Replace(strings.Replace(description, "\r", "", -1), "\n", " ", -1)))
			fmt.Fprint(w, "\n")
			if doubleLinefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, htmize(targetUrl)+"\n")
			if myComment != "" {
				fmt.Fprint(w, "\n")
				fmt.Fprint(w, htmize(strings.Replace(strings.Replace(myComment, "\r", "", -1), "\n", " ", -1)))
				fmt.Fprint(w, "\n")
			}
			count++
		}
	}
	fmt.Fprint(w, "</textarea>")
	fmt.Fprint(w, "<p>Count: "+htmize(strconv.FormatInt(int64(count), 10))+`</p>
</body></html>`)
}

func urlToDomainOnly(url string) string {
	rv := "View"
	i := strings.Index(url, "//")
	if i >= 0 {
		rv = url[i+2:]
		j := strings.Index(rv, "/")
		if j < 0 {
			rv = "View"
		} else {
			rv = rv[:j]
		}
	}
	if rv[0:4] == "www." {
		rv = rv[4:]
	}
	return rv
}

func showHomePage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string, showMusic bool) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	title := "Wayne Brain TV"
	if showMusic {
		title = "Music For Today"
	}
	fmt.Fprint(w, `<title>`+title+`</title>
<style>
body {
        font-family: "Trebuchet MS", "Helvetica", "Arial",  "Verdana", "sans-serif";
}
</style>
</head>
<body>
  <section>
    <h1>`+title+`</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	getform := r.Form
	_, showAll := getform["showall"]
	if showAll {
		sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC;"
	} else {
		_, showEmail := getform["emailonly"]
		if showEmail {
			// emailCutoff
			sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_email = 1) AND (is_pdf = 0) ORDER BY id_lnk DESC;"
			_, exists := getform["cutoff"]
			if exists {
				cutstr := getform["cutoff"][0]
				cutoff, err := strconv.ParseUint(cutstr, 10, 64)
				if err != nil {
					fmt.Fprintln(w, err)
					return
				}
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_email = 1) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;"
			}
		} else {
			_, videoOnly := getform["videoonly"]
			if videoOnly {
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
			} else {
				_, pdf_only := getform["pdfonly"]
				if pdf_only {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_pdf = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
				} else {

					_, doSearch := getform["search"]
					if doSearch {
						sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (description LIKE '%3D print%') OR (description LIKE '%3D-print%') ORDER BY id_lnk;"
					} else {
						_, music := getform["music"]
						if showMusic {
							music = true
						}
						if music {
							sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) AND (description LIKE 'Music for today%') ORDER BY id_lnk DESC;" //  LIMIT 0, 100;"
						} else {
							_, private := getform["private"]
							if private {
								sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) AND (is_public = 0) AND (description NOT LIKE 'Music for today%') ORDER BY id_lnk DESC LIMIT 0, 100;"
							} else {
								sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) AND (is_public = 1) AND (description NOT LIKE 'Music for today%') ORDER BY id_lnk DESC LIMIT 0, 100;"
							}
						}
					}
				}
			}
		}
	}
	first := true
	res, err := db.Start(sql)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			linkid := row.Uint64(0)
			// createdGmt = row.Uint64(1)
			targetUrl := row.Str(2)
			imageUrl := row.Str(3)
			description := row.Str(4)
			if first {
				fmt.Fprint(w, `<table border="0" cellpadding="4">`)
				first = false
			}
			fmt.Fprint(w, `<tr><td valign="top">`)
			if imageUrl != "" {
				fmt.Fprint(w, `<a href="`+targetUrl+`"><img width="196" src="`+imageUrl+`" alt="Thumbnail" /></a>`)
			}
			fmt.Fprint(w, ` </td><td valign="top"> `)
			fmt.Fprint(w, strings.Replace(htmize(description), "\n", "<br />", 32768))
			if userid == 1 {
				fmt.Fprint(w, `<br />`)
				fmt.Fprint(w, "<a href="+targetUrl+">"+htmize(targetUrl)+"</a>")
				fmt.Fprint(w, ` &middot; <a href="exposit?link=`+strconv.FormatUint(linkid, 10)+`">exposit</a> &middot; <a href="edit?link=`+strconv.FormatUint(linkid, 10)+`">edit</a> &middot; <a href="delete?link=`+strconv.FormatUint(linkid, 10)+`">delete</a>`)
				// } else {
				// fmt.Fprint(w, " <a href="+targetUrl+">"+urlToDomainOnly(targetUrl)+"</a>")
			}
			fmt.Fprint(w, " </td></tr>")
		}
	}
	if !first {
		fmt.Fprint(w, "</table>")
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func showSearchPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var showform bool
	var errorList map[string]string
	var uiSearchTerm string

	errorOccurred := false
	method := r.Method
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("parseform failed")
		}
		postform := r.Form
		if postform["search_term"][0] == "" {
			showform = true
			errorOccurred = true
			errorList = make(map[string]string)
			errorList["search_term"] = "Search term is empty"
		}
		if errorOccurred == false {
			// put search code here
			uiSearchTerm = postform["search_term"][0]
			header := w.Header()
			header.Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, getDoctype()+getStyle())
			fmt.Fprint(w, `<title>Search Results</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
`)
			showLinksMenuBar(w, userName)
			fmt.Fprint(w, `
    <h1>Search Results</h1>
`)
			db := accessdb.GetDbConnection()
			defer db.Close()
			sql := "SELECT id_lnk, description FROM link_link WHERE description LIKE ? ORDER BY id_lnk DESC LIMIT 0, 100;"
			sel, err := db.Prepare(sql)
			if err != nil {
				fmt.Println(err)
				panic("Prepare failed")
			}
			sel.Bind("%" + uiSearchTerm + "%")
			rows, _, err := sel.Exec()
			if err != nil {
				fmt.Println(err)
				panic("Bind/Exec failed")
			}
			fmt.Fprint(w, `<table border="0" cellpadding="4">
`)
			for _, row := range rows {
				linkId := row.Uint64(0)
				description := row.Str(1)
				fmt.Fprint(w, `<tr><td valign="top"> <a href="exposit?link=`+uintToStr(linkId)+`">`+uintToStr(linkId)+`</a> </td><td valign="top"> `+htmize(description)+`</td></tr>
`)
			}
			fmt.Fprint(w, `</table>
</body>
</html>`)
		}
	}
	if method == "GET" {
		showform = true
		uiSearchTerm = ""
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype()+getStyle())
		fmt.Fprint(w, `<title>Search</title>
<link rel="stylesheet" type="text/css" href="/style.css">
</head>
<body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Search</h1>

<form action="search" method="post">
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for errOn, errMsg := range errorList {
				fmt.Fprintln(w, "<li>"+htmize(errOn)+": "+htmize(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprintln(w, `

<table border="0" cellpadding="4">
<tr><td> Search Term </td><td> <input class="infield" name="search_term" id="search_term" type="text" value="`+htmize(uiSearchTerm)+`" /> </td></tr>
<tr><td colspan="2" align="center"> <input class="infield" name="submit" id="submit" type="submit" /> </td></tr>
</table>
</form>
</body>
</html>`)
		return
	}
}

func linesToParagraphBreaks(str string) string {
	return strings.Replace(str, "\r\n\r\n", "</p><p>", -1)
}

func linkify(text string) string {
	foundstuff := false
	i := strings.Index(text, "http:")
	for i >= 0 {
		foundstuff = true
		j := strings.IndexAny(text[i:], " \t\n\r") + i
		var link string
		var repl string
		if j >= i {
			link = text[i:j]
			repl = "gwikkagakkageek:" + text[i+5:j]
		} else {
			link = text[i:]
			repl = "gwikkagakkageek:" + text[i+5:]
		}
		text = strings.Replace(text, link, `<a href="`+repl+`">`+repl+"</a>", -1)
		i = strings.Index(text, "http:")
	}
	i = strings.Index(text, "https:")
	for i >= 0 {
		foundstuff = true
		j := strings.IndexAny(text[i:], " \t\n\r") + i
		var link string
		var repl string
		if j >= i {
			link = text[i:j]
			repl = "gwikkagakkageeks:" + text[i+6:j]
		} else {
			link = text[i:]
			repl = "gwikkagakkageeks:" + text[i+6:]
		}
		text = strings.Replace(text, link, `<a href="`+repl+`">`+repl+"</a>", -1)
		i = strings.Index(text, "https:")
	}
	if !foundstuff {
		return text
	}
	text = strings.Replace(text, "gwikkagakkageeks:", "https:", -1)
	text = strings.Replace(text, "gwikkagakkageek:", "http:", -1)
	return text
}

func genTrentBlogVersion(targetUrl string, imageUrl string, description string, myComment string) string {
	result := `<p><a href="` + targetUrl + `"><img border="0" src="` + imageUrl + `" width="200" style="float:left; margin-right:10px; margin-bottom-10px;" /></a>` + linesToParagraphBreaks(linkify(description)) + "</p>" + `
`
	if myComment != "" {
		result = result + "<p>My comment: " + linesToParagraphBreaks(linkify(myComment)) + "</p>" + `
`
	}
	result = result + "<p><a href=" + `"` + targetUrl + `"` + ">Read more...</a></p>"
	return result
}

func showTrentExposition(w http.ResponseWriter, db mysql.Conn, targetUrl string, imageUrl string, description string, myComment string, title string, createTimeGMT uint64) {
	fmt.Fprint(w, `
      <p><textarea class="infield" cols="80" rows="20">`)
	trentVersion := genTrentBlogVersion(targetUrl, imageUrl, description, myComment)
	fmt.Fprint(w, htmize(trentVersion))
	fmt.Fprint(w, `</textarea></p>
<p>`)
	fmt.Fprint(w, trentVersion)
	var record lnksqlfields
	record.linkId = 0
	record.createdGMT = createTimeGMT
	record.targetURL = targetUrl
	record.imageURL = imageUrl
	record.description = description
	record.myComment = myComment
	record.title = title
	sql := genSQLFromRecord(db, 9999998, record)
	fmt.Fprint(w, `</p><p><textarea class="infield" cols="80" rows="20">`)
	fmt.Fprint(w, htmize(sql))
	fmt.Fprint(w, `</textarea></p>
`)
	return
	// candelete after this
	fmt.Fprint(w, htmize(`<p><a href="`+targetUrl+`"><img border="0" src="`+imageUrl+`" width="200" style="float:left; margin-right:10px; margin-bottom-10px;" /></a>`))
	fmt.Fprint(w, linesToParagraphBreaks(description))
	fmt.Fprint(w, htmize("</p>"))
	fmt.Fprint(w, `
`)
	fmt.Fprint(w, htmize("<p><a href="+`"`))
	fmt.Fprint(w, targetUrl)
	fmt.Fprint(w, htmize(`"`+">Read more...</a></p>"))
	fmt.Fprint(w, `</textarea> <textarea class="infield" cols="80" rows="20">`)
	fmt.Fprint(w, "<p>"+linesToParagraphBreaks(myComment)+"</p>")
	fmt.Fprint(w, `</textarea></p>
<p> <img src="`+imageUrl+`" alt="Thumbnail" /> </p>
  </section>
`)
}

func showTrentExpoPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	postform := r.Form
	linkstr := postform["link"][0]
	linkid, err := strconv.ParseUint(linkstr, 10, 64)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>Trent Exposition</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Trent Exposition</h1>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment, title FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			// linkid = row.Uint64(0)
			createdGmt := row.Uint64(1)
			targetUrl := row.Str(2)
			imageUrl := row.Str(3)
			description := row.Str(4)
			myComment := row.Str(5)
			title := row.Str(6)
			// description = db.Escape(description)
			// myComment = db.Escape(myComment)
			showTrentExposition(w, db, targetUrl, imageUrl, description, myComment, title, createdGmt)
		}
	}
}

func descripShorten(description string) string {
	punct := ".,!?"
	for num := range punct {
		pmk := punct[num : num+1]
		i := strings.Index(description, pmk)
		if i >= 0 {
			i = strings.LastIndex(description, pmk)
			if pmk == "?" {
				// we keep question marks
				description = description[:i+1]
				description = strings.Trim(description, " \n\r\t")
				if len(description) < 65 { // this if prevents endless loop
					return description
				}
			} else {
				description = description[:i]
				description = strings.Trim(description, " \n\r\t")
				return description
			}
		}
	}
	description = description[:65]
	for (len(description) > 10) && (description[len(description)-1:len(description)] != " ") {
		description = description[:len(description)-1]
	}
	description = strings.Trim(description, " \n\r\t")
	return description
}

func genTitleFromDescription(description string) string {
	description = strings.Replace(description, `"`, "", -1)
	description = strings.Replace(description, "\n", " ", -1)
	description = strings.Replace(description, "\r", " ", -1)
	description = strings.Replace(description, "\t", " ", -1)
	description = strings.Replace(description, "  ", " ", -1)
	description = strings.Replace(description, "  ", " ", -1)
	description = strings.Replace(description, "  ", " ", -1)
	if len(description) > 17 {
		if description[:17] == "Music for today. " {
			description = description[17:]
		}
	}
	if len(description) > 18 {
		if description[:18] == "Sign O The Times. " {
			description = description[18:]
		}
	}
	// commonWords := [84]string{"the", "be", "to", "of", "and", "a", "in", "that", "have", "I", "it", "for", "on", "with", "he", "as", "you", "do", "at", "this", "but", "his", "by", "from", "they", "we", "say", "her", "she", "or", "will", "an", "my", "one", "all", "would", "there", "their", "what", "so", "up", "out", "if", "about", "who", "get", "which", "go", "when", "me", "can", "like", "no", "just", "him", "know", "take", "into", "your", "some", "could", "them", "see", "other", "than", "then", "now", "look", "only", "come", "its", "over", "also", "back", "use", "two", "how", "our", "even", "new", "any", "these", "give", "us"}
	// commonWords := [13]string{"the", "a", "an", "be", "of", "and", "that", "it", "this", "would", "could", "there", "so"}
	commonWords := [13]string{"the", "a", "an"}
	for num := range commonWords {
		word := commonWords[num]
		word = " " + word + " "
		i := strings.Index(description, word)
		for i >= 0 {
			description = strings.Replace(description, word, " ", -1)
			i = strings.Index(description, word)
		}
	}
	type srPair struct {
		searchFor string
		replaceBy string
	}
	searchAndReplace := []srPair{{"United States", "US"}}
	for num := range searchAndReplace {
		srch := searchAndReplace[num].searchFor
		repl := searchAndReplace[num].replaceBy
		i := strings.Index(description, srch)
		for i >= 0 {
			description = strings.Replace(description, srch, repl, -1)
			i = strings.Index(description, srch)
		}
	}
	// have to check for space after period because of ".js", ".gov", etc
	i := strings.Index(description, ". ")
	if i >= 0 {
		description = description[:i]
	}
	i = strings.Index(description, "!")
	if i >= 0 {
		description = description[:i]
	}
	i = strings.Index(description, "?")
	if i >= 0 {
		description = description[:i+1]
	}
	for len(description) > 65 {
		description = descripShorten(description)
	}
	description = strings.Trim(description, " \n\r\t.") // eliminate trailing periods left by period+space search above
	return description
}

func showGenTitlesPage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>Generate Titles</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Generate Titles</h1>
`)
	type updateLinkTitleRec struct {
		linkId uint64
		title  string
	}
	var updateList [8192]updateLinkTitleRec
	count := 0
	db := accessdb.GetDbConnection()
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, target_url, description FROM link_link WHERE title = '' ORDER BY id_lnk LIMIT 0, 8192;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			linkid := row.Uint64(0)
			// targetUrl := row.Str(1)
			description := row.Str(2)
			fmt.Fprint(w, `<p> Description "`+htmize(description)+`"</p>`)
			newTitle := genTitleFromDescription(description)
			fmt.Fprint(w, `<p> Title "`+htmize(newTitle)+`"</p>`)
			updateList[count].linkId = linkid
			updateList[count].title = newTitle
			count++
		}
	}
	for i := 0; i < count; i++ {
		updateLnkId := updateList[i].linkId
		updateTitle := updateList[i].title
		stmt, err := db.Prepare("UPDATE link_link SET title = ? WHERE id_lnk = ?;")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		// defer stmt.Close();
		stmt.Bind(updateTitle, updateLnkId)
		_, _, err = stmt.Exec()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
	}
}

func genSQLFromRecord(db mysql.Conn, idNum int, record lnksqlfields) string {
	timeZoneConvert := -25200
	url := strings.Replace(record.title, " ", "-", -1)
	postTime := time.Unix(int64(record.createdGMT)+int64(timeZoneConvert), 0)
	dateString := postTime.Format("2006-01-02 15:04:05")
	postContent := genTrentBlogVersion(record.targetURL, record.imageURL, record.description, record.myComment)
	outRec := make(map[string]string)
	outRec["ID"] = intToStr(idNum)
	outRec["post_date"] = "'" + db.Escape(dateString) + "'"
	outRec["post_date_gmt"] = "'" + db.Escape(dateString) + "'"
	outRec["description"] = "'" + db.Escape("") + "'"
	outRec["post_content"] = "'" + db.Escape(postContent) + "'"
	outRec["post_title"] = "'" + db.Escape(record.title) + "'"
	outRec["post_name"] = "'" + db.Escape(url) + "'"
	outRec["author_name"] = "'" + db.Escape("Wayne Radinsky") + "'"
	outRec["author_id"] = "1"
	outRec["post_status"] = "1"
	outRec["views"] = "1"
	outRec["header_image"] = "''"
	outRec["categoryid"] = "3"
	fields := ""
	values := ""
	for k := range outRec {
		fields = fields + ", " + k
		values = values + ", " + outRec[k]
	}
	fields = fields[2:]
	values = values[2:]
	sql := "INSERT INTO blog_posts (" + fields + ") VALUES (" + values + ");"
	return sql
}

func showGenTrentSQL(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	idNum := 99999998
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>Generate SQL</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Generate SQL</h1>
`)
	localfile := "/home/ec2-user/trentblog.sql"
	outFile, err := os.Create(localfile)
	if err != nil {
		fmt.Println("os.Create failed")
		fmt.Println(err)
		panic("os.Create failed")
	}
	defer outFile.Close()
	var record lnksqlfields
	db := accessdb.GetDbConnection()
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment, title FROM link_link WHERE (description NOT LIKE 'Music for today%') AND (description NOT LIKE 'Sign O The Times%') AND (is_pdf = 0) ORDER BY id_lnk;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			record.linkId = row.Uint64(0)
			record.createdGMT = row.Uint64(1)
			record.targetURL = row.Str(2)
			record.imageURL = row.Str(3)
			record.description = row.Str(4)
			record.myComment = row.Str(5)
			record.title = row.Str(6)
			sql := genSQLFromRecord(db, idNum, record)
			_, err = outFile.WriteString(sql + "\n\n")
			if err != nil {
				fmt.Println("outFile.WriteString failed")
				fmt.Println(err)
				panic("os.WriteString failed")
			}
		}
	}
	fmt.Fprint(w, `
<p>Done</p>
</section></body></html>`)
}

func showGenTrentCatchupSQL(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("parseform failed")
	}
	postform := r.Form
	fromNum := strToInt(postform["from"][0])
	idNum := strToInt(postform["target"][0])
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>Generate Catchup SQL</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Generate Catchup SQL</h1>
<textarea class="infield" rows="80" cols="80">`)
	var record lnksqlfields
	db := accessdb.GetDbConnection()
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, my_comment, title FROM link_link WHERE (id_lnk > " + intToStr(fromNum) + ") AND (description NOT LIKE 'Music for today%') AND (description NOT LIKE 'Sign O The Times%') AND (is_pdf = 0) ORDER BY id_lnk;")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if row == nil {
			break
		} else {
			record.linkId = row.Uint64(0)
			record.createdGMT = row.Uint64(1)
			record.targetURL = row.Str(2)
			record.imageURL = row.Str(3)
			record.description = row.Str(4)
			record.myComment = row.Str(5)
			record.title = row.Str(6)
			idNum++
			sql := genSQLFromRecord(db, idNum, record)
			fmt.Fprint(w, htmize(sql))
			fmt.Fprint(w, "\n\n")
		}
	}
	fmt.Fprint(w, `</textarea>
<p>Done</p>
</section></body></html>`)
}

func Handler(w http.ResponseWriter, r *http.Request, host string, op string, userid uint64, userName string) {
	music := false
	if host == "musicfortoday.tv" {
		music = true
	}
	if host == "musicfortoday.tv:4000" {
		music = true
	}
	switch {
	case op == "add":
		if userid == 1 {
			showAddPage(w, r, op, userid, userName)
		}
	case op == "list":
		showListPage(w, r, op, userid, userName)
	case op == "exposit":
		if userid == 1 {
			showExpositPage(w, r, op, userid, userName)
		}
	case op == "edit":
		if userid == 1 {
			showEditPage(w, r, op, userid, userName)
		}
	case op == "delete":
		if userid == 1 {
			showDeletePage(w, r, op, userid, userName)
		}
	case op == "email":
		showEmailPage(w, r, op, userid, userName)
	case op == "":
		showHomePage(w, r, op, userid, userName, music)
	case op == "home":
		showHomePage(w, r, op, userid, userName, music)
	case op == "search":
		showSearchPage(w, r, op, userid, userName)
	case op == "trent":
		showTrentExpoPage(w, r, op, userid, userName)
	// case op == "gentitles":
	//	showGenTitlesPage(w, r, op, userid, userName)
	// case op == "gensql":
	// 	showGenTrentSQL(w, r, op, userid, userName)
	case op == "catchup":
		showGenTrentCatchupSQL(w, r, op, userid, userName)
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/links/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
