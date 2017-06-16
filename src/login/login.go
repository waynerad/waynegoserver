package login

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// "os"

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

func floatToString(x float64) string {
	return strconv.FormatFloat(x, 'g', -1, 64)
}

func isFloat(x string) bool {
	_, err := strconv.ParseFloat(x, 64)
	if err != nil {
		return false
	}
	return true
}

func stringToFloat(x string) float64 {
	fv, err := strconv.ParseFloat(x, 64)
	if err != nil {
		panic("Cannot parse float")
	}
	return fv
}

func base64encode(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

func base64decode(str string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println("error:", err)
		panic("point 66")
	}
	return data, err
}

func generateSalt(chars int) (string, error) {
	bsalt := make([]byte, chars)
	_, err := io.ReadFull(rand.Reader, bsalt)
	str := base64encode(bsalt)
	return str, err
}

func computePasswordHash(password string, salt string) (string, error) {
	buf := []byte(password)
	bsalt, err := base64decode(salt)
	if err != nil {
		return "", err
	}
	c := sha512.New()
	c.Write(buf)
	c.Write(bsalt)
	the_hash := c.Sum(nil)
	return base64encode(the_hash), err
	// return fmt.Sprintf("%x", the_hash)
}

type uiFields struct {
	created        string
	email          string
	password       string
	fname          string
	lname          string
	gpsLat         string
	gpsLong        string
	timeZoneOffset string
}

// salt           string

