package toolbox

import (
	"strconv"
	"strings"
)

// UInt64ToStr is function to convert uint64 to string data type
func UInt64ToStr(val uint64) string {
	return strconv.FormatUint(val, 10)
}

// UInt64ToStrSlice is function to convert uint64 to string slice with converted uint64 argument as initial member
func UInt64ToStrSlice(val uint64) []string {
	return []string{UInt64ToStr(val)}
}

// StringPointer returns pointer of string
func StringPointer(val string) *string {
	return &val
}

// BoolPointer returns pointer of boolean
func BoolPointer(val bool) *bool {
	return &val
}

// UInt64SliceToCSV is function to convert uint64 slice to string with comma separated values (csv) format
func UInt64SliceToCSV(numbers []uint64) string {
	var stringVars []string

	for _, val := range numbers {
		stringVars = append(stringVars, UInt64ToStr(val))
	}

	return strings.Join(stringVars, ",")
}
