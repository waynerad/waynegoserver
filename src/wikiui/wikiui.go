package wikiui

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"wikitypes"
)

func htmlize(z string) string {
	return html.EscapeString(z)
}

func uintToStr(ii uint64) string {
	return strconv.FormatUint(ii, 10)
}

func ShowWikiHeadHeader(w http.ResponseWriter, displayInfo map[string]string) {
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
    width: 300px;
}

.intxtarea {
    font-size: 1.1em;
}

@media (min-width: 445px) {
  .ourform {
    display: grid;
    grid-template-columns: 150px 305px;
    grid-column-gap: 10px;
    grid-row-gap: 5px;
  }
  .ourlabel {
    justify-self: end;
  }
  .oursubmit {
    grid-column: 1 / 3;
    justify-self: center;
  }
}

@media (max-width: 444px) {
  .ourform {
    display: grid;
    grid-template-columns: 305px;
    grid-column-gap: 10px;
    grid-row-gap: 5px;
  }
  .ourlabel {
    justify-self: start;
    background-color: #FFEFE0;
  }
  .oursubmit {
    justify-self: center;
  }
}

        </style>
`)
}

func ShowWikiBodyHeader(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
    <body>
        <div id="header">
            <p>
                <a href="wikipagelist">List</a>
                &middot; <a href="wikipagelist?edit=1">Edit</a>
                &middot; `+displayInfo["hUserName"]+`
            </p>
        </div>
`)
}

func ShowWikiFooter(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
        <div id="footer">
            <p>&nbsp;</p>
        </div>
    </body>
</html>`)
}

func showErrorList(w http.ResponseWriter, errorList map[string]string) {
	if len(errorList) == 0 {
		return
	}
	fmt.Fprintln(w, `
<div class="errorsection">
    <div class="errtitle">
        <h1><font color="red">Error occurred</font></h1>
    </div>
    <div class="errlist">
`)
	for _, errMsg := range errorList {
		fmt.Fprintln(w, `
        <div class="errrow">
            <font color="red">* `+htmlize(errMsg)+`</font>
        </div>
`)
	}
	fmt.Fprintln(w, `
    </div>
</div>
`)
}

func ShowWikiPageEditForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo map[string]string) {
	fmt.Fprint(w, `
<form action="wikipageedit" method="post">
<input type="hidden" name="page" value="`+htmlize(userInput["page"])+`" />
<div class="titlesection">
    <h1>Edit Wiki Page</h1>
</div>
`)
	showErrorList(w, errorList)
	fmt.Fprint(w, `
<div class="ourform">
    <div class="ourlabel">
        Title:
    </div>
    <div class="ourinput">
        <input class="infield" name="title" id="title" type="text" value="`+htmlize(userInput["title"])+`" />
    </div>
    <div class="ourlabel">
        Content: 
    </div>
    <div class="ourinput">
        <textarea class="intxtarea" name="content" id="content" cols="106" rows="24">`+htmlize(userInput["content"])+`</textarea>
    </div>
    <div class="oursubmit">
        <input type="submit">
    </div>
</div>
`)
}

func ShowWikiPageList(w http.ResponseWriter, dbDataList wikitypes.WikiPageListData, editmode bool) {
	fmt.Fprint(w, `
<body>
  <section>
    <h1>List of Wiki Pages</h1>
`)
	started := false
	count := 0
	for _, wiki := range dbDataList {
		if count == 0 {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Title </th>`)
			if editmode {
				fmt.Fprint(w, `<th> Edit </th><th> Delete </th>`)
			}
			fmt.Fprint(w, `</tr>
`)
			started = true
		}
		count++
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		idStr := uintToStr(wiki.IdPage)
		fmt.Fprint(w, `<tr `+backgroundColor+`><td align="center"> <a href="wikipageshow?page=`+idStr+`"><span title="`+htmlize(wiki.Title)+`">`+htmlize(wiki.Title)+`</span></a> </td>
`)
		if editmode {
			fmt.Fprint(w, `<td> <a href="wikipageedit?page=`+idStr+`">Edit</a> </td><td> <a href="wikipagedelete?page=`+idStr+`">Delete</a> </td>
`)
		}
		fmt.Fprint(w, `</tr>
`)
	}
	if started {
		fmt.Fprint(w, `</table>
`)
	}
	if editmode {
		fmt.Fprint(w, `
        <p> <a href="wikipageedit?page=0">Add Wiki Page</a> </p>
`)
	}
	fmt.Fprint(w, `
  </section>
`)
}

func ShowWikiPageShowForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo map[string]string) {
	fmt.Fprint(w, `
<form action="wikipageshow" method="post">
<input type="hidden" name="page" value="`+htmlize(userInput["page"])+`" />
<div class="titlesection">
    <p><a href="wikipageedit?page=`+userInput["page"]+`">Edit this page</a></p>
    <h1>`+htmlize(displayInfo["title"])+`</h1>
</div>
`)
	showErrorList(w, errorList)
	fmt.Fprint(w, `
    `+displayInfo["content"]+`
`)
}