func editAccount(w http.ResponseWriter, r *http.Request, operation string, newAccount bool) {
	method := r.Method
	errorList := make(map[string]string)
	showform := true
	var userid int64
	var uiFrm uiFields
	var errorOccurred bool
	if method == "GET" {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 116")
		}
		postform := r.Form
		_, ok := postform["user"]
		if ok {
			userstr := postform["user"][0]
			userid, err = strconv.ParseInt(userstr, 10, 64)
		} else {
			if newAccount {
				userid = 0
			} else {
				fmt.Fprintln(w, "user id not specified")
				return
			}
		}
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 133")
		}
		db, err := getDbConnection()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 138")
		}
		if userid == 0 {
			uiFrm.created = ""
			uiFrm.email = ""
			uiFrm.password = ""
			// uiFrm.salt = ""
			uiFrm.fname = ""
			uiFrm.lname = ""
			uiFrm.gpsLat = "0.0"
			uiFrm.gpsLong = "0.0"
			uiFrm.timeZoneOffset = "-7"
		} else {
			res, err := db.Start("SELECT id_user, created_gmt, email, password, salt, fname, lname, gps_lat, gps_long, time_zone_offset FROM login_user WHERE id_user=" + strconv.FormatInt(userid, 10) + " ORDER BY id_user;")
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 154")
			}
			// defer res.Close();
			row, err := res.GetRow()
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 160")
			}
			if row == nil {
				fmt.Fprintln(w, "What? row == nil. Where is it? How did we get here?")
				panic("point 164")
			} else {
				var userData struct {
					userid         uint64
					createdGmt     uint64
					email          string
					password       string
					salt           string
					fname          string
					lname          string
					gpsLat         float64
					gpsLong        float64
					timeZoneOffset float64
				}
				userData.userid = row.Uint64(0)
				userData.createdGmt = row.Uint64(1)
				userData.email = row.Str(2)
				userData.password = row.Str(3)
				userData.salt = row.Str(4)
				userData.fname = row.Str(5)
				userData.lname = row.Str(6)
				userData.gpsLat = row.Float(7)
				userData.gpsLong = row.Float(8)
				userData.timeZoneOffset = row.Float(9)
				time_object := time.Unix(int64(userData.createdGmt), 0)
				// time_object.Format("Mon Jan 2 15:04:05 MST 2006  (MST is GMT-0700)
				createstr := time_object.Format("2006-01-02 15:04:05")
				uiFrm.created = createstr
				uiFrm.email = userData.email
				uiFrm.password = userData.password
				// uiFrm.salt = userData.salt
				uiFrm.fname = userData.fname
				uiFrm.lname = userData.lname
				uiFrm.gpsLat = floatToString(userData.gpsLat)
				uiFrm.gpsLong = floatToString(userData.gpsLong)
				uiFrm.timeZoneOffset = floatToString(userData.timeZoneOffset)
			}
		}
	}
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 208")
		}
		postform := r.Form
		if newAccount {
			userid = 0
		} else {
			userstr := postform["user"][0]
			userid, err = strconv.ParseInt(userstr, 10, 64)
		}
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 219")
		}

		_, ok := postform["created"]
		if ok {
			uiFrm.created = postform["created"][0]
		} else {
			uiFrm.created = ""
		}
		uiFrm.email = postform["email"][0]
		uiFrm.password = postform["password"][0]
		// uiFrm.salt = postform["salt"][0]
		uiFrm.fname = postform["fname"][0]
		uiFrm.lname = postform["lname"][0]
		uiFrm.gpsLat = postform["gps_lat"][0]
		uiFrm.gpsLong = postform["gps_long"][0]
		uiFrm.timeZoneOffset = postform["time_zone_offset"][0]
		if uiFrm.email == "" {
			showform = true
			errorOccurred = true
			errorList["email"] = "Email is empty"
		}
		if uiFrm.password == "" {
			showform = true
			errorOccurred = true
			errorList["password"] = "Password is empty"
		}
		if uiFrm.fname == "" {
			showform = true
			errorOccurred = true
			errorList["fname"] = "First name is empty"
		}
		if uiFrm.lname == "" {
			showform = true
			errorOccurred = true
			errorList["lname"] = "Last name is empty"
		}
		if !isFloat(uiFrm.gpsLat) {
			showform = true
			errorOccurred = true
			errorList["gps_lat"] = "Lattitude is empty or non-numeric"
		}
		if !isFloat(uiFrm.gpsLong) {
			showform = true
			errorOccurred = true
			errorList["gps_long"] = "Longitude is empty or non-numeric"
		}
		if !isFloat(uiFrm.timeZoneOffset) {
			showform = true
			errorOccurred = true
			errorList["time_zone_offset"] = "Time zone offset is empty or non-numeric"
		}
		if errorOccurred == false {
			if userid == 0 {
				db, err := getDbConnection()
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 276")
				}
				defer db.Close()
				stmt, err := db.Prepare("INSERT INTO login_user (created_gmt, email, password, salt, fname, lname, gps_lat, gps_long, time_zone_offset) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);")
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 282")
				}
				// defer stmt.Close();
				var saveRecord struct {
					createdGmt     uint64
					email          string
					password       string
					salt           string
					fname          string
					lname          string
					gpsLat         float64
					gpsLong        float64
					timeZoneOffset float64
				}
				saveRecord.createdGmt = uint64(time.Now().Unix())
				saveRecord.email = uiFrm.email
				saveRecord.salt, err = generateSalt(128)
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 301")
				}
				saveRecord.password, err = computePasswordHash(uiFrm.password, saveRecord.salt)
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 306")
				}
				saveRecord.fname = postform["fname"][0]
				saveRecord.lname = postform["lname"][0]
				saveRecord.gpsLat = stringToFloat(postform["gps_lat"][0])
				saveRecord.gpsLong = stringToFloat(postform["gps_long"][0])
				saveRecord.timeZoneOffset = stringToFloat(postform["time_zone_offset"][0])
				stmt.Bind(saveRecord.createdGmt, saveRecord.email, saveRecord.password, saveRecord.salt, saveRecord.fname, saveRecord.lname, saveRecord.gpsLat, saveRecord.gpsLong, saveRecord.timeZoneOffset)
				_, _, err = stmt.Exec()
				fmt.Println("THE Exec() CALL WAS DONE")
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 318")
				}
				fmt.Println("THE Exec() CALL WAS SUCCESSFUL")
			} else {
				db, err := getDbConnection()
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 325")
				}
				defer db.Close()
				// check for password change
				var userData struct {
					password string
					salt     string
				}
				res, err := db.Start("SELECT password, salt FROM login_user WHERE id_user=" + strconv.FormatInt(userid, 10) + " ORDER BY id_user;")
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 336")
				}
				// defer res.Close();
				row, err := res.GetRow()
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 342")
				}
				if row == nil {
					fmt.Fprintln(w, "What? row == nil. Where is it? How did we get here?")
					return
				} else {
					userData.password = row.Str(0)
					userData.salt = row.Str(1)
				}
				row, err = res.GetRow() // without this we get "reply is not completely read" error
				stmt, err := db.Prepare("UPDATE login_user SET email=?, password=?, salt=?, fname=?, lname=? gps_lat=?, gps_long=?, time_zone_offset=? WHERE id_user=?;")
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 355")
				}
				// defer stmt.Close();
				if userid == 0 {
					panic("User ID is zero")
				}
				var updateRecord struct {
					email          string
					password       string
					salt           string
					fname          string
					lname          string
					gpsLat         float64
					gpsLong        float64
					timeZoneOffset float64
					userId         uint64
				}
				updateRecord.email = uiFrm.email
				if uiFrm.password == userData.password {
					updateRecord.password = uiFrm.password
					updateRecord.salt = userData.salt
				} else {
					updateRecord.salt, err = generateSalt(128)
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 380")
					}
					updateRecord.password, err = computePasswordHash(uiFrm.password, updateRecord.salt)
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 385")
					}
				}
				updateRecord.fname = uiFrm.fname
				updateRecord.lname = uiFrm.lname
				updateRecord.gpsLat = stringToFloat(uiFrm.gpsLat)
				updateRecord.gpsLong = stringToFloat(uiFrm.gpsLong)
				updateRecord.timeZoneOffset = stringToFloat(uiFrm.timeZoneOffset)
				updateRecord.userId = uint64(userid)
				stmt.Bind(updateRecord.email, updateRecord.password, updateRecord.salt, updateRecord.fname, updateRecord.lname, updateRecord.gpsLat, updateRecord.gpsLong, updateRecord.timeZoneOffset, updateRecord.userId)
				_, _, err = stmt.Exec()
				if err != nil {
					fmt.Fprintln(w, err)
					panic("point 398")
				}
			}
			http.Redirect(w, r, "login", 302)
			return
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for _, errMsg := range errorList {
				// fmt.Fprintln(w, "<li>"+html.EscapeString(errOn)+": "+html.EscapeString(errMsg)+"</li>")
				fmt.Fprintln(w, "<li>"+html.EscapeString(errMsg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, getDoctype())
		fmt.Fprint(w, "<title>")
		if userid == 0 {
			fmt.Fprint(w, "Add")
		} else {
			fmt.Fprint(w, "Edit")
		}
		fmt.Fprint(w, ` User Account</title>
</head><body>
  <section>
    <h1>`)
		if userid == 0 {
			fmt.Fprint(w, "Add")
		} else {
			fmt.Fprint(w, "Edit")
		}
		fmt.Fprint(w, ` User Account</h1>
<form action="`)
		if userid == 0 {
			fmt.Fprint(w, "register")
		} else {
			fmt.Fprint(w, "edit")
		}
		fmt.Fprint(w, `" method="post">
<table border="0" cellpadding="5">
`)
		if userid != 0 {
			fmt.Fprint(w, `
<tr><td> Created </td><td> <input name="user" id="user" value="`+strconv.FormatInt(userid, 10)+`" type="hidden" /> <input name="created" id="created" type="text" value="`+html.EscapeString(uiFrm.created)+`" readonly="readonly" /> </td></tr>
`)
		}
		fmt.Fprint(w, `
<tr><td> Email </td><td> <input name="email" id="email" type="text" value="`+html.EscapeString(uiFrm.email)+`" style="width:400px;" /> </td></tr>
<tr><td> Password </td><td> <input name="password" id="password" type="password" value="`+html.EscapeString(uiFrm.password)+`" style="width:400px;" /> </td></tr>
<tr><td> First Name </td><td> <input name="fname" id="fname" type="text" value="`+html.EscapeString(uiFrm.fname)+`" style="width:400px;" /> </td></tr>
<tr><td> Last Name </td><td> <input name="lname" id="lname" type="text" value="`+html.EscapeString(uiFrm.lname)+`" style="width:400px;" /> </td></tr>
<tr><td> GPS Lat </td><td> <input name="gps_lat" id="gps_lat" type="text" value="`+html.EscapeString(uiFrm.gpsLat)+`" style="width:400px;" /> </td></tr>
<tr><td> GPS Long </td><td> <input name="gps_long" id="gps_long" type="text" value="`+html.EscapeString(uiFrm.gpsLong)+`" style="width:400px;" /> </td></tr>
<tr><td> Time zone offset </td><td> <input name="time_zone_offset" id="time_zone_offset" type="text" value="`+html.EscapeString(uiFrm.timeZoneOffset)+`" style="width:400px;" /> </td></tr>
</table>
<p><input name="submit" id="submit" type="submit" />
</form>
</body></html`)
		// <tr><td> Salt </td><td> <input name="salt" id="salt" type="text" value="`+html.EscapeString(uiFrm.salt)+`" style="width:400px;" /> </td></tr>
	}
}

