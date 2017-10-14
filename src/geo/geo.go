package geo

import (
	"accessdb"
	"fmt"
	"html"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// "github.com/ziutek/mymysql/mysql"
// _ "github.com/ziutek/mymysql/native" // Native engine

func getDoctype() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset=utf-8 />
`
}

func showExposition(w http.ResponseWriter, target_url string, description string) {
	fmt.Fprint(w, `
      <p>
Google+ / <a href="https://www.facebook.com/">Facebook</a> / <a href="http://www.linkedin.com/">LinkedIn</a> / <a href="http://www.twitter.com/">Twitter</a> / <a href="https://www.tumblr.com/blog/waynerad/new/link">Tumblr</a></p>
      <p><textarea cols="80" rows="20">`)
	fmt.Fprint(w, description)
	fmt.Fprint(w, `

`)
	fmt.Fprint(w, target_url)
	fmt.Fprint(w, `</textarea></p>
  </section>
`)
	return
}

func candeleteadd(w http.ResponseWriter, r *http.Request, op string) {
	var showform bool
	var error_list map[string]string
	var ui_grabbed_url string
	var ui_original_text string

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
			db := accessdb.GetDbConnection()
			defer db.Close()
			stmt, err := db.Prepare("INSERT INTO link_link (created_gmt, target_url, image_url, description, is_email, is_public, is_video) VALUES (?, ?, ?, ?, ?, ?, ?);")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			// defer stmt.Close();
			save_created_gmt := uint64(time.Now().Unix())
			save_target_url := postform["grabbed_url"][0]
			save_image_url := ""
			save_description := postform["original_text"][0]
			save_description = strings.Trim(save_description, " \r\n\t")
			save_is_email := 0
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
			stmt.Bind(save_created_gmt, save_target_url, save_image_url, save_description, save_is_email, save_is_public, save_is_video)
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
    <h1>URL grabber</h1>
	<a href="add">Next</a>
	<a href="list">List</a>
`)
			showExposition(w, save_target_url, save_description)
			fmt.Fprint(w, `
</body></html>`)
		}
		ui_grabbed_url = postform["grabbed_url"][0]
		ui_original_text = postform["original_text"][0]
	}
	if method == "GET" {
		showform = true
		ui_grabbed_url = ""
		ui_original_text = ""
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		db := accessdb.GetDbConnection()
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
<!-- link rel="stylesheet" type="text/css" href="/style.css" -->

<script>

/*jslint browser: true, devel: true */

// (C) 2013-2015 Wayne Radinsky

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

function execUrlgrab() {
    "use strict";
    var txt, i, j, k, link, setvideo;
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
    window.clearInterval(timerid);
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

// main

timerid = window.setInterval(execUrlgrab, 100);

</script>

</head>
<body>
  <section>
    <h1>URL grabber</h1>

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
<input class="biginput" name="grabbed_url" id="grabbed_url" type="text" value="`+html.EscapeString(ui_grabbed_url)+`" />
</p>
<p>
<input name="do_resize" id="do_resize" type="button" value="resize" />
<input name="submit" id="submit" type="submit" />
<input type="checkbox" checked="checked" id="email" name="email"> Email, 
<input type="checkbox" id="video" name="video"> Video
</p>

<p><textarea class="bigtextarea" name="original_text" id="original_text" cols="80" rows="20">`+html.EscapeString(ui_original_text)+`</textarea></p>
<p><input name="do_lcase" id="do_lcase" type="button" value="lcase" />
 <input name="do_underscores" id="do_underscores" type="button" value="underscores" />
 <input name="do_flip_quotes" id="do_flip_quotes" type="button" value="flip quotes" />
 <input name="do_analyze" id="do_analyze" type="button" value="analyze ascii" />
 <input name="do_single_quotes" id="do_singlequotes" type="button" value="single quotes" />
 <input name="do_quotes_hyphens" id="do_quoteshyphens" type="button" value="quotes hyphens" />
</p>
<p><textarea class="bigtextarea" name="analyze_result" id="analyze_result" cols="80" rows="20"></textarea></p>

</form>

  </section>
<script>

var startElement = document.getElementById('grab_url');
startElement.addEventListener('click', execUrlgrab, true);

var startElement = document.getElementById('do_lcase');
startElement.addEventListener('click', execDoTheLcase, true);

var startElement = document.getElementById('do_flip_quotes');
startElement.addEventListener('click', execDoFlipQuotes, true);

var startElement = document.getElementById('do_underscores');
startElement.addEventListener('click', execDoUnderscores, true);

var startElement = document.getElementById('do_resize');
startElement.addEventListener('click', execDoTheResize, true);

var startElement = document.getElementById('do_analyze');
startElement.addEventListener('click', execDoTheAnalyze, true);

var startElement = document.getElementById('do_quoteshyphens');
startElement.addEventListener('click', execDoTheQuotesHyphens, true);

var startElement = document.getElementById('do_singlequotes');
startElement.addEventListener('click', execDoTheSingleQuotes, true);

</script>
`)
		}
		fmt.Fprint(w, `
</body>
</html>`)
		return
	}

}

func candeletelist(w http.ResponseWriter, r *http.Request, op string) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List of URLs</title>
</head>
<body>
  <section>
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
	_, show_all := getform["showall"]
	if show_all {
		sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC;"
	} else {
		_, show_email := getform["emailonly"]
		if show_email {
			// email_cutoff
			sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (is_email = 1) ORDER BY id_lnk DESC;"
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
				_, do_search := getform["search"]
				if do_search {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE (description LIKE '%3D print%') OR (description LIKE '%3D-print%') ORDER BY id_lnk;"
				} else {
					sql = "SELECT id_lnk, created_gmt, target_url, image_url, description FROM link_link WHERE 1 ORDER BY id_lnk DESC LIMIT 0, 50;"
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

func candeleteexposit(w http.ResponseWriter, r *http.Request, op string) {
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
    <h1>Exposition</h1>
`)
	db := accessdb.GetDbConnection()
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
			// image_url = row.Str(3)
			description := row.Str(4)
			showExposition(w, target_url, description)
		}
	}
}

