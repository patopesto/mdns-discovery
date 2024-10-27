package main

import (
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	QUERY_INTERVAL = 11 // in seconds
	QUERY_TIMEOUT  = 10 // in seconds

	// Used to query all peers for their services
	// https://github.com/libp2p/specs/blob/master/discovery/mdns.md#dns-service-discovery
	MDNS_META_QUERY = "_services._dns-sd._udp"
)

type ServiceEntry = mdns.ServiceEntry // type alias

type Discovery struct {
	Params    *mdns.QueryParam
	Entries   []ServiceEntry
	entriesCh chan *mdns.ServiceEntry
	timer     *time.Ticker
	stop      chan struct{}
}

var discoveries []*Discovery
var interfaces []*net.Interface
var discoveryLogger *log.Logger

func InitDiscovery() {
	discoveries = make([]*Discovery, 0)
	discoveryLogger = log.New(ioutil.Discard, "", 0)

	interfaces = GetInterfaces()
	for _, itf := range interfaces {
		discovery := NewDiscovery(MDNS_META_QUERY, itf)
		discoveries = append(discoveries, discovery)
		discovery.Start()
	}
}

func NewDiscovery(service string, iface *net.Interface) *Discovery {
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	entries := make([]ServiceEntry, 0)
	return &Discovery{
		Entries:   entries,
		entriesCh: entriesCh,
		Params: &mdns.QueryParam{
			Service:             service,
			Domain:              "local",
			Timeout:             QUERY_TIMEOUT * time.Second,
			Entries:             entriesCh,
			Interface:           iface,
			WantUnicastResponse: false,
			DisableIPv4:         false,
			DisableIPv6:         false,
			Logger:              discoveryLogger,
		},
	}
}

func (d *Discovery) Start() {
	d.timer = time.NewTicker(QUERY_INTERVAL * time.Second)
	d.stop = make(chan struct{})

	go d.Run()
}

func (d *Discovery) Stop() {
	close(d.stop)
}

func (d *Discovery) Run() {
	defer d.timer.Stop()

	// Running the queries at interval in it's own goroutine
	go func() {
		mdns.Query(d.Params)
		for {
			select {
			case <-d.stop:
				return
			case <-d.timer.C:
				mdns.Query(d.Params)
			}
		}
	}()

	for {
		select {
		case <-d.stop:
			return
		case <-d.entriesCh:
			for entry := range d.entriesCh {
				found := false
				for _, existing := range d.Entries {
					if reflect.DeepEqual(entry, existing) {
						found = true
						break
					}
				}
				if !found {
					d.Entries = append(d.Entries, *entry)
				}
			}
		}
	}

}
