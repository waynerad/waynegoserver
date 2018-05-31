package langtextui

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
#langtocvt {
    width: 800px;
}
.utyp {
    color: #008080;
}
.gives {
}

h1 {
    color: #550000;
}

#langtocvt {
    font-size: 1.1em;
}

@media (max-width: 3px) {
  body {
    background-color: #FF00FF;
  }
}

        </style>
        <script>

/*jslint browser: true, passfail: true */

function advanceOnReturn(ev) {
    "use strict";
    if (ev.keyCode === 13) {
        ev.stopPropagation();
        window.document.forms.frmLang.submit();
        return false;
    }
    return false;
}

function showcodes() {
    "use strict";
    document.getElementById("thetable").style.display = "block";
    document.getElementById("showcodelnk").style.display = "none";
}

function flipToMultiLine() {
    "use strict";
    var content;
    content = document.getElementById("langtocvt").value;
    document.getElementById("txtplace").innerHTML = '<input id="multiline" name="multiline" value="1" type="hidden" /><textarea name="langtocvt" id="langtocvt" type="text" cols="120" rows="20" /></textarea><br /><input name="submit" id="submit" type="submit" />';
    document.getElementById("langtocvt").value = content;
}

        </script>
`)
}

func ShowBodyHeader(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
    <body onload="document.getElementById('langtocvt').focus();">
        <div id="header">
`)
	if displayInfo["hUserName"] != "" {
		fmt.Fprint(w, "Hello "+displayInfo["hUserName"]+"!<br />")
	} else {
		fmt.Fprint(w, "&nbsp;")
	}
	fmt.Fprint(w, `
        </div>
`)
	// <a href="logout.php?kn=`+displayInfo["kn"]+`" class="btn btn-default">Logout</a><br /><br />
}

func ShowFooter(w http.ResponseWriter, displayInfo map[string]string) {
	fmt.Fprint(w, `
        <div id="footer">
            <p>&nbsp;</p>
        </div>
    </body>
</html>`)
}

func htmlize(z string) string {
	return html.EscapeString(z)
}

