//+build !noexit

package xcobra

import (
	"os"
)

// ExitWithCode will exit the program with a error code if provided, else 1
func ExitWithCode(err error) {
	if e, ok := err.(ErrorWithCode); ok {
		os.Exit(e.Code)
	}
	os.Exit(1)
}