func listAccounts(w http.ResponseWriter, r *http.Request, operation string) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title>List Of Accounts</title>
</head>
<body>
  <section>
    <h1>List Of Accounts</h1>
    <table border="1">
      <tr><th> User ID </th><th> Created </th><th> Email </th><th> Password </th><th> Salt </th><th> First Name </th><th> Last Name </th><th> Actions </th></tr>
`)
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("point 479")
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("point 485")
	}
	getform := r.Form
	_, show_all := getform["showall"]
	if show_all {
		sql = "SELECT id_user, created_gmt, email, password, salt, fname, lname FROM login_user WHERE 1 ORDER BY id_user;"
	} else {
		sql = "SELECT id_user, created_gmt, email, password, salt, fname, lname FROM login_user WHERE 1 ORDER BY id_user LIMIT 0, 50;"
	}
	res, err := db.Start(sql)
	if err != nil {
		fmt.Fprintln(w, err)
		panic("point 497")
	}
	// defer res.Close();
	for {
		row, err := res.GetRow()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 504")
		}
		if row == nil {
			break
		} else {
			var userData struct {
				userid     uint64
				createdGmt uint64
				email      string
				password   string
				salt       string
				fname      string
				lname      string
			}
			userData.userid = row.Uint64(0)
			userData.createdGmt = row.Uint64(1)
			userData.email = row.Str(2)
			userData.password = row.Str(3)
			userData.salt = row.Str(4)
			userData.fname = row.Str(5)
			userData.lname = row.Str(6)
			fmt.Fprint(w, `<tr><td> `)
			fmt.Fprint(w, strconv.FormatUint(userData.userid, 10))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, strconv.FormatUint(userData.createdGmt, 10))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, html.EscapeString(userData.email))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, html.EscapeString(userData.password))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, html.EscapeString(userData.salt))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, html.EscapeString(userData.fname))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, html.EscapeString(userData.lname))
			fmt.Fprint(w, ` </td><td> `)
			fmt.Fprint(w, `<a href="edit?user=`+strconv.FormatUint(userData.userid, 10)+`">edit</a> &middot; <a href="delete?user=`+strconv.FormatUint(userData.userid, 10)+`">delete</a>