func checkedStr(is_set bool) string {
	if is_set {
		return `checked="checked"`
	}
	return ""
}

func candeleteedit(w http.ResponseWriter, r *http.Request, op string) {
	var error_occurred bool
	var error_list map[string]string
	var linkid uint64
	var linkstr string
	var ui_is_email bool
	var ui_is_public bool
	var ui_is_video bool
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
		db := accessdb.GetDbConnection()
		res, err := db.Start("SELECT id_lnk, created_gmt, target_url, image_url, description, is_email, is_public, is_video FROM link_link WHERE id_lnk=" + strconv.FormatUint(linkid, 10) + " ORDER BY id_lnk DESC;")
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
			db := accessdb.GetDbConnection()
			defer db.Close()
			stmt, err := db.Prepare("UPDATE link_link SET target_url=?, image_url=?, description=?, is_email=?, is_public=?, is_video=? WHERE id_lnk=?;")
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
			update_image_url := ""
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
			stmt.Bind(update_target_url, update_image_url, update_description, update_is_email, update_is_public, update_is_video, update_id_lnk)
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
</table>

<p><input name="submit" id="submit" type="submit" />

</form>

</body></html`)
	}
	return
}

func candeletedelete(w http.ResponseWriter, r *http.Request, op string) {
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
		db := accessdb.GetDbConnection()
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
		fmt.Fprint(w, getDoctype())
		fmt.Fprint(w, `<title>URL saver</title>
</head><body>
  <section>
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

func candeleteemail(w http.ResponseWriter, r *http.Request, op string) {
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
    <h1>For Emails</h1>
<textarea rows="100" cols="80">
=== News bits ===

(In reverse chronological order, except videos which are in their own section at the bottom)


`)
	db := accessdb.GetDbConnection()
	defer db.Close()

	count := 0

	// email_cutoff
	res, err := db.Start("SELECT target_url, description FROM link_link WHERE (is_email = 1) AND (is_video = 0) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
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
	res, err = db.Start("SELECT target_url, description FROM link_link WHERE (is_email=1) AND (is_video=1) AND (id_lnk < " + strconv.FormatUint(cutoff, 10) + ") ORDER BY id_lnk DESC;")
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

func planet(w http.ResponseWriter, r *http.Request, op string) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand Planet</title>
</head>
<body>
  <section>
    <h1>Geo Rand Planet</h1>
`)
	rand.Seed(time.Now().UnixNano())
	y := (math.Asin((rand.Float64()*2.0)-1.0) / (math.Pi / 2.0)) * 90.0
	x := (rand.Float64() * 360.0) - 180.0
	fmt.Fprintln(w, "<p>Cut & paste for Google Maps:</p>")
	fmt.Fprintln(w, `<input type="text" style="width:400px;" value="`+strconv.FormatFloat(y, 'f', -1, 64)+","+strconv.FormatFloat(x, 'f', -1, 64)+`" /></p>`)
	fmt.Fprintln(w, "</body></html>")
}

