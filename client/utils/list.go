package utils

func EqualListEntries(le, nle []any) bool {
	if len(le) != len(nle) {
		return false
	} else {
		for i, e := range nle {
			if le[i].(ListEntry) != e.(ListEntry) {
				return false
			}
		}
	}
	return true
}

type ListEntry struct {
	ID   string
	Text string
}
