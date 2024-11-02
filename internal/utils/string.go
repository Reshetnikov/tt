package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

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
    blue := "\033[1;34m"     // for Type
    yellow := "\033[1;33m"   // for Value
    green := "\033[1;32m"    // for Name
    reset := "\033[0m"       // reset color

    for i := 0; i < len(args); i += 2 {
		var label, value interface{}
        if i+1 >= len(args) {
            label, value = "Dump", args[i]
        } else {
			label, value = args[i], args[i+1]
		}

        fmt.Printf("%s%v:%s %s%T%s = %s%+v%s\n",
            green, label, reset,     // Name
            blue, value, reset,      // Type
            yellow, value, reset)    // Value
    }
}