func streetview(w http.ResponseWriter, r *http.Request, op string) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand Street View</title>
</head>
<body>
  <section>
    <h1>Geo Rand Street View</h1>
`)
	type City struct {
		name  string
		lat1  float64
		lat2  float64
		long1 float64
		long2 float64
	}
	now := time.Now()
	asNumber := now.UnixNano()
	rand.Seed(asNumber)
	citiesList := []City{{"Tokyo", 35.5, 35.88, 139.5, 140.0}, {"Jakarta", -6.24, -6.12, 106.71, 106.92}, {"Seoul", 37.48, 37.64, 126.89, 127.20}, {"Mexico City", 19.29, 19.68, -99.23, -99.06}, {"New York", 40.57, 40.88, -74.20, -73.80}, {"Sao Paulo", -23.71, -23.45, -46.78, -46.41}, {"Osaka", 34.60, 34.81, 135.35, 135.64}, {"Los Angeles", 33.73, 34.10, -118.40, -117.77}, {"Dhaka", 23.68, 23.81, 90.36, 90.45}, {"Moscow", 55.70, 55.79, 37.53, 37.74}, {"Buenos Aires", -34.80, -34.49, -58.61, -58.25}, {"Bangkok", 13.61, 13.79, 100.47, 100.64}, {"London", 51.46, 51.53, -0.19, 0.01}, {"Johannesburg", -26.22, -26.18, 28.02, 28.06}, {"Rio de Janeiro", -22.93, -22.76, -43.43, -43.17}, {"Paris", 48.81, 48.88, 2.27, 2.40}, {"Lima", -12.10, -12.01, -77.06, -76.97}, {"Chicago", 41.80, 41.85, -87.63, -87.62}, {"Bogota", 4.56, 4.65, -74.16, -74.06}, {"Washington DC", 38.88, 38.92, -77.01, -76.98}, {"Ho Chi Minh City", 10.73, 10.80, 106.59, 106.70}, {"San Francisco", 37.739694, 37.803983, -122.446680, -122.385215}, {"Kuala Lumpur", 3.01, 3.32, 101.58, 101.77}, {"Sydney", -33.94, -33.87, 151.08, 151.24}, {"Denver", 39.565420, 39.918791, -105.150862, -104.755332}, {"Boulder", 39.969426, 40.051865, -105.290948, -105.225628}, {"Oaxaca", 17.04, 17.09, -96.75, -96.70}, {"Miami", 25.75, 26.33, -80.29, -80.18}, {"Hong Kong", 22.27, 22.33, 114.13, 114.22}, {"Taipei", 24.99, 25.07, 121.43, 121.56}, {"Toronto", 43.64, 43.67, -79.41, -79.35}, {"Madrid", 40.42, 40.47, -3.72, -3.60}, {"Boston", 42.33, 42.42, -71.12, -71.01}, {"Houston", 29.66, 29.80, -95.56, -95.23}, {"Detroit", 42.36, 42.62, -83.11, -82.88}, {"Atlanta", 33.72, 33.79, -84.42, -83.36}, {"St. Petersburg", 59.86, 60.01, 30.25, 30.43}, {"Milano", 45.43, 45.51, 9.13, 9.25}, {"Dortmund (Ruhrgebiet)", 51.48, 51.55, 7.41, 7.52}, {"Guadalajara", 20.66, 20.73, -103.42, -103.27}, {"Belo Horizonte", -19.95, -19.80, -44.03, -43.89}, {"Prague", 50.03, 50.11, 14.40, 14.53}, {"Krasnodar", 45.01, 45.07, 38.93, 39.03}, {"Kyiv", 50.39, 50.52, 30.38, 30.66}, {"Bucharest", 44.37, 44.48, 26.03, 26.19}, {"Budapest", 47.41, 47.56, 19.07, 19.23}, {"Sofia", 42.65, 42.73, 23.27, 23.41}, {"Tel Aviv", 32.06, 32.09, 34.77, 34.85}, {"Helsinki", 60.17, 60.20, 24.92, 24.96}, {"Stockholm", 59.30, 59.35, 17.99, 18.10}, {"Oslo", 59.92, 59.93, 10.70, 10.87}}
	numCities := len(citiesList)
	cityPick := int(math.Floor(rand.Float64() * float64(numCities)))
	cityInfo := citiesList[cityPick]
	y := (rand.Float64() * (cityInfo.lat2 - cityInfo.lat1)) + cityInfo.lat1
	x := (rand.Float64() * (cityInfo.long2 - cityInfo.long1)) + cityInfo.long1
	fmt.Fprintln(w, "<p>Cut & paste for Google Maps:</p>")
	fmt.Fprintln(w, `<input type="text" style="width:400px;" value="`+strconv.FormatFloat(y, 'f', -1, 64)+","+strconv.FormatFloat(x, 'f', -1, 64)+`" /></p>`)
	fmt.Fprintln(w, "</body></html>")
}

func dice(w http.ResponseWriter, r *http.Request, op string) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand Dice</title>
</head>
<body>
  <section>
    <h1>Dice</h1>
`)
	method := r.Method
	if method == "POST" {
		now := time.Now()
		asNumber := now.UnixNano()
		rand.Seed(asNumber)
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("parseform failed")
		}
		postform := r.Form
		minstr := postform["min"][0]
		maxstr := postform["max"][0]
		minval, err := strconv.ParseInt(minstr, 10, 64)
		maxval, err := strconv.ParseInt(maxstr, 10, 64)
		pick := int64(math.Floor(rand.Float64()*float64(maxval-minval+1))) + minval
		fmt.Fprintln(w, strconv.FormatInt(pick, 10))
	} else {
		fmt.Fprintln(w, `
<form action="dice" method="post">
	<table border="0" cellpadding="3">
		<tr><td> Min </td><td> <input type="text" name="min" /> </td></tr>
		<tr><td> Max </td><td> <input type="text" name="max" /> </td></tr>
		<tr><td colspan="2"> <input type="submit" /> </td></tr>
	</table>
</form>
`)
	}
	fmt.Fprintln(w, "</body></html>")
}

