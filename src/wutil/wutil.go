package wutil

import (
	"strconv"
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
