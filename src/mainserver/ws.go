package main

import (
	"accessdb"
	"blind"
	"bookmark"
	"calcron"
	"fitb"
	"fmt"
	"geo"
	"hnfire"
	"langtext"
	"links"
	"login"
	"net/http"
	"static"
	"strconv"
	"streaklogic"
	"strings"
	"time"
	"umt"
	"youtube"
"zorpcore"
)

func dumpRequestInfoToBrowser(w http.ResponseWriter, r *http.Request, parseform bool) {
	if parseform {
		err := r.ParseForm()
		fmt.Fprintln(w, "called ParseForm!")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
	}
	method := r.Method
	fmt.Fprintln(w, "method", method)
	theurl := r.URL
	fmt.Fprintln(w, "theurl", theurl)
	proto := r.Proto
	fmt.Fprintln(w, "proto", proto)
	protomajor := r.ProtoMajor
	fmt.Fprintln(w, "protomajor", protomajor)
	protominor := r.ProtoMinor
	fmt.Fprintln(w, "protominor", protominor)
	header := r.Header
	fmt.Fprintln(w, "header", header)
	body := r.Body
	fmt.Fprintln(w, "body", body)
	contentlength := r.ContentLength
	fmt.Fprintln(w, "contentlength", contentlength)
	transferencoding := r.TransferEncoding
	fmt.Fprintln(w, "transferencoding", transferencoding)
	closed := r.Close
	fmt.Fprintln(w, "closed", closed)
	host := r.Host
	fmt.Fprintln(w, "host", host)
	form := r.Form
	fmt.Fprintln(w, "form", form)
	multipart := r.MultipartForm
	fmt.Fprintln(w, "multipart", multipart)
	trailer := r.Trailer
	fmt.Fprintln(w, "trailer", trailer)
	remoteaddr := r.RemoteAddr
	fmt.Fprintln(w, "remoteaddr", remoteaddr)
	requesturi := r.RequestURI
	fmt.Fprintln(w, "reuesturi", requesturi)
	tls := r.TLS
	fmt.Fprintln(w, "tls", tls)
	fmt.Fprintln(w, "Cookies:")
	for _, cookie := range r.Cookies() {
		fmt.Fprintln(w, cookie.Name)
		fmt.Fprintln(w, cookie.Value)
	}
	return
}

