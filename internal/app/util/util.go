package util

import "os"

// Use in testing
func Leave() {
out:
	for i := range 10 {
		if i == 9 {
			panic("i == 9")
		}

		e, _ := os.ReadDir("./")
		for _, f := range e {
			if f.Name() == "go.mod" {
				break out
			}
		}
		if err := os.Chdir("../"); err != nil {
			panic("cannot find project root")
		}
	}
}