</td></tr>`)
		}
	}
	fmt.Fprint(w, `
    </table>
    <p> <a href="edit?user=0">Add new account</a> </p>
  </section>
</body>
</html>`)
}

func comparePasswordHash(password string, salt string, verifyagainst string) (bool, error) {
	svHash, err := computePasswordHash(password, salt)
	if err != nil {
		return false, err
	}
	bvHash := []byte(svHash)
	bvAgainst := []byte(verifyagainst)
	if !bytes.Equal(bvHash, bvAgainst) {
		return false, err
	}
	return true, err
}

func doLogin(w http.ResponseWriter, r *http.Request, operation string) {
	// var uiFrmemail string
	// var uiFrm.password string
	var errorOccurred bool
	var cookie http.Cookie
	var pwValid bool
	method := r.Method
	errorList := make(map[string]string)
	showform := true
	if method == "GET" {
		showform = true
	}
	var uiFrm uiFields
	if method == "POST" {
		showform = false
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintln(w, err)
			panic("point 583")
		}
		postform := r.Form
		uiFrm.email = postform["epaqlzmhi"][0]
		uiFrm.password = postform["pzsaqwsxrdio"][0]
		if uiFrm.email == "" {
			showform = true
			errorOccurred = true
			errorList["email"] = "Email is empty"
		}
		if uiFrm.password == "" {
			showform = true
			errorOccurred = true
			errorList["password"] = "Password is empty"
		}
		if errorOccurred == false {
			db, err := getDbConnection()
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 602")
			}
			defer db.Close()

			stmt, err := db.Prepare("SELECT id_user, password, salt FROM login_user WHERE (email=?);")
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 609")
			}
			stmt.Bind(uiFrm.email)
			res, err := stmt.Run()
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 615")
			}
			row, err := res.GetRow()
			if err != nil {
				fmt.Fprintln(w, err)
				panic("point 620")
			}
			if row == nil {
				showform = true
				errorOccurred = true
				errorList["email"] = "Sorry, username or password does not match."
			} else {
				userid := row.Uint64(0)
				dbUserPassword := row.Str(1)
				dbUserSalt := row.Str(2)
				for row != nil {
					row, err = res.GetRow()
				}
				pwValid, err = comparePasswordHash(uiFrm.password, dbUserSalt, dbUserPassword)
				if !pwValid {
					showform = true
					errorOccurred = true
					errorList["email"] = "Sorry, username or password does not match."
				} else {
					xcheck, err := generateSalt(20)
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 643")
					}
					stmt, err = db.Prepare("INSERT INTO login_session (id_user, remote_addr, xcheck) VALUES (?, ?, ?);")
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 647")
					}
					remote_addr := r.RemoteAddr
					remote_addr = remote_addr[0:strings.Index(remote_addr, ":")]
					stmt.Bind(userid, remote_addr, xcheck)
					_, _, err = stmt.Exec()
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 655")
					}
					expiretime := time.Now()
					duration, err := time.ParseDuration("86400s")
					if err != nil {
						fmt.Fprintln(w, err)
						panic("point 661")
					}
					expiretime = expiretime.Add(duration)
					cookie.Name = "wgs_user"
					cookie.Value = strconv.FormatUint(userid, 10)
					cookie.Path = "/"
					// host := r.Host
					// host = host[0:strings.Index(host, ":")]
					// cookie.Domain = host
					cookie.Expires = expiretime
					// fmt.Fprintln(w, cookie)
					http.SetCookie(w, &cookie)
					cookie.Name = "wgs_xcheck"
					cookie.Value = xcheck
					cookie.Path = "/"
					// cookie.Domain = host
					cookie.Expires = expiretime
					// fmt.Fprintln(w, cookie)
					http.SetCookie(w, &cookie)
					// expiration := time.Now().Add(365 * 24 * time.Hour)
					// excookie := http.Cookie{Name: "username", Value: "astaxie", Expires: expiration}
					// http.SetCookie(w, &excookie)
					// fmt.Fprintln(w, excookie)
					http.Redirect(w, r, "apps", 302)
					return
				}
			}
		}
	}

	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype())
		fmt.Fprint(w, `<title>Login</title>