func distance(w http.ResponseWriter, r *http.Request, op string, userid uint64) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand Distance</title>
</head>
<body>
  <section>
    <h1>Distance</h1>
`)
	method := r.Method
	if method == "POST" {
		now := time.Now()
		asNumber := now.UnixNano()
		rand.Seed(asNumber)
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("parseform failed")
		}
		postform := r.Form
		diststr := postform["distance"][0]
		distval, err := strconv.ParseFloat(diststr, 64)
		// now we have our input, let's calculate
		// centery := 39.690517 // the house is the centerpoint
		// centerx := -104.917979
		var centery float64
		var centerx float64

		db := accessdb.GetDbConnection()
		defer db.Close()
		sql := "SELECT gps_lat, gps_long FROM login_user WHERE id_user = ?;"
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
		for _, row := range rows {
			centery = row.Float(0)
			centerx = row.Float(1)
		}

		// the grand junction house is at 39.08720277,-108.63616944
		radius := distval / 100.0 // radius in GPS units (calibrated to latitudal/equatorial longitudal)
		// circumference of Earth at the equator is 24,902 miles/40,075 km
		var deltay float64
		var deltax float64
		perimeter := false
		if op == "distp" {
			perimeter = true
		}
		if perimeter {
			deltay = (rand.Float64() * 2.0) - 1.0
			deltax = (rand.Float64() * 2.0) - 1.0
			hypotenuse := math.Sqrt((deltay * deltay) + (deltax * deltax))
			deltay = deltay / hypotenuse
			deltax = deltax / hypotenuse
			// At this point we have a point on a circle
		} else {
			keepgoing := true
			for keepgoing {
				deltay = (rand.Float64() * 2.0) - 1.0
				deltax = (rand.Float64() * 2.0) - 1.0
				if ((deltax * deltax) + (deltay * deltay)) <= 1.0 {
					keepgoing = false
				}
				// deltay = 1.0
				// deltax = 0.0
			}
			// At this point we have a point in a circle
		}
		resulty := centery + (deltay * radius)
		// use lattitude to correctly compute longitude
		resultx := centerx + (deltax * (radius / math.Cos((resulty/180.0)*math.Pi)))

		fmt.Fprintln(w, "<p>Cut & paste for Google Maps:</p>")
		fmt.Fprintln(w, `<input type="text" style="width:400px;" value="`+strconv.FormatFloat(resulty, 'f', -1, 64)+","+strconv.FormatFloat(resultx, 'f', -1, 64)+`" /></p>`)
	} else {
		action := "dist"
		if op == "distp" {
			action = "distp"
		}
		fmt.Fprintln(w, `
