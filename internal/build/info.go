package build

import (
	"runtime/debug"
	"strings"
)

var (
	Version = version()
	Commit  = commit()
	Date    = date()
)

func version() string {
	if v := strings.TrimSpace(versionFromLdflags); v != "" && v != "dev" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	return "dev"
}

func commit() string {
	if c := strings.TrimSpace(commitFromLdflags); c != "" {
		return c
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) >= 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}
	return ""
}

func date() string {
	if d := strings.TrimSpace(dateFromLdflags); d != "" {
		return d
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return ""
}

var (
	versionFromLdflags = "dev"
	commitFromLdflags  = ""
	dateFromLdflags    = ""
)