</head><body>
  <section>
    <h1>Login</h1>
`)
		if errorOccurred {
			fmt.Fprintln(w, "<h2>Error occurred</h2><ul>")
			for err_on, err_msg := range errorList {
				fmt.Fprintln(w, "<li>"+html.EscapeString(err_on)+": "+html.EscapeString(err_msg)+"</li>")
			}
			fmt.Fprintln(w, "</ul>")
		}
		fmt.Fprint(w, `

<form action="login" method="post">

<table border="0" cellspacing="5">
<tr><td> Email </td><td> <input name="epaqlzmhi" id="epaqlzmhi" type="text" value="`+html.EscapeString(uiFrm.email)+`" style="width:400px;" /> </td></tr>
<tr><td> Password </td><td> <input name="pzsaqwsxrdio" id="pzsaqwsxrdio" type="password" value="`+html.EscapeString(uiFrm.password)+`" style="width:400px;" /> </td></tr>
</table>

<p><input name="submit" id="submit" type="submit" />

</form>

</body></html>`)
	}
}

func deleteAccount(w http.ResponseWriter, r *http.Request, operation string) {
}

func showApps(w http.ResponseWriter) {
	fmt.Fprint(w, getDoctype())
	fmt.Fprint(w, `<title> Apps on this server</title>
