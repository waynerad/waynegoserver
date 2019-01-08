package wutil

import (
	"strconv"
	"strings"
)

func intToStr(ii int) string {
	return strconv.FormatInt(int64(ii), 10)
}

// show time interval, e.g. "5 sec", "14 min", etc
func TimeIntervalSecondsToEnglish(inv int) string {
	if inv < 0 {
		return "0 seconds"
	}
	increments := [5]int{60, 60, 24, 7, 52}
	singulars := [6]string{"second", "minute", "hour", "day", "week", "year"}
	plurals := [6]string{"seconds", "minutes", "hours", "days", "weeks", "years"}
	var breakdown [7]int
	var num int
	for i, x := range increments {
		num = inv / x // has to be integer division
		rem := inv - (num * x)
		breakdown[i] = rem
		inv = num
	}
	breakdown[6] = num
	toUse := 2
	isLeading := true
	rv := ""
	for i := 5; i >= 0; i-- {
		if isLeading {
			if breakdown[i] != 0 {
				isLeading = false
			}
		}
		if !isLeading {
			if toUse != 0 {
				if breakdown[i] == 1 {
					rv = rv + ", 1 " + singulars[i]
				} else {
					if breakdown[i] != 0 {
						rv = rv + ", " + intToStr(breakdown[i]) + " " + plurals[i]
					}
				}
				toUse--
			}
		}
	}
	return rv[1:]
}

func Trim(z string) string {
	return strings.Trim(z, " \r\n\t")
}

func FindInString(stg string, substr string, offset int) int {
	idx := strings.Index(stg, substr)
	if offset < 0 {
		return idx
	}
	if idx >= offset {
		return idx
	}
	part2 := stg[offset:]
	idx = strings.Index(part2, substr)
	if idx < 0 {
		return idx
	}
	return offset + idx
}

// based on:
// https://perishablepress.com/stop-using-unsafe-characters-in-urls/
// which in turn is based on the RFCs:
// https://www.ietf.org/rfc/rfc3986.txt
// https://www.ietf.org/rfc/rfc1738.txt
func ValidURLCharacter(ch byte) bool {
	// numbers are allowed
	if (ch >= 48) && (ch <= 57) {
		return true
	}
	// uppercase letters are allowed
	if (ch >= 65) && (ch <= 90) {
		return true
	}
	// lowercase letters are allowed
	if (ch >= 97) && (ch <= 122) {
		return true
	}
	// the following characters are allowed: # 35 $ 36 % 37 & 38 ' 39 ( 40 ) 41 * 42 + 43 , 44 - 45 . 46 / 47
	if (ch >= 35) && (ch <= 47) {
		return true
	}
	// the following characters are allowed: ! 33 : 58 ; 59 = 61 ? 63 @ 64 _ 95
	if (ch == 33) || (ch == 58) || (ch == 59) || (ch == 61) || (ch == 63) || (ch == 64) || (ch == 95) {
		return true
	}
	return false
}

func FindEndOfLink(content string, idx int) int {
	lx := len(content)
	for {
		if idx >= lx {
			return idx
		}
		ch := content[idx]
		if !ValidURLCharacter(ch) {
			return idx
		}
		idx++
	}
}

func ReplaceLink(content string, idx int) (string, int) {
	eolnk := FindEndOfLink(content, idx)
	linkURL := content[idx:eolnk]
	replacement := `<a href="` + linkURL + `">` + linkURL + "</a>"
	additional := len(replacement) - len(linkURL)
	return content[:idx] + replacement + content[eolnk:], eolnk + additional
}

func Linkize(content string) string {
	idx := FindInString(content, "http://", -1)
	for idx >= 0 {
		content, idx = ReplaceLink(content, idx)
		idx = FindInString(content, "http://", idx)
	}
	idx = FindInString(content, "https://", -1)
	for idx >= 0 {
		content, idx = ReplaceLink(content, idx)
		idx = FindInString(content, "https://", idx)
	}
	return content
}

type lineProcessing struct {
	content         string
	idx             int
	beginningOfLine int
	endOfLine       int
	inBulletList    bool
	inNumberedList  bool
	result          string
}

func ProcessLine(ch byte, lp *lineProcessing) {
	if ch == 13 {
		lp.endOfLine = lp.idx
		inLine := lp.content[lp.beginningOfLine:lp.endOfLine]
		if inLine == "" {
			if lp.inBulletList {
				lp.result += `</ul>
`
				lp.inBulletList = false
			}
			if lp.inNumberedList {
				lp.result += `</ol>
`
				lp.inNumberedList = false
			}
			lp.result += `</p>
<p>`
		} else {
			switch inLine[0] {
			case 124:
				// pipe indicates headers
				if lp.inBulletList {
					lp.result += `</ul>
`
					lp.inBulletList = false
				}
				if lp.inNumberedList {
					lp.result += `</ol>
`
					lp.inNumberedList = false
				}
				lp.result += "<h2>" + Trim(inLine[1:]) + `</h2>
`
			case 42:
				// asterisk indicates bullet list
				if lp.inNumberedList {
					lp.result += "</ol>"
					lp.inNumberedList = false
				}
				if !lp.inBulletList {
					lp.inBulletList = true
					lp.result += `<ul>
`
				}
				lp.result += "<li>" + Trim(inLine[1:]) + `</li>
`
			case 35:
				// hash indicates numbered list
				if !lp.inNumberedList {
					lp.inNumberedList = true
					lp.result += `
<ol>
`
				}
				lp.result += "<li>" + Trim(inLine[1:]) + `</li>
`
			default:
				if lp.inBulletList {
					lp.result += `</ul>
`
					lp.inBulletList = false
				}
				if lp.inNumberedList {
					lp.result += `
</ol>
`
					lp.inNumberedList = false
				}
				lp.result += inLine + `
`
			}
		}
		lp.beginningOfLine = -1
	}
	if ch == 10 {
		lp.beginningOfLine = lp.idx + 1
	}
}

func Markdownize(content string) string {
	// line-based processing
	var lp lineProcessing
	lp.content = content
	lp.result = "<p>"
	lp.idx = 0
	lp.beginningOfLine = 0
	lp.endOfLine = 0
	lp.inBulletList = false
	lp.inNumberedList = false
	lx := len(content)
	for lp.idx < lx {
		ch := content[lp.idx]
		ProcessLine(ch, &lp)
		lp.idx++
	}
	// we add a "virtual CR" to the end to process the last line
	// finalize everything
	if lp.beginningOfLine >= 0 {
		ch := byte(13)
		ProcessLine(ch, &lp)
		if lp.inBulletList {
			lp.result += `</ul>
`
		}
		if lp.inNumberedList {
			lp.result += `</ol>
`
		}
	}
	lp.result += "</p>"
	// linkize
	idxh := strings.Index(lp.result, "http://")
	idxhs := strings.Index(lp.result, "https://")
	if (idxh >= 0) || (idxhs >= 0) {
		lp.result = Linkize(lp.result)
	}
	return lp.result
}
