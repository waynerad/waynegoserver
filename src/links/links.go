package links

import (
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

func htm(str string) string {
	return html.EscapeString(str)
}

func showLinksMenuBar(w http.ResponseWriter, userName string) {
	fmt.Fprint(w, `
<p><a href="add">Add</a>
<a href="list">List</a> &middot;
<font color="grey">`+htm(userName)+`</font>
</p>`)
}

func showExposition(w http.ResponseWriter, targetUrl string, imageUrl string, description string) {
	fmt.Fprint(w, `
      <p>
Google+ / <a href="https://www.facebook.com/">Facebook</a> / <a href="http://www.linkedin.com/">LinkedIn</a> / <a href="http://www.twitter.com/">Twitter</a> / <a href="https://www.tumblr.com/blog/waynerad/new/link">Tumblr</a></p>
      <p><textarea cols="80" rows="20">`)
	fmt.Fprint(w, description)
	fmt.Fprint(w, `

`)
	fmt.Fprint(w, targetUrl)
	fmt.Fprint(w, `</textarea></p>
<p> <img src="`+imageUrl+`" alt="Thumbnail" /> </p>
  </section>
`)
	return
}

func add(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var showform bool
	var error_list map[string]string
	var ui_grabbed_url string
	var ui_original_text string
	var ui_image_url string

	error_occurred := false
	method := r.Method
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("parseform failed")
		}
		postform := r.Form
		if (postform["grabbed_url"][0] == "") || (postform["original_text"][0] == "") {
			showform = true
			error_occurred = true
			error_list = make(map[string]string)
			if postform["grabbed_url"][0] == "" {
				error_list["grabbed_url"] = "Grabbed URL is empty"
			}
			if postform["original_text"][0] == "" {
				error_list["original_text"] = "Original text is empty"
			}
		}
		if error_occurred == false {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			stmt, err := db.Prepare("INSERT INTO link_link (created_gmt, target_url, image_url, description, is_email, is_public, is_video, is_pdf) VALUES (?, ?, ?, ?, ?, ?, ?, ?);")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// defer stmt.Close();
			save_created_gmt := uint64(time.Now().Unix())
			save_target_url := postform["grabbed_url"][0]
			save_image_url := postform["image_url"][0]
			save_description := postform["original_text"][0]
			save_description = strings.Trim(save_description, " \r\n\t")
			save_is_email := 0
			save_is_pdf := 0
			_, ok := postform["email"]
			if ok {
				save_is_email = 1
			}
			save_is_public := save_is_email
			save_is_video := 0
			_, ok = postform["video"]
			if ok {
				save_is_video = 1
			}
			_, ok = postform["pdf"]
			if ok {
				save_is_pdf = 1
			}
			stmt.Bind(save_created_gmt, save_target_url, save_image_url, save_description, save_is_email, save_is_public, save_is_video, save_is_pdf)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}

			header := w.Header()
			header.Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, getDoctype())
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
			showExposition(w, save_target_url, save_image_url, save_description)
			fmt.Fprint(w, `
</body></html>`)
		}
		ui_grabbed_url = postform["grabbed_url"][0]
		ui_image_url = postform["image_url"][0]
		ui_original_text = postform["original_text"][0]
	}
	if method == "GET" {
		showform = true
		ui_grabbed_url = ""
		ui_image_url = ""
		ui_original_text = ""
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
		sql := "SELECT COUNT(*) FROM link_link WHERE is_email=1;"
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
		email_count := row.Uint64(0)
		fmt.Fprint(w, `<title>URL saver</title>
<link rel="stylesheet" type="text/css" href="/style.css">

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

function execDoTheQuotesHyphens() {
    "use strict";
    var str, result, i, a;
    str = document.getElementById("original_text").value;
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
    document.getElementById("original_text").value = result;
}

function execDoTheSingleQuotes() {
    "use strict";
    var str;
    str = document.getElementById("original_text").value;
    str = strReplaceAll(str, '"', "'");
    document.getElementById("original_text").value = str;
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

<p>Email count: `+html.EscapeString(strconv.FormatUint(email_count, 10))+`

<form action="add" method="post">
`)
		if error_occurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for err_on, err_msg := range error_list {
				fmt.Fprintln(w, "<li>"+html.EscapeString(err_on)+": "+html.EscapeString(err_msg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		if email_count < 80 {
			fmt.Fprint(w, `

<p>
<input name="grab_url" id="grab_url" type="button" value="Grab URL" />
URL: <input class="biginput" name="grabbed_url" id="grabbed_url" type="text" value="`+html.EscapeString(ui_grabbed_url)+`" />
</p>
<p>
<input name="do_resize" id="do_resize" type="button" value="resize" />
<input name="submit" id="submit" type="submit" />
<input type="checkbox" checked="checked" id="email" name="email"> Email, 
<input type="checkbox" id="video" name="video"> Video
<input type="checkbox" id="pdf" name="pdf"> PDF
</p>
<p>
<input name="grab_image" id="grab_image" type="button" value="Grab URL" />
Image URL: <input class="biginput" name="image_url" id="image_url" type="text" value="`+html.EscapeString(ui_image_url)+`" /> </p>
<p><textarea class="bigtextarea" name="original_text" id="original_text" cols="80" rows="20">`+html.EscapeString(ui_original_text)+`</textarea></p>
<p><input name="do_lcase" id="do_lcase" type="button" value="lcase" />
 <input name="do_underscores" id="do_underscores" type="button" value="underscores" />
 <input name="do_flip_quotes" id="do_flip_quotes" type="button" value="flip quotes" />
 <input name="do_analyze" id="do_analyze" type="button" value="analyze ascii" />
 <input name="do_single_quotes" id="do_singlequotes" type="button" value="single quotes" />
 <input name="do_quotes_hyphens" id="do_quoteshyphens" type="button" value="quotes hyphens" />
 <input name="do_both" id="do_both" type="button" value="both" />
</p>
<p><textarea class="bigtextarea" name="analyze_result" id="analyze_result" cols="80" rows="20"></textarea></p>

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

func list(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of URLs</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>List of URLs</h1>
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
	getform := r.Form
	_, show_all := getform["showall"]
	if show_all {
		sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC;"
	} else {
		_, show_email := getform["emailonly"]
		if show_email {
			// email_cutoff
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
			_, video_only := getform["videoonly"]
			if video_only {
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
			} else {
				_, pdf_only := getform["pdfonly"]
				if pdf_only {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_pdf = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
				} else {

					_, do_search := getform["search"]
					if do_search {
						sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (description LIKE '%3D print%') OR (description LIKE '%3D-print%') ORDER BY id_lnk;"
					} else {
						sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC LIMIT 0, 50;"
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
			// created_gmt = row.Uint64(1)
			target_url := row.Str(2)
			// image_url = row.Str(3)
			description := row.Str(4)
			fmt.Fprint(w, `<p>`)
			fmt.Fprint(w, html.EscapeString(description))
			fmt.Fprint(w, `<br />`)
			fmt.Fprint(w, "<a href="+target_url+">"+html.EscapeString(target_url)+"</a>")
			fmt.Fprint(w, ` &middot; <a href="exposit?link=`+strconv.FormatUint(linkid, 10)+`">exposit</a> &middot; <a href="edit?link=`+strconv.FormatUint(linkid, 10)+`">edit</a> &middot; <a href="delete?link=`+strconv.FormatUint(linkid, 10)+`">delete</a>
</p>`)
		}
	}
	fmt.Fprint(w, `
  </section>
</body>
</html>`)
}

func exposit(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
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
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>URL exposition</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>Exposition</h1>
`)
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer db.Close()
	res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
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
			// created_gmt = row.Uint64(1)
			target_url := row.Str(2)
			image_url := row.Str(3)
			description := row.Str(4)
			showExposition(w, target_url, image_url, description)
		}
	}
}

func checkedStr(is_set bool) string {
	if is_set {
		return `checked="checked"`
	}
	return ""
}

func edit(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var error_occurred bool
	var error_list map[string]string
	var linkid uint64
	var linkstr string
	var ui_is_email bool
	var ui_is_public bool
	var ui_is_video bool
	var ui_is_pdf bool
	var ui_created string
	var ui_target_url string
	var ui_image_url string
	var ui_description string
	method := r.Method
	error_list = make(map[string]string)
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
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, is_email, is_public, is_video, is_pdf FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
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
			created_gmt := row.Uint64(1)
			target_url := row.Str(2)
			image_url := row.Str(3)
			description := row.Str(4)
			time_object := time.Unix(int64(created_gmt), 0)
			// time_object.Format("Mon Jan 2 15:04:05 MST 2006  (MST is GMT-0700)
			createstr := time_object.Format("2006-01-02 15:04:05")
			ui_created = createstr
			ui_target_url = target_url
			ui_image_url = image_url
			ui_description = description
			ui_is_email = row.Bool(5)
			ui_is_public = row.Bool(6)
			ui_is_video = row.Bool(7)
			ui_is_pdf = row.Bool(8)
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
		ui_created = postform["created"][0]
		ui_target_url = postform["target_url"][0]
		ui_image_url = postform["image_url"][0]
		ui_description = postform["description"][0]
		_, ui_is_email = postform["email"]
		_, ui_is_public = postform["public"]
		_, ui_is_video = postform["video"]
		_, ui_is_pdf = postform["pdf"]
		if (postform["target_url"][0] == "") || (postform["description"][0] == "") {
			showform = true
			error_occurred = true
			if postform["target_url"][0] == "" {
				error_list["target_url"] = "Target URL is empty"
			}
			if postform["description"][0] == "" {
				error_list["description"] = "Description is empty"
			}
		}
		if error_occurred == false {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			defer db.Close()
			stmt, err := db.Prepare("UPDATE link_link SET target_url=?, image_url=?, description=?, is_email=?, is_public=?, is_video=?, is_pdf=? WHERE id_lnk=?;")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// defer stmt.Close();
			linkstr = postform["link"][0]
			linkid, err = strconv.ParseUint(linkstr, 10, 64)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// update_created_gmt = uint64(time.Now().Unix())
			update_target_url := postform["target_url"][0]
			update_image_url := postform["image_url"][0]
			update_description := postform["description"][0]
			update_description = strings.Trim(update_description, " \r\n\t")
			update_id_lnk := linkid
			update_is_email := 0
			_, ok := postform["email"]
			if ok {
				update_is_email = 1
			}
			_, ok = postform["public"]
			update_is_public := 0
			if ok {
				update_is_public = 1
			}
			_, ok = postform["video"]
			update_is_video := 0
			if ok {
				update_is_video = 1
			}
			_, ok = postform["pdf"]
			update_is_pdf := 0
			if ok {
				update_is_pdf = 1
			}
			stmt.Bind(update_target_url, update_image_url, update_description, update_is_email, update_is_public, update_is_video, update_is_pdf, update_id_lnk)
			_, _, err = stmt.Exec()
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
		}
		http.Redirect(w, r, "exposit?link="+linkstr, 302)
		return
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")

		if error_occurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for err_on, err_msg := range error_list {
				fmt.Fprintln(w, "<li>"+html.EscapeString(err_on)+": "+html.EscapeString(err_msg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, getDoctype())
		fmt.Fprint(w, `<title>URL saver</title>
</head><body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Edit URL</h1>

<form action="edit" method="post">

<table border="1">
<tr><td> Created </td><td> <input name="link" id="link" value="`+strconv.FormatUint(linkid, 10)+`" type="hidden" /> <input name="created" id="created" type="text" value="`+html.EscapeString(ui_created)+`" readonly="readonly" /> </td></tr>
<tr><td> Target URL </td><td> <input name="target_url" id="target_url" type="text" value="`+html.EscapeString(ui_target_url)+`" style="width:400px;" /> </td></tr>
<tr><td> Image URL </td><td> <input name="image_url" id="image_url" type="text" value="`+html.EscapeString(ui_image_url)+`" style="width:400px;" /> </td></tr>
<tr><td> Description </td><td> <textarea name="description" id="description" rows="20" cols="80">`+html.EscapeString(ui_description)+`</textarea> </td></tr>
<tr><td> Email </td><td> <input type="checkbox" name="email" id="email" `+checkedStr(ui_is_email)+`> Email </td></tr>
<tr><td> Public </td><td> <input type="checkbox" name="public" id="public" `+checkedStr(ui_is_public)+`> Public </td></tr>
<tr><td> Video </td><td> <input type="checkbox" name="video" id="video" `+checkedStr(ui_is_video)+`> Video </td></tr>
<tr><td> PDF </td><td> <input type="checkbox" name="pdf" id="video" `+checkedStr(ui_is_pdf)+`> PDF </td></tr>
</table>

<p><input name="submit" id="submit" type="submit" />

</form>

</body></html`)
	}
	return
}

func delete(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var linkid uint64
	var ui_target_url string
	var ui_description string
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
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
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
			// created_gmt = row.Uint64(1)
			target_url := row.Str(2)
			// image_url = row.Str(3)
			description := row.Str(4)
			// time_object = time.Unix(int64(created_gmt), 0)
			// time_object.Format("Mon Jan 2 15:04:05 MST 2006  (MST is GMT-0700)
			// createstr = time_object.Format("2006-01-02 15:04:05")
			// ui_created = createstr
			ui_target_url = target_url
			// ui_image_url = image_url
			ui_description = description
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
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
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
		fmt.Fprint(w, getDoctype())
		fmt.Fprint(w, `<title>URL saver</title>
</head><body>
  <section>
`)
		showLinksMenuBar(w, userName)
		fmt.Fprint(w, `
    <h1>Delete URL</h1>

<form action="delete" method="post">

<p> Delete link to: <a href="`+html.EscapeString(ui_target_url)+`">`+html.EscapeString(ui_target_url)+`</a>? </p>

<p> `+html.EscapeString(ui_description)+`</p>

<p><input type="submit" id="submit" name="submit" value="Delete" />
<input name="link" id="link" value="`+strconv.FormatUint(linkid, 10)+`" type="hidden" />
</p>

</form>

</body></html`)
	}
	return
}

func email(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	var cutoff uint64
	double_linefeeds := true
	method := r.Method
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		getform := r.Form
		_, nodouble := getform["nodouble"]
		double_linefeeds = !nodouble
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
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of URLs</title>
</head>
<body>
  <section>
`)
	showLinksMenuBar(w, userName)
	fmt.Fprint(w, `
    <h1>For Emails</h1>
<textarea rows="100" cols="80">
=== News bits ===

(In reverse chronological order, except videos which are in their own section at the bottom)


`)
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer db.Close()

	count := 0

	// email_cutoff
	res, err := db.Start("SELECT target_url, description FROM link_link WHERE (is_email = 1) AND (is_video = 0) AND (is_pdf = 0) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
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
			target_url := row.Str(0)
			description := row.Str(1)
			fmt.Fprint(w, "\n")
			if double_linefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, html.EscapeString(strings.Replace(strings.Replace(description, "\r", "", -1), "\n", " ", -1)))
			fmt.Fprint(w, "\n")
			if double_linefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, html.EscapeString(target_url)+"\n")
			count++
		}
	}
	fmt.Fprint(w, `



=== Video ===


`)
	// email_cutoff
	res, err = db.Start("SELECT target_url, description FROM link_link WHERE (is_email = 1) AND (is_video = 1) AND (is_pdf = 0) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
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
			target_url := row.Str(0)
			description := row.Str(1)
			fmt.Fprint(w, "\n")
			if double_linefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, html.EscapeString(strings.Replace(strings.Replace(description, "\r", "", -1), "\n", " ", -1)))
			fmt.Fprint(w, "\n")
			if double_linefeeds {
				fmt.Fprint(w, "\n")
			}
			fmt.Fprint(w, html.EscapeString(target_url)+"\n")
			count++
		}
	}
	fmt.Fprint(w, "</textarea>")
	fmt.Fprint(w, "<p>Count: "+html.EscapeString(strconv.FormatInt(int64(count), 10))+`</p>
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

func homepage(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string, showMusic bool) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
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
	getform := r.Form
	_, show_all := getform["showall"]
	if show_all {
		sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC;"
	} else {
		_, show_email := getform["emailonly"]
		if show_email {
			// email_cutoff
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
			_, video_only := getform["videoonly"]
			if video_only {
				sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_video = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
			} else {
				_, pdf_only := getform["pdfonly"]
				if pdf_only {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_pdf = 1) ORDER BY id_lnk DESC LIMIT 0, 200;"
				} else {

					_, do_search := getform["search"]
					if do_search {
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
			// created_gmt = row.Uint64(1)
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
			fmt.Fprint(w, strings.Replace(html.EscapeString(description), "\n", "<br />", 32768))
			if userid == 1 {
				fmt.Fprint(w, `<br />`)
				fmt.Fprint(w, "<a href="+targetUrl+">"+html.EscapeString(targetUrl)+"</a>")
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

func Handler(w http.ResponseWriter, r *http.Request, host string, op string, userid uint64, userName string) {
	fmt.Println("op in links Handler", op)
	fmt.Println("userid", userid)
	fmt.Println("userName", userName)
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
			add(w, r, op, userid, userName)
		}
	case op == "list":
		list(w, r, op, userid, userName)
	case op == "exposit":
		if userid == 1 {
			exposit(w, r, op, userid, userName)
		}
	case op == "edit":
		if userid == 1 {
			edit(w, r, op, userid, userName)
		}
	case op == "delete":
		if userid == 1 {
			delete(w, r, op, userid, userName)
		}
	case op == "email":
		email(w, r, op, userid, userName)
	case op == "":
		homepage(w, r, op, userid, userName, music)
	case op == "home":
		homepage(w, r, op, userid, userName, music)
	default:
		// fmt.Fprintln(w, "Could not find page:", op)
		filename := "/home/ec2-user/wayneserver/staticappcontent/links/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}