</head><body>
  <section>
    <h1>Apps on this server</h1>
    <ul>
        <li><a href="../bookmark/list">Bookmarks</a></li>
        <li><a href="../links/add">Links</a></li>
        <li><a href="../calcron/list">Calcron Chimes</a></li>
        <li><a href="../fitb/listtopics">Fitb</a></li>
        <li><a href="../rand/list">Rand</a></li>
        <li><a href="../umt/umt">UMT</a></li>
    </ul>
  </section>
 </body></html`)
}

func Handler(w http.ResponseWriter, r *http.Request, operation string, userid uint64, userName string) {
	switch {
	case operation == "register":
		editAccount(w, r, operation, true)
	case operation == "edit":
		if userid == 1 {
			editAccount(w, r, operation, false)
		}
	case operation == "list":
		if userid == 1 {
			listAccounts(w, r, operation)
		}
	case operation == "login":
		doLogin(w, r, operation)
	case operation == "delete":
		if userid == 1 {
			deleteAccount(w, r, operation)
		}
	case operation == "register":
		editAccount(w, r, operation, true)
	case operation == "apps":
		showApps(w)
	default:
		fmt.Fprintln(w, "Could not find operation:", operation)
	}
}

func IdentifyLoggedInUser(w http.ResponseWriter, r *http.Request) (uint64, string) {
	var userid uint64
	var err error
	xcheckC := ""
	userid = 0
	for _, cookie := range r.Cookies() {
		// I'm sure this is the wrong way to do it, but I don't know the right way
		if cookie.Name == "wgs_user" {
			userid, err = strconv.ParseUint(cookie.Value, 10, 64)
			if err != nil {
				return 0, ""
			}
		}
		if cookie.Name == "wgs_xcheck" {
			xcheckC = cookie.Value
		}
	}
	db, err := getDbConnection()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("login can't access database point 675")
	}
	// res, err := db.Start("SELECT id_user FROM login_session WHERE (id_user = ?) AND (xcheck = '?');", userid, xcheckC)
	res, err := db.Start("SELECT id_user FROM login_session WHERE (id_user = " + strconv.FormatUint(userid, 10) + ") AND (xcheck = '" + mysql.Escape(db, xcheckC) + "');") // , userid, xcheckC
	if err != nil {
		fmt.Fprintln(w, err)
		panic("query failed point 683")
	}
	row, err := res.GetRow()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("get row failed point 683")
	}
	if row == nil {
		return 0, ""
	}
	for row != nil {
		row, err = res.GetRow()
	}
	stmt, err := db.Prepare("SELECT fname, lname FROM login_user WHERE (id_user=?);")
	if err != nil {
		fmt.Println(err)
		panic("point 811")
	}
	stmt.Bind(userid)
	res, err = stmt.Run()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("point 817")
	}
	row, err = res.GetRow()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("point 822")
	}
	fullName := ""
	if row != nil {
		fname := row.Str(0)
		lname := row.Str(1)
		for row != nil {
			row, err = res.GetRow()
		}
		fullName = fname + " " + lname
	}
	return userid, fullName
}
