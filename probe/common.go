package probe

import (
	"fmt"
	"strings"
)

// CommandLine will return the whole command line which includes command and all arguments
func CommandLine(cmd string, args []string) string {
	result := cmd
	for _, arg := range args {
		result += " " + arg
	}
	return result
}

// CheckOutput checks the output text,
// - if it contains a configured string then return nil
// - if it does not contain a configured string then return nil
func CheckOutput(Contain, NotContain string, Output string) error {

	if len(Contain) > 0 && !strings.Contains(Output, Contain) {

		return fmt.Errorf("the output does not contain [%s]", Contain)
	}

	if len(NotContain) > 0 && strings.Contains(Output, NotContain) {
		return fmt.Errorf("the output contains [%s]", NotContain)

	}
	return nil
}

// CheckEmpty return "empty" if the string is empty
func CheckEmpty(s string) string {
	if len(strings.TrimSpace(s)) <= 0 {
		return "empty"
	}
	return s
}
