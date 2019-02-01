package login

import (
	"accessdb"
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

type UserInformationRecord struct {
	UserId   uint64
	UserName string
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

</style>

`
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
		db := accessdb.GetDbConnection()
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
				db := accessdb.GetDbConnection()
				defer db.Close()
				stmt, err := db.Prepare("INSERT INTO login_user (created_gmt, email, password, salt, fname, lname, gps_lat, gps_long, time_zone_offset) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);")
				if err != nil {
					fmt.Println(err)
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
					fmt.Println(err)
					fmt.Fprintln(w, err)
					panic("point 301")
				}
				saveRecord.password, err = computePasswordHash(uiFrm.password, saveRecord.salt)
				if err != nil {
					fmt.Println(err)
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
					fmt.Println(err)
					fmt.Fprintln(w, err)
					panic("point 318")
				}
				fmt.Println("THE Exec() CALL WAS SUCCESSFUL")
			} else {
				db := accessdb.GetDbConnection()
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
				stmt, err := db.Prepare("UPDATE login_user SET email=?, password=?, salt=?, fname=?, lname=?, gps_lat=?, gps_long=?, time_zone_offset=? WHERE id_user=?;")
				if err != nil {
					fmt.Fprintln(w, err)
					fmt.Println(w, err)
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
						fmt.Println(err)
						fmt.Fprintln(w, err)
						panic("point 380")
					}
					updateRecord.password, err = computePasswordHash(uiFrm.password, updateRecord.salt)
					if err != nil {
						fmt.Println(err)
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
					fmt.Println(err)
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
		fmt.Fprint(w, getDoctype()+getStyle())
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
</body></html>`)
		// <tr><td> Salt </td><td> <input name="salt" id="salt" type="text" value="`+html.EscapeString(uiFrm.salt)+`" style="width:400px;" /> </td></tr>
	}
}

func listAccounts(w http.ResponseWriter, r *http.Request, operation string) {
	var sql string
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title>List Of Accounts</title>
</head>
<body>
  <section>
    <h1>List Of Accounts</h1>
    <table border="1">
      <tr><th> User ID </th><th> Created </th><th> Email </th><th> Password </th><th> Salt </th><th> First Name </th><th> Last Name </th><th> Actions </th></tr>
`)
	db := accessdb.GetDbConnection()
	defer db.Close()
	err := r.ParseForm()
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
			db := accessdb.GetDbConnection()
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
					// duration, err := time.ParseDuration("86400s") // 1 day
					// duration, err := time.ParseDuration("604800s") // 1 week
					duration, err := time.ParseDuration("2629744s") // 1 month
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
					http.Redirect(w, r, "programs", 302)
					return
				}
			}
		}
	}
	if showform {
		header := w.Header()
		header.Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, getDoctype()+getStyle())
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

func showPrograms(w http.ResponseWriter) {
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title> Programs on this server</title>
</head><body>
  <section>
    <h1>Programs on this server</h1>
    <ul>
        <li><a href="../bookmark/list">Bookmarks</a></li>
        <li><a href="../links/add">Links</a></li>
        <li><a href="../chimes/list">Chimes</a></li>
        <li><a href="../fitb/listtopics">FITB</a></li>
        <li><a href="../streak/tasklist">Streaks</a></li>
        <li><a href="../waynetype/waynetype">WayneType</a></li>
        <li><a href="../georand/list">Georand</a></li>
        <li><a href="../umt/umt">UMT</a></li>
        <li><a href="../nback/nback.html">NBACK</a></li>
        <li><a href="../stopwatch/index.html">Stopwatch</a></li>
        <li><a href="../nato/nato.html">NATO Alphabet</a></li>
        <li><a href="../bpm/bpm.html">BPM</a></li>
        <li><a href="../focusoninfinity/focusoninfinity.html">Focus On Infinity</a></li>
    </ul>
  </section>
 </body></html`)
}

func showLogoutAll(w http.ResponseWriter) {
	fmt.Fprint(w, getDoctype()+getStyle())
	db := accessdb.GetDbConnection()
	stmt, err := db.Prepare("TRUNCATE login_session;")
	if err != nil {
		fmt.Println(err)
	}
	_, err = stmt.Run()
	if err != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprint(w, `<title> Logout All </title>
</head><body>
  <section>
    <h1>Logout All</h1>
  </section>
 </body></html`)
}

func showTLSRequired(w http.ResponseWriter) {
	fmt.Fprint(w, getDoctype()+getStyle())
	fmt.Fprint(w, `<title> HTTPS required </title>
</head><body>
  <section>
    <h1>HTTPS required</h1>
    <p>HTTPS (secure http) is required for user account operations</p>
  </section>
 </body></html`)
}

func Handler(w http.ResponseWriter, r *http.Request, operation string, userid uint64, userName string, secure bool) {
	if !secure {
		showTLSRequired(w)
		return
	}
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
	case operation == "programs":
		showPrograms(w)
	case operation == "logoutall":
		showLogoutAll(w)
	default:
		fmt.Fprintln(w, "Could not find operation:", operation)
	}
}

func IdentifyLoggedInUser(db mysql.Conn, w http.ResponseWriter, r *http.Request) *UserInformationRecord {
	var userid uint64
	var err error
	var rv UserInformationRecord
	xcheckC := ""
	userid = 0
	for _, cookie := range r.Cookies() {
		// I'm sure this is the wrong way to do it, but I don't know the right way
		if cookie.Name == "wgs_user" {
			userid, err = strconv.ParseUint(cookie.Value, 10, 64)
			if err != nil {
				rv.UserId = 0
				rv.UserName = ""
				return &rv
			}
		}
		if cookie.Name == "wgs_xcheck" {
			xcheckC = cookie.Value
		}
	}
	// db := accessdb.GetDbConnection()
	// res, err := db.Start("SELECT id_user FROM login_session WHERE (id_user = ?) AND (xcheck = '?');", userid, xcheckC)
	// res, err := db.Start("SELECT id_user FROM login_session WHERE (id_user = " + strconv.FormatUint(userid, 10) + ") AND (xcheck = '" + mysql.Escape(db, xcheckC) + "');") // , userid, xcheckC

	stmt, err := db.Prepare("SELECT id_user FROM login_session WHERE (id_user = ?) AND (xcheck = ?);")
	if err != nil {
		fmt.Println(err)
		panic("prepare failed, login point 799")
	}
	stmt.Bind(userid, xcheckC)
	res, err := stmt.Run()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("run failed, login point 805")
	}
	row, err := res.GetRow()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("getrow failed, login point 810")
	}
	if row == nil {
		rv.UserId = 0
		rv.UserName = ""
		return &rv
	}
	for row != nil {
		row, err = res.GetRow()
	}
	// Verified session, now let's get the user's name
	stmt, err = db.Prepare("SELECT fname, lname FROM login_user WHERE (id_user=?);")
	if err != nil {
		fmt.Println(err)
		panic("prepare failed, login point 827")
	}
	stmt.Bind(userid)
	res, err = stmt.Run()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("run failed, login point 833")
	}
	row, err = res.GetRow()
	if err != nil {
		fmt.Fprintln(w, err)
		panic("getrow failed, login point 838")
	}
	fullname := ""
	if row != nil {
		fname := row.Str(0)
		lname := row.Str(1)
		for row != nil {
			row, err = res.GetRow()
		}
		fullname = fname + " " + lname
	}

	rv.UserId = userid
	rv.UserName = fullname
	return &rv
}
