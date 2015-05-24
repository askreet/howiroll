// A set of functions for repeating a function execution until
// an expected value is met, with some contextual output.
//
// TODO: There is probably a way to implement timeouts using chans
//       that would make a lot more sense here.
package waitfor

import (
	"fmt"
	"time"
)

const timeout = 300

func contains(needle string, haystack []string) bool {
	for _, i := range haystack {
		if i == needle {
			return true
		}
	}
	return false
}

func AdditionalString(msg string, fn func() []string, knownSet []string) string {
	start := time.Now()

	for {
		for _, i := range fn() {
			if !contains(i, knownSet) {
				fmt.Println("")
				return i
			}
		}
		fmt.Printf("\r(%3.0fs) %s        ", time.Since(start).Seconds(), msg)
		time.Sleep(3 * time.Second)
	}
}

func Strings(msg string, fn func() string, acceptable []string) {
	start := time.Now()

	for {
		val := fn()
		for _, acc := range acceptable {
			if acc == val {
				fmt.Println("")
				return
			}
		}
		fmt.Printf("\r(%3.0fs) %s        ", time.Since(start).Seconds(), msg)
		time.Sleep(3 * time.Second)
	}
}
