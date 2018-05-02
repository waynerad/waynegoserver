package langtext

import (
	"fmt"
	"forms"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"html"
	"langtextui"
	"login"
	"net/http"
	"static"
	"strconv"
)

func htmlize(z string) string {
	return html.EscapeString(z)
}

////////////////////////////////////////////////////////////////
// Lang Text Form
////////////////////////////////////////////////////////////////

type replaceEntry struct {
	match       []rune
	replacement rune
}

func mkReplaceEntry(srch string, replacement rune) replaceEntry {
	var rv replaceEntry
	rv.match = []rune(srch)
	rv.replacement = replacement
	return rv
}

func testReplaceMapCreatedCorrectly(replaceMap []replaceEntry, rpllen int) {
	for i := 0; i < rpllen; i++ {
		if replaceMap[i].match == nil {
			fmt.Println("Match is nil for entry:", i)
		}
		if replaceMap[i].replacement == 0 {
			fmt.Println("Replacement is 0 for entry:", i)
		}
	}
}

func constructReplaceMap() []replaceEntry {
	replaceMap := make([]replaceEntry, 262)
	backslash := "\\"

	replaceMap[0] = mkReplaceEntry("!!!", 161)
	replaceMap[1] = mkReplaceEntry(backslash+"cent", 162)
	replaceMap[2] = mkReplaceEntry(backslash+"pound", 163)
	replaceMap[3] = mkReplaceEntry(backslash+"curren", 164)
	replaceMap[4] = mkReplaceEntry(backslash+"yen", 165)
	replaceMap[5] = mkReplaceEntry(backslash+"brvbar", 166)
	replaceMap[6] = mkReplaceEntry(backslash+"sect", 167)
	replaceMap[7] = mkReplaceEntry(backslash+"umlaut", 168)
	replaceMap[8] = mkReplaceEntry(backslash+"copy", 169)
	replaceMap[9] = mkReplaceEntry(backslash+"ordf", 170)
	replaceMap[10] = mkReplaceEntry("<<<", 171)
	replaceMap[11] = mkReplaceEntry(backslash+"not", 172)
	replaceMap[12] = mkReplaceEntry(backslash+"shy", 173)
	replaceMap[13] = mkReplaceEntry(backslash+"reg", 174)
	replaceMap[14] = mkReplaceEntry(backslash+"macr", 175)
	replaceMap[15] = mkReplaceEntry(backslash+"deg", 176)
	replaceMap[16] = mkReplaceEntry(backslash+"pm", 177)
	replaceMap[17] = mkReplaceEntry(backslash+"sup2", 178)
	replaceMap[18] = mkReplaceEntry(backslash+"sup3", 179)
	replaceMap[19] = mkReplaceEntry(backslash+"acute", 180)
	replaceMap[20] = mkReplaceEntry("u//", 181)
	replaceMap[21] = mkReplaceEntry(backslash+"para", 182)
	replaceMap[22] = mkReplaceEntry(backslash+"cdot", 183)
	replaceMap[23] = mkReplaceEntry(backslash+"cdil", 184)
	replaceMap[24] = mkReplaceEntry(backslash+"sup1", 185)
	replaceMap[25] = mkReplaceEntry(backslash+"ordm", 186)
	replaceMap[26] = mkReplaceEntry(">>>", 187)
	replaceMap[27] = mkReplaceEntry(backslash+"frac14", 188)
	replaceMap[28] = mkReplaceEntry(backslash+"frac12", 189)
	replaceMap[29] = mkReplaceEntry(backslash+"frac34", 190)
	replaceMap[30] = mkReplaceEntry("???", 191)
	replaceMap[31] = mkReplaceEntry("``A", 192)
	replaceMap[32] = mkReplaceEntry(",,A", 192)
	replaceMap[33] = mkReplaceEntry("''A", 193)
	replaceMap[34] = mkReplaceEntry("^^A", 194)
	replaceMap[35] = mkReplaceEntry("<<A", 194)

	replaceMap[36] = mkReplaceEntry("~~A", 195)
	replaceMap[37] = mkReplaceEntry(`""A`, 196)
	replaceMap[38] = mkReplaceEntry("oOoA", 197)
	replaceMap[39] = mkReplaceEntry("AAEE", 198)
	replaceMap[40] = mkReplaceEntry("C//", 199)
	replaceMap[41] = mkReplaceEntry("``E", 200)
	replaceMap[42] = mkReplaceEntry(",,E", 200)
	replaceMap[43] = mkReplaceEntry("''E", 201)
	replaceMap[44] = mkReplaceEntry("^^E", 202)
	replaceMap[45] = mkReplaceEntry("<<E", 202)

	replaceMap[46] = mkReplaceEntry(`""E`, 203)
	replaceMap[47] = mkReplaceEntry("``I", 204)
	replaceMap[48] = mkReplaceEntry(",,I", 204)
	replaceMap[49] = mkReplaceEntry("''I", 205)
	replaceMap[50] = mkReplaceEntry("^^I", 206)
	replaceMap[51] = mkReplaceEntry("<<I", 206)

	replaceMap[52] = mkReplaceEntry(`""I`, 207)
	replaceMap[53] = mkReplaceEntry("--D-", 208)
	replaceMap[54] = mkReplaceEntry("~~N", 209)
	replaceMap[55] = mkReplaceEntry("``O", 210)
	replaceMap[56] = mkReplaceEntry(",,O", 210)
	replaceMap[57] = mkReplaceEntry("''O", 211)
	replaceMap[58] = mkReplaceEntry("^^O", 212)
	replaceMap[59] = mkReplaceEntry("<<O", 212)

	replaceMap[60] = mkReplaceEntry("~~O", 213)
	replaceMap[61] = mkReplaceEntry(`""O`, 214)
	replaceMap[62] = mkReplaceEntry(backslash+"times", 215)
	replaceMap[63] = mkReplaceEntry("/O/", 216)
	replaceMap[64] = mkReplaceEntry("``U", 217)
	replaceMap[65] = mkReplaceEntry(",,U", 217)
	replaceMap[66] = mkReplaceEntry("''U", 218)
	replaceMap[67] = mkReplaceEntry("^^U", 219)
	replaceMap[68] = mkReplaceEntry("<<U", 219)

	replaceMap[69] = mkReplaceEntry(`""U`, 220)
	replaceMap[70] = mkReplaceEntry("''Y", 221)
	replaceMap[71] = mkReplaceEntry("PPPP", 222)
	replaceMap[72] = mkReplaceEntry("ssss", 223)
	replaceMap[73] = mkReplaceEntry("``a", 224)
	replaceMap[74] = mkReplaceEntry(",,a", 224)
	replaceMap[75] = mkReplaceEntry("''a", 225)
	replaceMap[76] = mkReplaceEntry("^^a", 226)
	replaceMap[77] = mkReplaceEntry("<<a", 226)

	replaceMap[78] = mkReplaceEntry("~~a", 227)
	replaceMap[79] = mkReplaceEntry(`""a`, 228)
	replaceMap[80] = mkReplaceEntry("oOoa", 229)
	replaceMap[81] = mkReplaceEntry("aaee", 230)
	replaceMap[82] = mkReplaceEntry("c//", 231)
	replaceMap[83] = mkReplaceEntry("``e", 232)
	replaceMap[84] = mkReplaceEntry(",,e", 232)
	replaceMap[85] = mkReplaceEntry("''e", 233)
	replaceMap[86] = mkReplaceEntry("^^e", 234)
	replaceMap[87] = mkReplaceEntry("<<e", 234)

	replaceMap[88] = mkReplaceEntry(`""e`, 235)
	replaceMap[89] = mkReplaceEntry("``i", 236)
	replaceMap[90] = mkReplaceEntry(",,i", 236)
	replaceMap[91] = mkReplaceEntry("''i", 237)
	replaceMap[92] = mkReplaceEntry("^^i", 238)
	replaceMap[93] = mkReplaceEntry("<<i", 238)

	replaceMap[94] = mkReplaceEntry(`""i`, 239)
	replaceMap[95] = mkReplaceEntry("--d--", 240)
	replaceMap[96] = mkReplaceEntry("~~n", 241)
	replaceMap[97] = mkReplaceEntry("``o", 242)
	replaceMap[98] = mkReplaceEntry(",,o", 242)
	replaceMap[99] = mkReplaceEntry("''o", 243)
	replaceMap[100] = mkReplaceEntry("^^o", 244)
	replaceMap[101] = mkReplaceEntry("<<o", 244)

	replaceMap[102] = mkReplaceEntry("~~o", 245)
	replaceMap[103] = mkReplaceEntry(`""o`, 246)
	replaceMap[104] = mkReplaceEntry(backslash+"div", 247)
	replaceMap[105] = mkReplaceEntry("/o/", 248)
	replaceMap[106] = mkReplaceEntry("``u", 249)
	replaceMap[107] = mkReplaceEntry(",,u", 249)
	replaceMap[108] = mkReplaceEntry("''u", 250)
	replaceMap[109] = mkReplaceEntry("^^u", 251)
	replaceMap[110] = mkReplaceEntry("<<u", 251)

	replaceMap[111] = mkReplaceEntry(`""u`, 252)
	replaceMap[112] = mkReplaceEntry("''y", 253)
	replaceMap[113] = mkReplaceEntry("pppp", 254)
	replaceMap[114] = mkReplaceEntry(`""y`, 255)
	replaceMap[115] = mkReplaceEntry("OOEE", 338)
	replaceMap[116] = mkReplaceEntry("ooee", 339)
	replaceMap[117] = mkReplaceEntry("^vSv^", 352)
	replaceMap[118] = mkReplaceEntry("^vsv^", 353)
	replaceMap[119] = mkReplaceEntry(`""Y`, 376)
	replaceMap[120] = mkReplaceEntry(backslash+"fnof", 402)
	replaceMap[121] = mkReplaceEntry(backslash+"circ", 710)
	replaceMap[122] = mkReplaceEntry(backslash+"tilde", 732)
	replaceMap[123] = mkReplaceEntry(backslash+"Alpha", 913)
	replaceMap[124] = mkReplaceEntry(backslash+"Beta", 914)
	replaceMap[125] = mkReplaceEntry(backslash+"Gamma", 915)
	replaceMap[126] = mkReplaceEntry(backslash+"Delta", 916)
	replaceMap[127] = mkReplaceEntry(backslash+"Epsilon", 917)
	replaceMap[128] = mkReplaceEntry(backslash+"Zeta", 918)
	replaceMap[129] = mkReplaceEntry(backslash+"Eta", 919)
	replaceMap[130] = mkReplaceEntry(backslash+"Theta", 920)
	replaceMap[131] = mkReplaceEntry(backslash+"Iota", 921)
	replaceMap[132] = mkReplaceEntry(backslash+"Kappa", 922)
	replaceMap[133] = mkReplaceEntry(backslash+"Lambda", 923)
	replaceMap[134] = mkReplaceEntry(backslash+"Mu", 924)
	replaceMap[135] = mkReplaceEntry(backslash+"Nu", 925)
	replaceMap[136] = mkReplaceEntry(backslash+"Xi", 926)
	replaceMap[137] = mkReplaceEntry(backslash+"Omicron", 927)
	replaceMap[138] = mkReplaceEntry(backslash+"Pi", 928)
	replaceMap[139] = mkReplaceEntry(backslash+"Rho", 929)
	replaceMap[140] = mkReplaceEntry(backslash+"Sigma", 931)
	replaceMap[141] = mkReplaceEntry(backslash+"Tau", 932)
	replaceMap[142] = mkReplaceEntry(backslash+"Upsilon", 933)
	replaceMap[143] = mkReplaceEntry(backslash+"Phi", 934)
	replaceMap[144] = mkReplaceEntry(backslash+"Chi", 935)
	replaceMap[145] = mkReplaceEntry(backslash+"Psi", 936)
	replaceMap[146] = mkReplaceEntry(backslash+"Omega", 937)
	replaceMap[147] = mkReplaceEntry(backslash+"alpha", 945)
	replaceMap[148] = mkReplaceEntry(backslash+"beta", 946)
	replaceMap[149] = mkReplaceEntry(backslash+"gamma", 947)
	replaceMap[150] = mkReplaceEntry(backslash+"delta", 948)
	replaceMap[151] = mkReplaceEntry(backslash+"epsilon", 949)
	replaceMap[152] = mkReplaceEntry(backslash+"zeta", 950)
	replaceMap[153] = mkReplaceEntry(backslash+"eta", 951)
	replaceMap[154] = mkReplaceEntry(backslash+"theta", 952)
	replaceMap[155] = mkReplaceEntry(backslash+"iota", 953)
	replaceMap[156] = mkReplaceEntry(backslash+"kappa", 954)
	replaceMap[157] = mkReplaceEntry(backslash+"lambda", 955)
	replaceMap[158] = mkReplaceEntry(backslash+"mu", 956)
	replaceMap[159] = mkReplaceEntry(backslash+"nu", 957)
	replaceMap[160] = mkReplaceEntry(backslash+"xi", 958)
	replaceMap[161] = mkReplaceEntry(backslash+"omicron", 959)
	replaceMap[162] = mkReplaceEntry(backslash+"pi", 960)
	replaceMap[163] = mkReplaceEntry(backslash+"rho", 961)
	replaceMap[164] = mkReplaceEntry(backslash+"sigmf", 962)
	replaceMap[165] = mkReplaceEntry(backslash+"sigma", 963)
	replaceMap[166] = mkReplaceEntry(backslash+"tau", 964)
	replaceMap[167] = mkReplaceEntry(backslash+"upsilon", 965)
	replaceMap[168] = mkReplaceEntry(backslash+"phi", 966)
	replaceMap[169] = mkReplaceEntry(backslash+"chi", 967)
	replaceMap[170] = mkReplaceEntry(backslash+"psi", 968)
	replaceMap[171] = mkReplaceEntry(backslash+"omega", 969)
	replaceMap[172] = mkReplaceEntry(backslash+"thetsym", 970)
	replaceMap[173] = mkReplaceEntry(backslash+"upsih", 978)
	replaceMap[174] = mkReplaceEntry(backslash+"piv", 982)
	replaceMap[175] = mkReplaceEntry(backslash+"ndash", 8211)
	replaceMap[176] = mkReplaceEntry(backslash+"mdash", 8212)
	replaceMap[177] = mkReplaceEntry(backslash+"horbar", 8213)
	replaceMap[178] = mkReplaceEntry(backslash+"lsquo", 8216)
	replaceMap[179] = mkReplaceEntry(backslash+"rsquo", 8217)
	replaceMap[180] = mkReplaceEntry(backslash+"sbquo", 8218)
	replaceMap[181] = mkReplaceEntry(backslash+"ldquo", 8220)
	replaceMap[182] = mkReplaceEntry(backslash+"rdquo", 8221)
	replaceMap[183] = mkReplaceEntry(backslash+"bdquo", 8222)
	replaceMap[184] = mkReplaceEntry(backslash+"dagger", 8224)
	replaceMap[185] = mkReplaceEntry(backslash+"ddagger", 8225)
	replaceMap[186] = mkReplaceEntry(backslash+"bullet", 8226)
	replaceMap[187] = mkReplaceEntry(backslash+"ldots", 8230)
	replaceMap[188] = mkReplaceEntry(backslash+"permile", 8240)
	replaceMap[189] = mkReplaceEntry(backslash+"prime", 8242)
	replaceMap[190] = mkReplaceEntry(backslash+"Prime", 8243)
	replaceMap[191] = mkReplaceEntry(backslash+"lsaquo", 8249)
	replaceMap[192] = mkReplaceEntry(backslash+"rsaquo", 8250)
	replaceMap[193] = mkReplaceEntry(backslash+"oline", 8254)
	replaceMap[194] = mkReplaceEntry(backslash+"frasl", 8260)
	replaceMap[195] = mkReplaceEntry(backslash+"euro", 8364)
	replaceMap[196] = mkReplaceEntry(backslash+"image", 8465)
	replaceMap[197] = mkReplaceEntry(backslash+"weierp", 8472)
	replaceMap[198] = mkReplaceEntry(backslash+"real", 8476)
	replaceMap[199] = mkReplaceEntry(backslash+"trade", 8482)
	replaceMap[200] = mkReplaceEntry(backslash+"alefsym", 8501)

	replaceMap[201] = mkReplaceEntry(backslash+"leftarrow", 8592)
	replaceMap[202] = mkReplaceEntry(backslash+"uparrow", 8593)
	replaceMap[203] = mkReplaceEntry(backslash+"rightarrow", 8594)
	replaceMap[204] = mkReplaceEntry(backslash+"downarrow", 8595)
	replaceMap[205] = mkReplaceEntry(backslash+"leftrightarrow", 8596)

	replaceMap[206] = mkReplaceEntry(backslash+"cr", 8629)

	replaceMap[207] = mkReplaceEntry(backslash+"Leftarrow", 8656)
	replaceMap[208] = mkReplaceEntry(backslash+"Uparrow", 8657)
	replaceMap[209] = mkReplaceEntry(backslash+"Rightarrow", 8658)
	replaceMap[210] = mkReplaceEntry(backslash+"Downarrow", 8659)
	replaceMap[211] = mkReplaceEntry(backslash+"Leftrightarrow", 8660)
	replaceMap[212] = mkReplaceEntry(backslash+"forall", 8704)
	replaceMap[213] = mkReplaceEntry(backslash+"partial", 8706)
	replaceMap[214] = mkReplaceEntry(backslash+"exists", 8707)
	replaceMap[215] = mkReplaceEntry(backslash+"emptyset", 8709)
	replaceMap[216] = mkReplaceEntry(backslash+"nabla", 8711)
	replaceMap[217] = mkReplaceEntry(backslash+"in", 8712)
	replaceMap[218] = mkReplaceEntry(backslash+"notin", 8713)
	replaceMap[219] = mkReplaceEntry(backslash+"ni", 8715)
	replaceMap[220] = mkReplaceEntry(backslash+"prod", 8719)
	replaceMap[221] = mkReplaceEntry(backslash+"sum", 8721)
	replaceMap[222] = mkReplaceEntry(backslash+"minusminus", 8722)
	replaceMap[223] = mkReplaceEntry(backslash+"lowast", 8727)
	replaceMap[224] = mkReplaceEntry(backslash+"surd", 8730)
	replaceMap[225] = mkReplaceEntry(backslash+"propto", 8733)
	replaceMap[226] = mkReplaceEntry(backslash+"infty", 8734)
	replaceMap[227] = mkReplaceEntry(backslash+"angle", 8736)
	replaceMap[228] = mkReplaceEntry(backslash+"wedge", 8743)
	replaceMap[229] = mkReplaceEntry(backslash+"vee", 8744)
	replaceMap[230] = mkReplaceEntry(backslash+"cap", 8745)
	replaceMap[231] = mkReplaceEntry(backslash+"cup", 8746)
	replaceMap[232] = mkReplaceEntry(backslash+"int", 8747)
	replaceMap[233] = mkReplaceEntry(backslash+"therefore", 8756)
	replaceMap[234] = mkReplaceEntry(backslash+"sim", 8764)
	replaceMap[235] = mkReplaceEntry(backslash+"cong", 8773)
	replaceMap[236] = mkReplaceEntry(backslash+"approx", 8776)
	replaceMap[237] = mkReplaceEntry(backslash+"neq", 8800)
	replaceMap[238] = mkReplaceEntry(backslash+"equiv", 8801)
	replaceMap[239] = mkReplaceEntry(backslash+"leq", 8804)
	replaceMap[240] = mkReplaceEntry(backslash+"geq", 8805)
	replaceMap[241] = mkReplaceEntry(backslash+"subset", 8834)
	replaceMap[242] = mkReplaceEntry(backslash+"supset", 8835)
	replaceMap[243] = mkReplaceEntry(backslash+"nsub", 8836)
	replaceMap[244] = mkReplaceEntry(backslash+"sube", 8838)
	replaceMap[245] = mkReplaceEntry(backslash+"supe", 8839)
	replaceMap[246] = mkReplaceEntry(backslash+"oplus", 8853)
	replaceMap[247] = mkReplaceEntry(backslash+"otimes", 8855)
	replaceMap[248] = mkReplaceEntry(backslash+"perp", 8869)
	replaceMap[249] = mkReplaceEntry(backslash+"sdot", 8901)
	replaceMap[250] = mkReplaceEntry(backslash+"lceil", 8968)
	replaceMap[251] = mkReplaceEntry(backslash+"rceil", 8969)
	replaceMap[252] = mkReplaceEntry(backslash+"lfloor", 8970)
	replaceMap[253] = mkReplaceEntry(backslash+"rfloor", 8971)
	replaceMap[254] = mkReplaceEntry(backslash+"lang", 9001)
	replaceMap[255] = mkReplaceEntry(backslash+"rang", 9002)
	replaceMap[256] = mkReplaceEntry(backslash+"diamond", 9674)
	replaceMap[257] = mkReplaceEntry(backslash+"spadessuit", 9824)
	replaceMap[258] = mkReplaceEntry(backslash+"clubssuit", 9827)
	replaceMap[259] = mkReplaceEntry(backslash+"heartssuit", 9829)
	replaceMap[260] = mkReplaceEntry(backslash+"diamondssuit", 9830)
	replaceMap[261] = mkReplaceEntry("'''", 0) // special entry to solve L'''e problem
	return replaceMap
}