func dumpRequestInfoToConsole(w http.ResponseWriter, r *http.Request, parseform bool) {
	fmt.Println("-------- dump request to console")
	if parseform {
		err := r.ParseForm()
		fmt.Println("called ParseForm!")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	method := r.Method
	fmt.Println("method", method)
	theurl := r.URL
	fmt.Println("theurl", theurl)
	proto := r.Proto
	fmt.Println("proto", proto)
	protomajor := r.ProtoMajor
	fmt.Println("protomajor", protomajor)
	protominor := r.ProtoMinor
	fmt.Println("protominor", protominor)
	header := r.Header
	fmt.Println("header", header)
	body := r.Body
	fmt.Println("body", body)
	contentlength := r.ContentLength
	fmt.Println("contentlength", contentlength)
	transferencoding := r.TransferEncoding
	fmt.Println("transferencoding", transferencoding)
	closed := r.Close
	fmt.Println("closed", closed)
	host := r.Host
	fmt.Println("host", host)
	form := r.Form
	fmt.Println("form", form)
	multipart := r.MultipartForm
	fmt.Println("multipart", multipart)
	trailer := r.Trailer
	fmt.Println("trailer", trailer)
	remoteaddr := r.RemoteAddr
	fmt.Println("remoteaddr", remoteaddr)
	requesturi := r.RequestURI
	fmt.Println("reuesturi", requesturi)
	tls := r.TLS
	fmt.Println("tls", tls)
	fmt.Println("Cookies:")
	for _, cookie := range r.Cookies() {
		fmt.Println(cookie.Name)
		fmt.Println(cookie.Value)
	}
	fmt.Println("-------- end")
	return
}

func findTopHost(host string) (string, string, int) {
	fmt.Println("ws-debug", "findTopHost", "host", host)
	portNum := 80
	// chop off port number, if there is one
	colon := strings.Index(host, ":")
	if colon > 0 {
		lPort64, err := strconv.ParseInt(host[colon+1:], 10, 0)
		if err == nil {
			portNum = int(lPort64)
		}
		host = host[:colon]
	}
	// chops off the "top" part of the host name so the subdomains can be freely wildcarded
	pieces := strings.Split(host, ".")
	lx := len(pieces)
	if lx == 2 {
		return "", pieces[lx-2] + "." + pieces[lx-1], portNum
	}
	return pieces[lx-3], pieces[lx-2] + "." + pieces[lx-1], portNum
}

func testFindTopHost() {
	fmt.Println("Running findTopHost tests")
	aa, bb, cc := findTopHost("waynebrain.tv")
	if aa != "" {
		fmt.Println("Test 148 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 151 failed")
	}
	if cc != 80 {
		fmt.Println("Test 154 failed")
	}
	aa, bb, cc = findTopHost("waynebrain.tv:4000")
	if aa != "" {
		fmt.Println("Test 158 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 161 failed")
	}
	if cc != 4000 {
		fmt.Println("Test 164 failed")
	}
	aa, bb, cc = findTopHost("secure.waynebrain.tv")
	if aa != "secure" {
		fmt.Println("Test 168 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 171 failed")
	}
	if cc != 80 {
		fmt.Println("Test 174 failed")
	}
	aa, bb, cc = findTopHost("secure.waynebrain.tv:4000")
	if aa != "secure" {
		fmt.Println("Test 178 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 181 failed")
	}
	if cc != 4000 {
		fmt.Println("Test 184 failed")
	}
	aa, bb, cc = findTopHost("www.secure.waynebrain.tv:4000")
	if aa != "secure" {
		fmt.Println("Test 188 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 191 failed")
	}
	if cc != 4000 {
		fmt.Println("Test 194 failed")
	}
	aa, bb, cc = findTopHost("hello.world.www.secure.waynebrain.tv:4000")
	if aa != "secure" {
		fmt.Println("Test 198 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 201 failed")
	}
	if cc != 4000 {
		fmt.Println("Test 204 failed")
	}
	aa, bb, cc = findTopHost("hello.world.www.secure.waynebrain.tv:notnumeric")
	if aa != "secure" {
		fmt.Println("Test 208 failed")
	}
	if bb != "waynebrain.tv" {
		fmt.Println("Test 211 failed")
	}
	if cc != 80 {
		fmt.Println("Test 214 failed")
	}
	fmt.Println("End of findTopHost tests")
}

func operationIsApp(appName string, operation string) bool {
	withSlashes := "/" + appName + "/"
	lws := len(withSlashes)
	if len(operation) >= lws {
		if operation[:lws] == withSlashes {
			return true
		}
	}
	return false
}

func operationForApp(appName string, operation string) string {
	return operation[(len(appName) + 2):]
}

type wayneGoServerHandler struct {
	name string
}

func (self wayneGoServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// dumpRequestInfoToBrowser(w, r, false)
	// dumpRequestInfoToConsole(w, r, true)
	header := w.Header()
	header.Set("Server", "Wayneserver (Linux/CentOS)")
	host := r.Host
	subdomain := ""
	topHost := ""
	// have no idea how it's possible for host to be empty
	if host != "" {
		// testFindTopHost()
		subdomain, topHost, _ = findTopHost(host)
	}
	// By convention, we require all "secure" subdomain requests to have TLS turned on
	secure := r.TLS != nil
	if subdomain == "secure" {
		if !secure {
			return
		}
	}
	// userid, userName := login.IdentifyLoggedInUser(w, r)
	db := accessdb.GetDbConnection()
	defer db.Close()
	userInfo := login.IdentifyLoggedInUser(db, w, r)
	userid := userInfo.UserId
	userName := userInfo.UserName
	operation := r.URL.Path

	// static.StaticDirHandler(w, r, "secure", "/index.txt")
	// return

	fmt.Println("userid", userid, "operation", operation)
	if strings.Contains(operation, "..") {
		w.WriteHeader(404)
		return
	}
	//
	// First section
	// Apps referenced by URL
	if operationIsApp("login", operation) {
		login.Handler(w, r, operationForApp("login", operation), userid, userName, secure)
		return
	}
	if operationIsApp("links", operation) {
		links.Handler(w, r, topHost, operationForApp("links", operation), userid, userName)
		return
	}
	if operationIsApp("umt", operation) {
		umt.Handler(w, r, operationForApp("links", operation), userid, userName)
		return
	}
	if operationIsApp("georand", operation) {
		geo.Handler(w, r, operationForApp("georand", operation), userid, userName)
		return
	}
	if operationIsApp("calcron", operation) {
		calcron.Handler(w, r, operationForApp("calcron", operation), userid, userName)
		return
	}
	if operationIsApp("chimes", operation) {
		calcron.Handler(w, r, operationForApp("chimes", operation), userid, userName)
		return
	}
	if operationIsApp("blind", operation) {
		blind.Handler(w, r, operationForApp("blind", operation), userid, userName)
		return
	}
	if operationIsApp("fitb", operation) {
		fitb.Handler(w, r, operationForApp("fitb", operation), userid, userName)
		return
	}
	if operationIsApp("bookmark", operation) {
		bookmark.Handler(w, r, operationForApp("bookmark", operation), userid, userName)
		return
	}
	if operationIsApp("nback", operation) {
		static.StaticDirHandler(w, r, "nback", operation)
		return
	}
	if operationIsApp("nato", operation) {
		static.StaticDirHandler(w, r, "nato", operation)
		return
	}
	if operationIsApp("stopwatch", operation) {
		static.StaticDirHandler(w, r, "stopwatch", operation)
		return
	}
	if operationIsApp("bpm", operation) {
		static.StaticDirHandler(w, r, "bpm", operation)
		return
	}
	if operationIsApp("focusoninfinity", operation) {
		static.StaticDirHandler(w, r, "focusoninfinity", operation)
		return
	}
	if operationIsApp("datascience", operation) {
		static.StaticDirHandler(w, r, "datascience", operation)
		return
	}
	if operationIsApp("youtube", operation) {
		youtube.Handler(w, r, operationForApp("youtube", operation), userid, userName)
		return
	}
	if operationIsApp("hnfire", operation) {
		hnfire.Handler(w, r, operationForApp("hnfire", operation), userid, userName)
		return
	}
	if operationIsApp("streak", operation) {
		streaklogic.Handler(w, r, operationForApp("streak", operation), userInfo)
		return
	}
	if operationIsApp("waynetype", operation) {
		langtext.Handler(w, r, operationForApp("waynetype", operation), userInfo)
		return
	}
	if operationIsApp("zorp", operation) {
		zorpcore.Handler(w, r, operationForApp("zorp", operation), userInfo)
		return
	}
	if operationIsApp("files", operation) {
		static.StaticDirHandler(w, r, "files", operation)
		return
	}
	//
	// Second section: Apps referenced by domain name
	if topHost == "ultimatemusictoy.org" {
		umt.Handler(w, r, operation[1:], userid, userName)
		return
	}
	if topHost == "ultimatemusictoy.com" {
		umt.Handler(w, r, operation[1:], userid, userName)
		return
	}
	if topHost == "camelcasejslint.com" {
		if operation == "/" {
			operation = "/jslint.html"
		}
		static.StaticDirHandler(w, r, "jslint", operation)
		return
	}
	if topHost == "blindmovieclub.org" {
		static.StaticDirHandler(w, r, "blind", operation)
		return
	}
	if topHost == "fallingblocks.xyz" {
		static.StaticDirHandler(w, r, "blocks", operation)
		return
	}
	if topHost == "learndeeplearning.study" {
		static.StaticDirHandler(w, r, "learndeeplearning", operation)
		return
	}
	if topHost == "waynebrain.tv" {
		links.Handler(w, r, topHost, operation[1:], userid, userName)
		return
	}
	if topHost == "musicfortoday.tv" {
		links.Handler(w, r, topHost, operation[1:], userid, userName)
		return
	}
	//
	// Third section
	// Everything else, we let fall through to an "App" that serves static HTML pages
	static.StaticHandler(w, r, operation)
}

func startRegularServer() {
	var regularServerHandler wayneGoServerHandler
	regularServerHandler.name = "Regular Server"

	// Change port here for testing
	regularSrv := &http.Server{
		Addr:           ":80",
		Handler:        regularServerHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting regular server")
	err := regularSrv.ListenAndServe()
	fmt.Println(err)

	// log.Fatal
}

func startSecureServer() {
	var secureServerHandler wayneGoServerHandler
	secureServerHandler.name = "Secure Server"

	secureSrv := &http.Server{
		Addr:           ":443",
		Handler:        secureServerHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting secure server")
	err := secureSrv.ListenAndServeTLS("/home/ec2-user/wayneserver/certificate-677199.crt", "/home/ec2-user/wayneserver/server-677199.key")
	if err == nil {
		fmt.Println("Secure server started")
	} else {
		fmt.Println(err)
	}
}

func main() {
	// This commented-out code is the old code for when our server was "one" server (http only). Now it is two! http and https!
	// http.HandleFunc("/", handler)
	// err := http.ListenAndServeTLS(":443", "/home/ec2-user/wayneserver/certificate-677199.crt", "/home/ec2-user/wayneserver/server-677199.key", handle)
	// fmt.Println(err)
	// err := http.ListenAndServe(":4000", handler)
	// fmt.Println(err)

	// This code creates two servers, one for http (port 80) and one for https (port 443)
	// Both work by passing a wayneGoServer as "Handler", which results in the same wayneGoServer
	// ServeHTTP() function getting called for both. Inside the ServeHTTP() function, it checks
	// to see if r.TLS is not nil for services that are required to be secure

	// ListenAndServe is blocking so we need to create our own goroutines
	go startSecureServer()
	// we let the regular server run on the current goroutine so the process doesn't exit
	startRegularServer()
}
