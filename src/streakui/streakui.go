package streakui

import (
	"fmt"
	"html"
	"net/http"
)

func ShowHeadHeader(w http.ResponseWriter, displayInfo map[string]string) {
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

        </style>
`)
}

func ShowBodyHeader(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
    <body>

Hello `+displayInfo["hUserName"]+`!<br />
<a href="logout.php?kn=`+displayInfo["kn"]+`" class="btn btn-default">Logout</a><br /><br />
`)
}

func ShowFooter(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
        <div>
            [Footer goes here]
        </div>
    </body>
</html>`)
}

func htmlize(z string) string {
	return html.EscapeString(z)
}

func ShowTaskEditForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo map[string]string) {
	fmt.Fprint(w, `
<form action="taskedit" method="post">
Taks = `+htmlize(userInput["task"])+`
<input type="hidden" name="task" value="`+htmlize(userInput["task"])+`" />
<div class="container">
    <div class="row">
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <h1>Edit Streak Task</h1>
        </div>
    </div>
`)
	if len(errorList) > 0 {
		fmt.Fprintln(w, `
    <div class="row">
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <h1><font color="red">Error occurred</font></h1>
        </div>
    </div>
`)
		for _, errMsg := range errorList {
			fmt.Fprintln(w, `
    <div class="row">
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <font color="red">* `+htmlize(errMsg)+`</font>
        </div>
    </div>
`)
		}
	}
	fmt.Fprint(w, `
    <div class="row">
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            Title:
        </div>
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            <input name="name" id="name" type="text" value="`+htmlize(userInput["title"])+`" />
        </div>
    </div>
    <div class="row">
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            Description: 
        </div>
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            <input name="description" id="description" type="text" value="`+htmlize(userInput["description"])+`" />
        </div>
    </div>
    <div class="row">
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            Cycle Days:
        </div>
        <div class="col-lg-6 col-md-6 col-sm-6 col-xs-12">
            <input name="cycle_days" id="description" type="text" value="`+htmlize(userInput["cycle_days"])+`" />
        </div>
    </div>
    <div class="row">
        <div class="col-lg-12 col-md-12 col-sm-12 col-xs-12">
            <input type="submit">
        </div>
    </div>
</div>
`)
}
