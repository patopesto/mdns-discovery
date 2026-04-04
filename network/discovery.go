package network

import (
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	QUERY_INTERVAL = 11 // in seconds
	QUERY_TIMEOUT  = 10 // in seconds

	// Used to query all peers for their services
	// https://github.com/libp2p/specs/blob/master/discovery/mdns.md#dns-service-discovery
	MDNS_META_QUERY = "_services._dns-sd._udp"
	DEFAULT_DOMAIN  = "local"
)

type ServiceEntry = mdns.ServiceEntry // type alias

// Discovery manages all the DiscoveryServices
type Discovery struct {
	Interfaces []*Interface
	Domains    []string

	services  map[string][]*DiscoveryService
	mu        sync.RWMutex
	EntriesCh chan ServiceEntry // Channel for newly discovered entries
}

func InitDiscovery(ifaces []string, domains []string, entriesCh chan ServiceEntry) *Discovery {

	d := &Discovery{
		Domains:   domains,
		services:  make(map[string][]*DiscoveryService, 0),
		EntriesCh: entriesCh,
	}

	if len(ifaces) == 0 {
		d.Interfaces = GetInterfaces()
	} else {
		d.Interfaces = GetInterfacesByName(ifaces)
	}

	for _, itf := range d.Interfaces {
		for _, domain := range d.Domains {
			service := NewDiscoveryService(MDNS_META_QUERY, domain, itf.Interface, d.EntriesCh)
			d.services[itf.Name] = append(d.services[itf.Name], service)
			service.Start()
		}
	}

	return d
}

// EnableInterface adds an interface to discovery and starts services for it
func (d *Discovery) EnableInterface(iface *Interface) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, itf := range d.Interfaces {
		if itf.Name == iface.Name { // already enabled
			return nil
		}
	}

	d.Interfaces = append(d.Interfaces, iface)

	for _, domain := range d.Domains {
		service := NewDiscoveryService(MDNS_META_QUERY, domain, iface.Interface, d.EntriesCh)
		d.services[iface.Name] = append(d.services[iface.Name], service)
		service.Start()
	}

	return nil
}

// DisableInterface removes an interface from discovery and stops its services
func (d *Discovery) DisableInterface(iface *Interface) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	found := false
	for i, itf := range d.Interfaces {
		if itf.Name == iface.Name {
			d.Interfaces = append(d.Interfaces[:i], d.Interfaces[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	if services, ok := d.services[iface.Name]; ok {
		for _, service := range services {
			service.Stop()
		}
		delete(d.services, iface.Name)
	}

	return nil
}

// IsInterfaceEnabled checks if an interface is currently enabled
func (d *Discovery) IsInterfaceEnabled(name string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, itf := range d.Interfaces {
		if itf.Name == name {
			return true
		}
	}
	return false
}

// A DiscoveryService queries the network for a single domain on a single interface
type DiscoveryService struct {
	Params      *mdns.QueryParam
	Entries     []ServiceEntry
	entriesCh   chan *mdns.ServiceEntry
	timer       *time.Ticker
	stop        chan struct{}
	discoveryCh chan ServiceEntry // Channel to send discovered entries back to Discovery
}

func NewDiscoveryService(service string, domain string, iface *net.Interface, discoveryCh chan ServiceEntry) *DiscoveryService {
	entriesCh := make(chan *mdns.ServiceEntry, 10)
	entries := make([]ServiceEntry, 0)
	return &DiscoveryService{
		Entries:     entries,
		entriesCh:   entriesCh,
		discoveryCh: discoveryCh,
		Params: &mdns.QueryParam{
			Service:             service,
			Domain:              domain,
			Timeout:             QUERY_TIMEOUT * time.Second,
			Entries:             entriesCh,
			Interface:           iface,
			WantUnicastResponse: false,
			DisableIPv4:         false,
			DisableIPv6:         false,
			Logger:              log.New(ioutil.Discard, "", 0),
		},
	}
}

func (d *DiscoveryService) Start() {
	d.timer = time.NewTicker(QUERY_INTERVAL * time.Second)
	d.stop = make(chan struct{})

	go d.Run()
}

func (d *DiscoveryService) Stop() {
	close(d.stop)
}

func (d *DiscoveryService) Run() {
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
		case entry := <-d.entriesCh:
			if entry != nil {
				found := false
				for _, existing := range d.Entries {
					if reflect.DeepEqual(entry, existing) {
						found = true
						break
					}
				}
				if !found {
					d.Entries = append(d.Entries, *entry)
					// Send the new entry through the discovery channel
					d.discoveryCh <- *entry
				}
			}
		}
	}

}
