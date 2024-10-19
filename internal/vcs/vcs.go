package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var (
		time     string
		revision string
	)

	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.time":
				time = s.Value
			case "vcs.revision":
				revision = s.Value
			}
		}
	}

	return fmt.Sprintf("%s-%s", time, revision)
}
