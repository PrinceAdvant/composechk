// Command composechk prints structural warnings for Compose YAML files.
package main

import (
	"fmt"
	"os"

	"github.com/PrinceAdvant/composechk"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: composechk file.yml [file.yml...]")
		os.Exit(2)
	}
	bad := false
	for _, path := range os.Args[1:] {
		r, err := composechk.CheckFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			bad = true
			continue
		}
		fmt.Println(path)
		for _, f := range r.Findings {
			fmt.Printf("  %s: %s\n", f.Level, f.Message)
			if f.Level == "error" {
				bad = true
			}
		}
	}
	if bad {
		os.Exit(1)
	}
}
