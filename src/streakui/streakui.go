package streakui

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"streak"
)

func htmlize(z string) string {
	return html.EscapeString(z)
}

func uintToStr(ii uint64) string {
	return strconv.FormatUint(ii, 10)
}

func intToStr(ii int) string {
	return strconv.FormatInt(int64(ii), 10)
}

func ShowStreakHeadHeader(w http.ResponseWriter, displayInfo map[string]string) {
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

func ShowStreakBodyHeader(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
    <body>
        <div id="header">
            <p>
                <a href="tasklist">List</a>
                <a href="tasklist?edit=1">Edit</a>
                &middot; `+displayInfo["hUserName"]+`
            </p>
        </div>
`)
	// +` &middot; <a href="logout.php?kn=`+displayInfo["kn"]+`" class="btn btn-default">Logout</a><br />
}

func ShowStreakFooter(w http.ResponseWriter, displayInfo map[string]string) {
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

func ShowStreakTaskEditForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo map[string]string) {
	// Taks = `+htmlize(userInput["task"])+`
	fmt.Fprint(w, `
<form action="taskedit" method="post">
<input type="hidden" name="task" value="`+htmlize(userInput["task"])+`" />
<div class="titlesection">
    <h1>Edit Streak Task</h1>
</div>
`)
	showErrorList(w, errorList)
	fmt.Fprint(w, `
<div class="ourform">
    <div class="ourlabel">
        Name:
    </div>
    <div class="ourinput">
        <input class="infield" name="name" id="name" type="text" value="`+htmlize(userInput["name"])+`" />
    </div>
    <div class="ourlabel">
        Description: 
    </div>
    <div class="ourinput">
        <input class="infield" name="description" id="description" type="text" value="`+htmlize(userInput["description"])+`" />
    </div>
    <div class="ourlabel">
        Cycle Days:
    </div>
    <div class="ourinput">
        <input class="infield" name="cycle_days" id="description" type="text" value="`+htmlize(userInput["cycle_days"])+`" />
    </div>
    <div class="oursubmit">
        <input type="submit">
    </div>
</div>
`)
}

func ShowStreakTaskList(w http.ResponseWriter, dbDataList streak.TaskListData, editmode bool) {
	fmt.Fprint(w, `
<body>
  <section>
    <h1>List of Streaks</h1>
`)
	started := false
	count := 0
	for _, task := range dbDataList {
		if count == 0 {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Name </th><th> Cycle Days </th><th> Current Streak Length </th><th> Done Today! </th><th> Edit </th></tr>
`)
			started = true
		}
		count++
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		idStr := uintToStr(task.IdTask)
		streakLenStr := ""
		if task.CurrentStreakLen != 0 {
			streakLenStr = intToStr(task.CurrentStreakLen)
		}

		fmt.Fprint(w, "<tr "+backgroundColor+">")
		var markDone string
		if task.ShowMarkDone {
			markDone = `<a href="markdone?task=` + idStr + `">Mark Done!</a>`
		} else {
			markDone = "&nbsp;"
		}
		fmt.Fprint(w, `<tr `+backgroundColor+`><td align="center"> <span title="`+htmlize(task.Description)+`">`+htmlize(task.Name)+`</span> </td><td align="center"> `+intToStr(task.CycleDays)+`</td><td align="center"> `+streakLenStr+`</td><td align="center"> `+task.TimeRemaining+` </td><td> `+markDone+` </td>
`)
		if editmode {
			fmt.Fprint(w, `<td> <a href="taskedit?task=`+idStr+`">Edit</a> </td><td> <a href="taskdelete?task=`+idStr+`">Delete</a> </td>
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
        <p> <a href="taskedit?task=0">Add Streak</a> </p>
`)
	}
	fmt.Fprint(w, `
  </section>
`)
}

func ShowStreakMarkDoneForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo streak.TaskDisplayData) {
	fmt.Fprint(w, `
<form action="markdone" method="post">
<input type="hidden" name="task" value="`+htmlize(userInput["task"])+`" />
<div class="titlesection">
    <h1>Mark Done!</h1>
</div>
`)
	showErrorList(w, errorList)
	fmt.Fprint(w, `
<div class="ourform">
    <div class="ourlabel">
        Name:
    </div>
    <div class="ourinput">
	`+htmlize(displayInfo.Name)+`
    </div>
    <div class="ourlabel">
        Description:
    </div>
    <div class="ourinput">
	`+htmlize(displayInfo.Description)+`
    </div>
    <div class="ourlabel">
        Cycle Days:
    </div>
    <div class="ourinput">
	`+htmlize(intToStr(displayInfo.CycleDays))+`
    </div>
    <div class="ourlabel">
        Current Streak:
    </div>
    <div class="ourinput">
	`+htmlize(intToStr(displayInfo.CurrentStreakLen))+` days
    </div>
    <div class="ourlabel">
        Time Remaining:
    </div>
    <div class="ourinput">
	`+htmlize(displayInfo.TimeRemaining)+`
    </div>
    <div class="oursubmit">
        <input type="submit">
    </div>
</div>
`)
}

func ShowStreakHistoryForm(w http.ResponseWriter, dbDataList streak.DayHistoryListData) {
	fmt.Fprint(w, `
<body>
  <section>
    <h1>Streak Day History</h1>
`)
	started := false
	count := 0
	for _, dayEntry := range dbDataList {
		if count == 0 {
			fmt.Fprint(w, `<form><table border="0" > <tr> <th> Actual Time </th><th> Day Num </th><th> Consecutive </th><th> Gap </th></tr>
`)
			started = true
		}
		count++
		backgroundColor := " style=\"background-color: #FFFFFF;\""
		if (count & 1) == 1 {
			backgroundColor = " style=\"background-color: #E8F0E8;\""
		}
		fmt.Fprint(w, "<tr "+backgroundColor+">")
		fmt.Fprint(w, "<td> "+uintToStr(dayEntry.ActualTimeGmt)+" </td><td> "+uintToStr(dayEntry.DayNum)+" </td><td> "+intToStr(dayEntry.Consecutive)+" </td><td> "+intToStr(dayEntry.Gap)+"</td>"+`
`)
	}
	if started {
		fmt.Fprint(w, "</table>")
	}
}
