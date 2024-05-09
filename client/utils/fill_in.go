package utils

import "strings"

func FillIn(s string, l int) string {
	tl := len(s) > l
	ss := make([]string, l, l)
	for i, v := range s {
		if i >= l {
			break
		} else if i > 6 && tl {
			ss[i] = "."
		} else {
			ss[i] = string(v)
		}
	}
	for i, v := range ss {
		if string(v) == "" {
			ss[i] = " "
		}
	}
	return strings.Join(ss, "")
}
