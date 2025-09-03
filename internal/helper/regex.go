package helper

import "regexp"

func HashRegex() *regexp.Regexp {
	regex, err := regexp.Compile("^[0-9a-f]{64}$")
	if err != nil {
		panic("Regex Failed to Compile!")
	}
	return regex
}
