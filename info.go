package main

import (
	"os"
	"runtime"
	"runtime/debug"
	"text/template"
)

// Set by LDLFLAGS on build
var (
	Version   string
	Branch    string
	BuildDate string
)

// inspired by https://github.com/prometheus/common/blob/main/version/info.go
var infoTmpl = `{{.program}} version {{.version}} 
    (revision: {{.revision}}, date: {{.buildDate}} go: {{.goVersion}}, platform: {{.platform}})
`

func PrintInfo() {

	info := map[string]string{
		"program":   os.Args[0],
		"version":   Version,
		"branch":    Branch,
		"revision":  "unknown",
		"buildDate": BuildDate,
		"goVersion": runtime.Version(),
		"platform":  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// info["version"] = buildInfo.Main.Version
		dirty := false
		for _, v := range buildInfo.Settings {
			if v.Key == "vcs.revision" {
				info["revision"] = v.Value
			}
			if v.Key == "vcs.modified" {
				if v.Value == "true" {
					dirty = true
				}
			}
		}

		if dirty {
			info["revision"] += "-dirty"
		}
	}

	tmpl, _ := template.New("info").Parse(infoTmpl)
	tmpl.Execute(os.Stdout, info)
	return
}
