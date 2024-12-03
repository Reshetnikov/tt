package utils

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

func StrColor(str, color string) string {
	return color + str + Colors.Reset
}
