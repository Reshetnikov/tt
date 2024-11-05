package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type ANSIColors struct {
	Blue   string
	Yellow string
	Green  string
	Red    string
	Reset  string
}

var Colors = ANSIColors{
	Blue:   "\033[1;34m",
	Yellow: "\033[1;33m",
	Green:  "\033[1;32m",
	Red:    "\033[31m",
	Reset:  "\033[0m",
}

func Ukfirst(s string) string {
	if len(s) == 0 {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToUpper(string(r)) + s[size:]
}

// Example:
// utils.Dump("Config", cfg)
// Output:
// Config: *config.Config = &{AppEnv:development DBHost:postgres}
func Dump(args ...interface{}) {
	for i := 0; i < len(args); i += 2 {
		var label, value interface{}
		if i+1 >= len(args) {
			label, value = "Dump", args[i]
		} else {
			label, value = args[i], args[i+1]
		}

		fmt.Printf("%s%v:%s %s%T%s = %s%+v%s\n",
			Colors.Green, label, Colors.Reset, // Name
			Colors.Blue, value, Colors.Reset, // Type
			Colors.Yellow, value, Colors.Reset) // Value
	}
}

func StrColor(str, color string) string {
	return color + str + Colors.Reset
}
