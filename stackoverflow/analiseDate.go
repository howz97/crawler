package stackoverflow

import (
	"strconv"
	"strings"
)

var (
	count     int64 = 1
	temporary int
)

func checkDate(dateNumber string, lastDate *int) bool {
	d := dateToInt(dateNumber)

	if count == 1 {
		temporary = d
	}
	if d > *lastDate {
		count++
		return true
	} else {
		*lastDate = temporary
		count++
		return false
	}
}

func dateToInt(dateNumber string) int {
	ss := []string{"00", "00", "00"}

	a := strings.SplitN(dateNumber, "T", 2)
	sDay := strings.SplitN(a[0], "-", 3)
	sSecond := strings.SplitN(a[1], ":", 3)
	if len(sSecond) == 2 {
		ss[0] = sSecond[0]
		ss[1] = sSecond[1]
		sSecond = ss
	}
	if len(sSecond) == 1 {
		ss[0] = sSecond[0]
		sSecond = ss
	}
	if len(sSecond) == 0 {
		sSecond = ss
	}

	s := sDay[0] + sDay[1] + sDay[2] + sSecond[0] + sSecond[1] + sSecond[2]
	i, err := strconv.Atoi(s)
	if err != nil {
		panic("string to int error !")
	}

	return i

}