func ProcessInternationalLanguageCharacterConversion(sometext string) string {
	replaceMap := constructReplaceMap()

	rpllen := len(replaceMap)

	rv := ""

	statemap := make([]int, rpllen)
	buffer := make([]rune, 100) // must be big enough to hold the largest replace match entry
	bufend := 0
	for i := 0; i < rpllen; i++ {
		statemap[i] = 0 // default everything to no matches
	}
	for _, currentChar := range sometext {
		anyMatch := false
		outputDone := false
		for i := 0; i < rpllen; i++ {
			if len(replaceMap[i].match) > statemap[i] {
				if replaceMap[i].match[statemap[i]] == currentChar {
					anyMatch = true
					statemap[i]++
					if len(replaceMap[i].match) == statemap[i] {
						if i == 261 {
							// whacky hack to get L''e to work
							// If the user typed 3 's in a row, we chop the first one out of the state map
							// but we leave it in the buffer so it will get outputted
							for j := 0; j < rpllen; j++ {
								if len(replaceMap[j].match) > 1 {
									if (replaceMap[j].match[0] == 39) && (replaceMap[j].match[1] == 39) {
										statemap[j] = 2
									}
								}
							}
						} else {
							// we've reached the end
							// empty everything in the buffer that isn't us
							for j := 0; j < bufend-(statemap[i]-1); j++ {
								rv = rv + string(buffer[j])
							}
							// replace the part that is us
							rv = rv + string(replaceMap[i].replacement)
							// clear the buffer
							bufend = 0
							// reset the statemap
							for j := 0; j < rpllen; j++ {
								statemap[j] = 0
							}
							outputDone = true
							i = rpllen // short-circuit the loop
						}
					}
				} else {
					statemap[i] = 0
				}
			}
		}
		if anyMatch {
			if !outputDone {
				buffer[bufend] = currentChar
				bufend++
			}
		} else {
			for j := 0; j < bufend; j++ {
				rv = rv + string(buffer[j])
			}
			rv = rv + string(currentChar)
			bufend = 0
		}
	}
	if bufend != 0 {
		for j := 0; j < bufend; j++ {
			rv = rv + string(buffer[j])
		}
	}
	return rv
}

