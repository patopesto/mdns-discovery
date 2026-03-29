package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"runtime"
	"runtime/debug"
	"text/template"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/fang/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gitlab.com/patopest/mdns-discovery/app"
	"gitlab.com/patopest/mdns-discovery/network"
)

// Set by LDLFLAGS on build
var (
	Version   string
	Branch    string
	BuildDate string
)

// inspired by https://github.com/prometheus/common/blob/main/version/info.go
var versionTmpl = `{{.program}} version {{.version}} 
    (revision: {{.revision}}, date: {{.buildDate}}, go: {{.goVersion}}, platform: {{.platform}})
`

func GetVersion() string {
	info := map[string]string{
		"program":   filepath.Base(os.Args[0]),
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

	var buf strings.Builder
	tmpl, _ := template.New("info").Parse(versionTmpl)
	tmpl.Execute(&buf, info)
	return buf.String()
}


func main() {

	var cmd = &cobra.Command{
		Use:   "mdns-discovery",
		Short: "A TUI for discovering mDNS services",
		Run: func(cmd *cobra.Command, args []string) {

			if viper.GetBool("debug") {
				f, err := tea.LogToFile("debug.log", "")
				if err != nil {
					log.Fatal("fatal:", err)
					os.Exit(1)
				}
				defer f.Close()
			} else {
				log.SetOutput(io.Discard)
			}

			log.Println("Hello! Starting up...")

			var m *app.App
			itfs := viper.GetStringSlice("interface")
			domains := viper.GetStringSlice("domain")

			if viper.GetBool("fake") {
				m = app.NewApp(itfs, []string{"test.com"})
				m.InjectFakeData(network.FakeDataLong)
				// m.InjectFakeData(network.FakeData)
			} else {
				m = app.NewApp(itfs, domains)
			}

			p := tea.NewProgram(m)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
		},
	}

	var ifaces []string
	var domain []string
	var debugFile bool
	var fake bool

	cmd.Flags().StringSliceVarP(&ifaces, "interface", "i", nil, "Use specified interface(s). ex: '-i eth0,wlan0' (default: all available interfaces)")
	cmd.Flags().StringSliceVarP(&domain, "domain", "d", []string{network.DEFAULT_DOMAIN}, "Domain(s) to use, usually '.local' !!! Do not change unless you know what you're doing !!!")
	cmd.Flags().BoolVarP(&debugFile, "debug", "", false, "Write logs to file")
	cmd.Flags().BoolVarP(&fake, "fake", "", false, "Use fake data instead")

	cmd.Flags().MarkHidden("debug")
	cmd.Flags().MarkHidden("fake")

	cmd.SetVersionTemplate(GetVersion())

	// env variable bindings
	viper.BindPFlags(cmd.Flags())
	viper.SetEnvPrefix("mdns")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}