func ShowLangTextForm(w http.ResponseWriter, errorList map[string]string, userInput map[string]string, displayInfo map[string]string, theTable string) {
	singleLine := true
	_, ok := userInput["multiline"]
	if ok {
		singleLine = false
	}
	_, dvorak := userInput["dvorak"]
	if dvorak {
		if userInput["dvorak"] == "0" {
			dvorak = false
		}
	}
	fmt.Fprint(w, `
        <form name="frmLang" id="frmLang" action="waynetype" method="post" noonsubmit="return false;">
            <input name="dvorak" id="dvorak" type="hidden" value="`+htmlize(userInput["dvorak"])+`" /><br />
            <h1>WayneType`)
	if dvorak {
		fmt.Fprint(w, " - Dvorak Version")
	}
	fmt.Fprint(w, `</h1>
`)
	if len(errorList) > 1 {
		fmt.Fprintln(w, `
            <h1><font color="red">Error occurred</font></h1>
`)
		for _, errMsg := range errorList {
			fmt.Fprintln(w, `
                <font color="red">* `+htmlize(errMsg)+`</font>
`)
		}
	}
	fmt.Fprint(w, `
            Text:<br /> 
            <div id="txtplace">
`)
	if singleLine {
		fmt.Fprint(w, `<table><tr><td align="right">
                <input name="langtocvt" id="langtocvt" type="text" onkeypress="advanceOnReturn(event);" value="`+htmlize(userInput["langtocvt"])+`" /><br />
                <a id="mlflip" href="javascript:flipToMultiLine();">Multi-line</a>
</td></tr></table>
`)
	} else {
		fmt.Fprint(w, `
                <input id="multiline" name="multiline" value="1" type="hidden" />
                <textarea name="langtocvt" id="langtocvt" type="text" cols="120" rows="20" />`+htmlize(userInput["langtocvt"])+`</textarea><br />
                <input name="submit" id="submit" type="submit" />
`)
	}
	fmt.Fprint(w, `
            </div>
            <br />
            <a id="showcodelnk" href="javascript:showcodes();">Show codes</a>
            <div id="thetable" style="display: none;">

<table>

<tr><td align="right"> <span class="gives">&#224;</span> </td><td> <span class="utyp">`+"``"+`a &nbsp; ,,a</span> </td></tr>
<tr><td align="right"> <span class="gives">&#225;</span> </td><td> <span class="utyp">''a</span> </td></tr>
<tr><td align="right"> <span class="gives">&#226;</span> </td><td> <span class="utyp">^^a &nbsp; &lt;&lt;a</span> </td></tr>
<tr><td align="right"> <span class="gives">&#227;</span> </td><td> <span class="utyp">~~a</span> </td></tr>
<tr><td align="right"> <span class="gives">&#228;</span> </td><td> <span class="utyp">""a</span> </td></tr>
<tr><td align="right"> <span class="gives">&#229;</span> </td><td> <span class="utyp">oOoa</span> </td></tr>
<tr><td align="right"> <span class="gives">&#230;</span> </td><td> <span class="utyp">aaee</span> </td></tr>
<tr><td align="right"> <span class="gives">&#231;</span> </td><td> <span class="utyp">c//</span> </td></tr>
<tr><td align="right"> <span class="gives">&#240;</span> </td><td> <span class="utyp">--d--</span> </td></tr>
<tr><td align="right"> <span class="gives">&#232;</span> </td><td> <span class="utyp">`+"``"+`e &nbsp; ,,e</span> </td></tr>
<tr><td align="right"> <span class="gives">&#233;</span> </td><td> <span class="utyp">''e</span> </td></tr>
<tr><td align="right"> <span class="gives">&#234;</span> </td><td> <span class="utyp">^^e &nbsp; &lt;&lt;e</span> </td></tr>
<tr><td align="right"> <span class="gives">&#235;</span> </td><td> <span class="utyp">""e</span> </td></tr>
<tr><td align="right"> <span class="gives">&#236;</span> </td><td> <span class="utyp">`+"``"+`i &nbsp; ,,i</span> </td></tr>
<tr><td align="right"> <span class="gives">&#237;</span> </td><td> <span class="utyp">''i</span> </td></tr>
<tr><td align="right"> <span class="gives">&#238;</span> </td><td> <span class="utyp">^^i &nbsp; &lt;&lt;i</span> </td></tr>
<tr><td align="right"> <span class="gives">&#239;</span> </td><td> <span class="utyp">""i</span> </td></tr>
<tr><td align="right"> <span class="gives">&#241;</span> </td><td> <span class="utyp">~~n</span> </td></tr>
<tr><td align="right"> <span class="gives">&#242;</span> </td><td> <span class="utyp">`+"``"+`o &nbsp; ,,o</span> </td></tr>
<tr><td align="right"> <span class="gives">&#243;</span> </td><td> <span class="utyp">''o</span> </td></tr>
<tr><td align="right"> <span class="gives">&#244;</span> </td><td> <span class="utyp">^^o &nbsp; &lt;&lt;o</span> </td></tr>
<tr><td align="right"> <span class="gives">&#245;</span> </td><td> <span class="utyp">~~o</span> </td></tr>
<tr><td align="right"> <span class="gives">&#246;</span> </td><td> <span class="utyp">""o</span> </td></tr>
<tr><td align="right"> <span class="gives">&#248;</span> </td><td> <span class="utyp">/o/</span> </td></tr>
<tr><td align="right"> <span class="gives">&#339;</span> </td><td> <span class="utyp">ooee</span> </td></tr>
<tr><td align="right"> <span class="gives">&#254;</span> </td><td> <span class="utyp">pppp</span> </td></tr>
<tr><td align="right"> <span class="gives">&#249;</span> </td><td> <span class="utyp">`+"``"+`u &nbsp; ,,u</span> </td></tr>
<tr><td align="right"> <span class="gives">&#250;</span> </td><td> <span class="utyp">''u</span> </td></tr>
<tr><td align="right"> <span class="gives">&#251;</span> </td><td> <span class="utyp">^^u &nbsp; &lt;&lt;u</span> </td></tr>
<tr><td align="right"> <span class="gives">&#252;</span> </td><td> <span class="utyp">""u</span> </td></tr>
<tr><td align="right"> <span class="gives">&#253;</span> </td><td> <span class="utyp">''y</span> </td></tr>
<tr><td align="right"> <span class="gives">&#255;</span> </td><td> <span class="utyp">""y</span> </td></tr>

<tr><td align="right"> <span class="gives">&#192;</span> </td><td> <span class="utyp">`+"``"+`A &nbsp; ,,A</span> </td></tr>
<tr><td align="right"> <span class="gives">&#193;</span> </td><td> <span class="utyp">''A</span> </td></tr>
<tr><td align="right"> <span class="gives">&#194;</span> </td><td> <span class="utyp">^^A &nbsp; &lt;&lt;A</span> </td></tr>
<tr><td align="right"> <span class="gives">&#195;</span> </td><td> <span class="utyp">~~A</span> </td></tr>
<tr><td align="right"> <span class="gives">&#196;</span> </td><td> <span class="utyp">""A</span> </td></tr>
<tr><td align="right"> <span class="gives">&#197;</span> </td><td> <span class="utyp">oOoA</span> </td></tr>
<tr><td align="right"> <span class="gives">&#198;</span> </td><td> <span class="utyp">AAEE</span> </td></tr>
<tr><td align="right"> <span class="gives">&#199;</span> </td><td> <span class="utyp">C//</span> </td></tr>
<tr><td align="right"> <span class="gives">&#208;</span> </td><td> <span class="utyp">--D-</span> </td></tr>
<tr><td align="right"> <span class="gives">&#200;</span> </td><td> <span class="utyp">`+"``"+`E &nbsp; ,,E</span> </td></tr>
<tr><td align="right"> <span class="gives">&#201;</span> </td><td> <span class="utyp">''E</span> </td></tr>
<tr><td align="right"> <span class="gives">&#202;</span> </td><td> <span class="utyp">^^E &nbsp; &lt;&lt;E</span> </td></tr>
<tr><td align="right"> <span class="gives">&#203;</span> </td><td> <span class="utyp">""E</span> </td></tr>
<tr><td align="right"> <span class="gives">&#204;</span> </td><td> <span class="utyp">`+"``"+`I &nbsp; ,,I</span> </td></tr>
<tr><td align="right"> <span class="gives">&#205;</span> </td><td> <span class="utyp">''I</span> </td></tr>
<tr><td align="right"> <span class="gives">&#206;</span> </td><td> <span class="utyp">^^I &nbsp; &lt;&lt;I</span> </td></tr>
<tr><td align="right"> <span class="gives">&#207;</span> </td><td> <span class="utyp">""I</span> </td></tr>
<tr><td align="right"> <span class="gives">&#209;</span> </td><td> <span class="utyp">~~N</span> </td></tr>
<tr><td align="right"> <span class="gives">&#210;</span> </td><td> <span class="utyp">`+"``"+`O &nbsp; ,,O</span> </td></tr>
<tr><td align="right"> <span class="gives">&#211;</span> </td><td> <span class="utyp">''O</span> </td></tr>
<tr><td align="right"> <span class="gives">&#212;</span> </td><td> <span class="utyp">^^O &nbsp; &lt;&lt;O</span> </td></tr>
<tr><td align="right"> <span class="gives">&#213;</span> </td><td> <span class="utyp">~~O</span> </td></tr>
<tr><td align="right"> <span class="gives">&#214;</span> </td><td> <span class="utyp">""O</span> </td></tr>
<tr><td align="right"> <span class="gives">&#216;</span> </td><td> <span class="utyp">/O/</span> </td></tr>
<tr><td align="right"> <span class="gives">&#338;</span> </td><td> <span class="utyp">OOEE</span> </td></tr>
<tr><td align="right"> <span class="gives">&#222;</span> </td><td> <span class="utyp">PPPP</span> </td></tr>
<tr><td align="right"> <span class="gives">&#223;</span> </td><td> <span class="utyp">ssss</span> </td></tr>
<tr><td align="right"> <span class="gives">&#217;</span> </td><td> <span class="utyp">`+"``"+`U &nbsp; ,,U</span> </td></tr>
<tr><td align="right"> <span class="gives">&#218;</span> </td><td> <span class="utyp">''U</span> </td></tr>
<tr><td align="right"> <span class="gives">&#219;</span> </td><td> <span class="utyp">^^U &nbsp; &lt;&lt;U</span> </td></tr>
<tr><td align="right"> <span class="gives">&#220;</span> </td><td> <span class="utyp">""U</span> </td></tr>
<tr><td align="right"> <span class="gives">&#221;</span> </td><td> <span class="utyp">''Y</span> </td></tr>
<tr><td align="right"> <span class="gives">&#376;</span> </td><td> <span class="utyp">""Y</span> </td></tr>

<tr><td align="right"> <span class="gives">&#171;</span> </td><td> <span class="utyp">&lt;&lt;&lt;</span> </td></tr>
<tr><td align="right"> <span class="gives">&#187;</span> </td><td> <span class="utyp">&gt;&gt;&gt;</span> </td></tr>
<tr><td align="right"> <span class="gives">&#161;</span> </td><td> <span class="utyp">!!!</span> </td></tr>
<tr><td align="right"> <span class="gives">&#191;</span> </td><td> <span class="utyp">???</span> </td></tr>

<tr><td align="right"> <span class="gives">&#352;</span> </td><td> <span class="utyp">^vSv</span> </td></tr>
<tr><td align="right"> <span class="gives">&#353;</span> </td><td> <span class="utyp">^vsv</span> </td></tr>

<tr><td align="right"> <span class="gives">&#215;</span> </td><td> <span class="utyp">\times</span> </td></tr>
<tr><td align="right"> <span class="gives">&#247;</span> </td><td> <span class="utyp">\div</span> </td></tr>
<tr><td align="right"> <span class="gives">&#402;</span> </td><td> <span class="utyp">\fnof</span> </td></tr>
<tr><td align="right"> <span class="gives">&#710;</span> </td><td> <span class="utyp">\circ</span> </td></tr>
<tr><td align="right"> <span class="gives">&#732;</span> </td><td> <span class="utyp">\tilde</span> </td></tr>
<tr><td align="right"> <span class="gives">&#913;</span> </td><td> <span class="utyp">\Alpha</span> </td></tr>
<tr><td align="right"> <span class="gives">&#914;</span> </td><td> <span class="utyp">\Beta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#915;</span> </td><td> <span class="utyp">\Gamma</span> </td></tr>
<tr><td align="right"> <span class="gives">&#916;</span> </td><td> <span class="utyp">\Delta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#917;</span> </td><td> <span class="utyp">\Epsilon</span> </td></tr>
<tr><td align="right"> <span class="gives">&#918;</span> </td><td> <span class="utyp">\Zeta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#919;</span> </td><td> <span class="utyp">\Eta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#920;</span> </td><td> <span class="utyp">\Theta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#921;</span> </td><td> <span class="utyp">\Iota</span> </td></tr>
<tr><td align="right"> <span class="gives">&#922;</span> </td><td> <span class="utyp">\Kappa</span> </td></tr>
<tr><td align="right"> <span class="gives">&#923;</span> </td><td> <span class="utyp">\Lambda</span> </td></tr>
<tr><td align="right"> <span class="gives">&#924;</span> </td><td> <span class="utyp">\Mu</span> </td></tr>
<tr><td align="right"> <span class="gives">&#925;</span> </td><td> <span class="utyp">\Nu</span> </td></tr>
<tr><td align="right"> <span class="gives">&#926;</span> </td><td> <span class="utyp">\Xi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#927;</span> </td><td> <span class="utyp">\Omicron</span> </td></tr>
<tr><td align="right"> <span class="gives">&#928;</span> </td><td> <span class="utyp">\Pi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#929;</span> </td><td> <span class="utyp">\Rho</span> </td></tr>
<tr><td align="right"> <span class="gives">&#931;</span> </td><td> <span class="utyp">\Sigma</span> </td></tr>
<tr><td align="right"> <span class="gives">&#932;</span> </td><td> <span class="utyp">\Tau</span> </td></tr>
<tr><td align="right"> <span class="gives">&#933;</span> </td><td> <span class="utyp">\Upsilon</span> </td></tr>
<tr><td align="right"> <span class="gives">&#934;</span> </td><td> <span class="utyp">\Phi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#935;</span> </td><td> <span class="utyp">\Chi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#936;</span> </td><td> <span class="utyp">\Psi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#937;</span> </td><td> <span class="utyp">\Omega</span> </td></tr>
<tr><td align="right"> <span class="gives">&#945;</span> </td><td> <span class="utyp">\alpha</span> </td></tr>
<tr><td align="right"> <span class="gives">&#946;</span> </td><td> <span class="utyp">\beta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#947;</span> </td><td> <span class="utyp">\gamma</span> </td></tr>
<tr><td align="right"> <span class="gives">&#948;</span> </td><td> <span class="utyp">\delta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#949;</span> </td><td> <span class="utyp">\epsilon</span> </td></tr>
<tr><td align="right"> <span class="gives">&#950;</span> </td><td> <span class="utyp">\zeta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#951;</span> </td><td> <span class="utyp">\eta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#952;</span> </td><td> <span class="utyp">\theta</span> </td></tr>
<tr><td align="right"> <span class="gives">&#953;</span> </td><td> <span class="utyp">\iota</span> </td></tr>
<tr><td align="right"> <span class="gives">&#954;</span> </td><td> <span class="utyp">\kappa</span> </td></tr>
<tr><td align="right"> <span class="gives">&#955;</span> </td><td> <span class="utyp">\lambda</span> </td></tr>
<tr><td align="right"> <span class="gives">&#956;</span> </td><td> <span class="utyp">\mu</span> </td></tr>
<tr><td align="right"> <span class="gives">&#957;</span> </td><td> <span class="utyp">\nu</span> </td></tr>
<tr><td align="right"> <span class="gives">&#958;</span> </td><td> <span class="utyp">\xi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#959;</span> </td><td> <span class="utyp">\omicron</span> </td></tr>
<tr><td align="right"> <span class="gives">&#960;</span> </td><td> <span class="utyp">\pi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#961;</span> </td><td> <span class="utyp">\rho</span> </td></tr>
<tr><td align="right"> <span class="gives">&#962;</span> </td><td> <span class="utyp">\sigmf</span> </td></tr>
<tr><td align="right"> <span class="gives">&#963;</span> </td><td> <span class="utyp">\sigma</span> </td></tr>
<tr><td align="right"> <span class="gives">&#964;</span> </td><td> <span class="utyp">\tau</span> </td></tr>
<tr><td align="right"> <span class="gives">&#965;</span> </td><td> <span class="utyp">\upsilon</span> </td></tr>
<tr><td align="right"> <span class="gives">&#966;</span> </td><td> <span class="utyp">\phi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#967;</span> </td><td> <span class="utyp">\chi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#968;</span> </td><td> <span class="utyp">\psi</span> </td></tr>
<tr><td align="right"> <span class="gives">&#969;</span> </td><td> <span class="utyp">\omega</span> </td></tr>
<tr><td align="right"> <span class="gives">&#970;</span> </td><td> <span class="utyp">\thetsym</span> </td></tr>
<tr><td align="right"> <span class="gives">&#978;</span> </td><td> <span class="utyp">\upsih</span> </td></tr>
<tr><td align="right"> <span class="gives">&#982;</span> </td><td> <span class="utyp">\piv</span> </td></tr>

<tr><td align="right"> <span class="gives">&#162;</span> </td><td> <span class="utyp">\cent</span> </td></tr>
<tr><td align="right"> <span class="gives">&#163;</span> </td><td> <span class="utyp">\pound</span> </td></tr>
<tr><td align="right"> <span class="gives">&#164;</span> </td><td> <span class="utyp">\curren</span> </td></tr>
<tr><td align="right"> <span class="gives">&#165;</span> </td><td> <span class="utyp">\yen</span> </td></tr>
<tr><td align="right"> <span class="gives">&#166;</span> </td><td> <span class="utyp">\brvbar</span> </td></tr>
<tr><td align="right"> <span class="gives">&#167;</span> </td><td> <span class="utyp">\sect</span> </td></tr>
<tr><td align="right"> <span class="gives">&#168;</span> </td><td> <span class="utyp">\umlaut</span> </td></tr>
<tr><td align="right"> <span class="gives">&#169;</span> </td><td> <span class="utyp">\copy</span> </td></tr>
<tr><td align="right"> <span class="gives">&#170;</span> </td><td> <span class="utyp">\ordf</span> </td></tr>
<tr><td align="right"> <span class="gives">&#172;</span> </td><td> <span class="utyp">\not</span> </td></tr>
<tr><td align="right"> <span class="gives">&#173;</span> </td><td> <span class="utyp">\shy</span> </td></tr>
<tr><td align="right"> <span class="gives">&#174;</span> </td><td> <span class="utyp">\reg</span> </td></tr>
<tr><td align="right"> <span class="gives">&#175;</span> </td><td> <span class="utyp">\macr</span> </td></tr>
<tr><td align="right"> <span class="gives">&#176;</span> </td><td> <span class="utyp">\deg</span> </td></tr>
<tr><td align="right"> <span class="gives">&#177;</span> </td><td> <span class="utyp">\pm</span> </td></tr>
<tr><td align="right"> <span class="gives">&#178;</span> </td><td> <span class="utyp">\sup2</span> </td></tr>
<tr><td align="right"> <span class="gives">&#179;</span> </td><td> <span class="utyp">\sup3</span> </td></tr>
<tr><td align="right"> <span class="gives">&#180;</span> </td><td> <span class="utyp">\acute</span> </td></tr>
<tr><td align="right"> <span class="gives">&#181;</span> </td><td> <span class="utyp">u//</span> </td></tr>
<tr><td align="right"> <span class="gives">&#182;</span> </td><td> <span class="utyp">\para</span> </td></tr>
<tr><td align="right"> <span class="gives">&#183;</span> </td><td> <span class="utyp">\cdot</span> </td></tr>
<tr><td align="right"> <span class="gives">&#184;</span> </td><td> <span class="utyp">\cdil</span> </td></tr>
<tr><td align="right"> <span class="gives">&#185;</span> </td><td> <span class="utyp">\sup1</span> </td></tr>
<tr><td align="right"> <span class="gives">&#186;</span> </td><td> <span class="utyp">\ordm</span> </td></tr>
<tr><td align="right"> <span class="gives">&#188;</span> </td><td> <span class="utyp">\frac14</span> </td></tr>
<tr><td align="right"> <span class="gives">&#189;</span> </td><td> <span class="utyp">\frac12</span> </td></tr>
<tr><td align="right"> <span class="gives">&#190;</span> </td><td> <span class="utyp">\frac34</span> </td></tr>

<tr><td align="right"> <span class="gives">&#8211;</span> </td><td> <span class="utyp">\ndash</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8212;</span> </td><td> <span class="utyp">\mdash</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8213;</span> </td><td> <span class="utyp">\horbar</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8216;</span> </td><td> <span class="utyp">\lsquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8217;</span> </td><td> <span class="utyp">\rsquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8218;</span> </td><td> <span class="utyp">\sbquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8220;</span> </td><td> <span class="utyp">\ldquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8221;</span> </td><td> <span class="utyp">\rdquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8222;</span> </td><td> <span class="utyp">\bdquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8224;</span> </td><td> <span class="utyp">\dagger</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8225;</span> </td><td> <span class="utyp">\ddagger</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8226;</span> </td><td> <span class="utyp">\bullet</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8230;</span> </td><td> <span class="utyp">\ldots</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8240;</span> </td><td> <span class="utyp">\permile</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8242;</span> </td><td> <span class="utyp">\prime</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8243;</span> </td><td> <span class="utyp">\Prime</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8249;</span> </td><td> <span class="utyp">\lsaquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8250;</span> </td><td> <span class="utyp">\rsaquo</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8254;</span> </td><td> <span class="utyp">\oline</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8260;</span> </td><td> <span class="utyp">\frasl</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8364;</span> </td><td> <span class="utyp">\euro</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8465;</span> </td><td> <span class="utyp">\image</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8472;</span> </td><td> <span class="utyp">\weierp</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8476;</span> </td><td> <span class="utyp">\real</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8482;</span> </td><td> <span class="utyp">\trade</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8501;</span> </td><td> <span class="utyp">\aleph</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8592;</span> </td><td> <span class="utyp">\leftarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8593;</span> </td><td> <span class="utyp">\uparrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8594;</span> </td><td> <span class="utyp">\rightarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8595;</span> </td><td> <span class="utyp">\downarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8596;</span> </td><td> <span class="utyp">\leftrightarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8629;</span> </td><td> <span class="utyp">\cr</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8656;</span> </td><td> <span class="utyp">\Leftarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8657;</span> </td><td> <span class="utyp">\Uparrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8658;</span> </td><td> <span class="utyp">\Rightarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8659;</span> </td><td> <span class="utyp">\Downarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8660;</span> </td><td> <span class="utyp">\Leftrightarrow</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8704;</span> </td><td> <span class="utyp">\forall</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8706;</span> </td><td> <span class="utyp">\partial</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8707;</span> </td><td> <span class="utyp">\exists</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8709;</span> </td><td> <span class="utyp">\emptyset</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8711;</span> </td><td> <span class="utyp">\nabla</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8712;</span> </td><td> <span class="utyp">\in</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8713;</span> </td><td> <span class="utyp">\notin</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8715;</span> </td><td> <span class="utyp">\ni</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8719;</span> </td><td> <span class="utyp">\prod</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8721;</span> </td><td> <span class="utyp">\sum</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8722;</span> </td><td> <span class="utyp">\minusminus</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8727;</span> </td><td> <span class="utyp">\lowast</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8730;</span> </td><td> <span class="utyp">\surd</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8733;</span> </td><td> <span class="utyp">\propto</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8734;</span> </td><td> <span class="utyp">\infty</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8736;</span> </td><td> <span class="utyp">\angle</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8743;</span> </td><td> <span class="utyp">\wedge</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8744;</span> </td><td> <span class="utyp">\vee</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8745;</span> </td><td> <span class="utyp">\cap</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8746;</span> </td><td> <span class="utyp">\cup</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8747;</span> </td><td> <span class="utyp">\int</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8756;</span> </td><td> <span class="utyp">\therefore</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8764;</span> </td><td> <span class="utyp">\sim</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8773;</span> </td><td> <span class="utyp">\cong</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8776;</span> </td><td> <span class="utyp">\approx</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8800;</span> </td><td> <span class="utyp">\neq</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8801;</span> </td><td> <span class="utyp">\equiv</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8804;</span> </td><td> <span class="utyp">\leq</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8805;</span> </td><td> <span class="utyp">\geq</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8834;</span> </td><td> <span class="utyp">\subset</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8835;</span> </td><td> <span class="utyp">\supset</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8836;</span> </td><td> <span class="utyp">\nsub</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8838;</span> </td><td> <span class="utyp">\sube</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8839;</span> </td><td> <span class="utyp">\supe</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8853;</span> </td><td> <span class="utyp">\oplus</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8855;</span> </td><td> <span class="utyp">\otimes</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8869;</span> </td><td> <span class="utyp">\perp</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8901;</span> </td><td> <span class="utyp">\sdot</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8968;</span> </td><td> <span class="utyp">\lceil</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8969;</span> </td><td> <span class="utyp">\rceil</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8970;</span> </td><td> <span class="utyp">\lfloor</span> </td></tr>
<tr><td align="right"> <span class="gives">&#8971;</span> </td><td> <span class="utyp">\rfloor</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9001;</span> </td><td> <span class="utyp">\lang</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9002;</span> </td><td> <span class="utyp">\rang</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9674;</span> </td><td> <span class="utyp">\Diamond</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9824;</span> </td><td> <span class="utyp">\spadesuit</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9827;</span> </td><td> <span class="utyp">\clubsuit</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9829;</span> </td><td> <span class="utyp">\heartsuit</span> </td></tr>
<tr><td align="right"> <span class="gives">&#9830;</span> </td><td> <span class="utyp">\diamondsuit</span> </td></tr>
</table>

`+theTable+`
            </div>

        </form>

`)
	// <input name="submit" id="submit" type="submit" /><br />
	// <textarea name="langtocvt" id="langtocvt" type="text" cols="80" rows="5" onkeypress="advanceOnReturn(event);" />`+htmlize(userInput["langtocvt"])+`</textarea>
}