func toStr(xy int) string {
	return strconv.FormatInt(int64(xy), 10)
}

func makeReplaceTableForUi() string {
	return "" // we disable this except when necessary to re-create the table in the UI (which is otherwise hard-coded)
	replaceMap := constructReplaceMap()
	rv := `<table>
`
	for _, entry := range replaceMap {
		if entry.replacement != 0 {
			rv += `<tr><td align="right"> <span class="gives">`
			rv += "&#" + toStr(int(entry.replacement)) + ";"
			rv += `</span> </td><td> <span class="utyp">`
			for _, ch := range entry.match {
				if ch == 60 {
					rv += "&lt;"
				} else {
					if ch == 62 {
						rv += "&gt;"
					} else {
						rv += string(ch)
					}
				}
			}
			rv += `</span> </td></tr>
`
		}
	}
	rv += `</table>
`
	return rv
}

// ----------------------------------------------------------------
// Lang Type Page
// ----------------------------------------------------------------

func showLangTextForm(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string, theTable string) {
	displayInfo := make(map[string]string)
	displayInfo["hTitle"] = "Lang Type"
	displayInfo["hUserName"] = htmlize(userInfo.UserName)
	displayInfo["kn"] = "0"
	langtextui.ShowHeadHeader(w, displayInfo)
	langtextui.ShowBodyHeader(w, displayInfo)
	langtextui.ShowLangTextForm(w, errorList, userInput, displayInfo, theTable)
	langtextui.ShowFooter(w, displayInfo)
}

