package main

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	flag "github.com/spf13/pflag"

	"gitlab.com/patopest/mdns-discovery/app"
	"gitlab.com/patopest/mdns-discovery/network"
)

var ifaces = flag.StringSliceP("interface", "i", nil, "Use specified interface(s). ex: '-i eth0,wlan0' (default: all available interfaces)")
var doms = flag.StringSliceP("domain", "d", []string{network.DEFAULT_DOMAIN}, "Domain(s) to use, usually '.local' \t\t!!! Do not change unless you know what you're doing !!!")
var info = flag.BoolP("version", "v", false, "Print version info")
var usage = flag.BoolP("help", "h", false, "Print this help message")
var debugFile = flag.Bool("debug", false, "Write logs to file")
var fake = flag.Bool("fake", false, "Use fake data instead")

func main() {
	flag.CommandLine.SortFlags = false
	flag.CommandLine.MarkHidden("debug")
	flag.CommandLine.MarkHidden("fake")
	flag.Parse()

	if *debugFile {
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

	if *info {
		PrintInfo()
		os.Exit(0)
	}
	if *usage {
		flag.Usage()
		os.Exit(0)
	}

	var m *app.App

	if *fake {
		m = app.NewApp(*ifaces, []string{"test.com"})
		m.InjectFakeData(network.FakeDataLong)
		// m.InjectFakeData(network.FakeData)
	} else {
		m = app.NewApp(*ifaces, *doms)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