<form action="`+action+`" method="post">
	<table border="0" cellpadding="3">
		<tr><td> Distance </td><td> <input type="text" name="distance" /> </td></tr>
		<tr><td colspan="2"> <input type="submit" /> </td></tr>
	</table>
</form>
`)
	}
	fmt.Fprintln(w, "</body></html>")
}

func eightball(w http.ResponseWriter, r *http.Request, op string) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand 8 ball</title>
</head>
<body>
  <section>
`)
	now := time.Now()
	asNumber := now.UnixNano()
	rand.Seed(asNumber)
	answers := []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}
	fmt.Fprintln(w, "<p>Magic 8-Ball says: ", answers[rand.Intn(len(answers))], "</p>")
	fmt.Fprintln(w, "</body></html>")
}

func showList(w http.ResponseWriter) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>Geo Rand List</title>
</head>
<body>
  <section>
    <ul>
        <li><a href="planet">Planet</a></li>
        <li><a href="streetview">Street View</a></li>
        <li><a href="dice">Dice</a></li>
        <li><a href="dist">Distance (within)</a></li>
        <li><a href="distp">Distance (perimeter)</a></li>
        <li><a href="8ball">8 Ball</a></li>
    </ul>
</body></html>`)

}

func Handler(w http.ResponseWriter, r *http.Request, op string, userid uint64, userName string) {
	switch {

	case op == "planet":
		planet(w, r, op)
	case op == "streetview":
		streetview(w, r, op)
	case op == "dice":
		dice(w, r, op)
	case op == "dist":
		distance(w, r, op, userid)
	case op == "distp":
		distance(w, r, op, userid)
	case op == "8ball":
		eightball(w, r, op)

	case op == "list":
		showList(w)
	// case op == "colorado":
	// colorado(w, r, op)
	// case op == "denver":
	// denver(w, r, op)
	// case op == "boulder":
	// boulder(w, r, op)
	// case op == "denverboulder":
	// denverboulder(w, r, op)
	// case op == "usa":
	// usa(w, r, op)
	// case op == "wayne":
	// wayne(w, r, op)
	default:
		fmt.Fprintln(w, "Could not find page:", op)
	}
}