type taskEditForm struct {
	objectName string
}

func (self *taskEditForm) GetDefaults(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string) map[string]string {
	rv := make(map[string]string)
	rv["langtocvt"] = ""
	return rv
}

func (self *taskEditForm) GetDBDataAndShowForm(db mysql.Conn, w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord, errorList map[string]string, userInput map[string]string) {
	// no DB data
	// This is a hack -- we process right before we display because we stay on this form and give the user their input back to them -- processed
	langtocvt := userInput["langtocvt"]
	// fmt.Println("langtocvt is", langtocvt)
	processed := ProcessInternationalLanguageCharacterConversion(langtocvt)
	// fmt.Println("new langtocvt is", processed)
	userInput["langtocvt"] = processed
	theTable := makeReplaceTableForUi()
	showLangTextForm(w, r, op, userInfo, errorList, userInput, theTable)
}

func (self *taskEditForm) CheckForErrors(db mysql.Conn, userInput map[string]string) (map[string]string, map[string]string) {
	errorList := make(map[string]string)
	errorList["null"] = "null" // force an error so we stay on this page
	return errorList, nil
}

func (self *taskEditForm) SaveForm(db mysql.Conn, userInfo *login.UserInformationRecord, userInput map[string]string, alreadyProcessed map[string]string) map[string]string {
	return nil
}

// ----------------------------------------------------------------
// End of langtext page
// ----------------------------------------------------------------

func showLangTextPage(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	var formObject taskEditForm
	formObject.objectName = "Task Edit Form"
	forms.HandleForm(&formObject, w, r, op, userInfo)
}

func Handler(w http.ResponseWriter, r *http.Request, op string, userInfo *login.UserInformationRecord) {
	fmt.Println("op is", op)
	switch {
	case op == "waynetype":
		if userInfo.UserId != 0 {
			showLangTextPage(w, r, op, userInfo)
		}
	default:
		filename := "/home/ec2-user/wayneserver/staticappcontent/langtext/" + op
		static.OutputStaticFileWithContentType(w, filename)
	}
}