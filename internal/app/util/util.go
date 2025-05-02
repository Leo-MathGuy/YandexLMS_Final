package util

import "os"

// Use in testing
func Leave() {
out:
	for i := range 5 {
		if i == 4 {
			panic("i == 4")
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
