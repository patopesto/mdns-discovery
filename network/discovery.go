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
	Interfaces []*net.Interface
	Domains    []string

	services   map[string][]*DiscoveryService
	servicesMu sync.RWMutex
	EntriesCh  chan ServiceEntry // Channel for newly discovered entries
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
			service := NewDiscoveryService(MDNS_META_QUERY, domain, itf, d.EntriesCh)
			d.services[itf.Name] = append(d.services[itf.Name], service)
			service.Start()
		}
	}

	return d
}

// EnableInterface adds an interface to discovery and starts services for it
func (d *Discovery) EnableInterface(iface *net.Interface) error {
	d.servicesMu.Lock()
	defer d.servicesMu.Unlock()

	// Check if already enabled
	for _, itf := range d.Interfaces {
		if itf.Name == iface.Name {
			return nil
		}
	}

	d.Interfaces = append(d.Interfaces, iface)

	for _, domain := range d.Domains {
		service := NewDiscoveryService(MDNS_META_QUERY, domain, iface, d.EntriesCh)
		d.services[iface.Name] = append(d.services[iface.Name], service)
		service.Start()
	}

	return nil
}

// DisableInterface removes an interface from discovery and stops its services
func (d *Discovery) DisableInterface(ifaceName string) error {
	d.servicesMu.Lock()
	defer d.servicesMu.Unlock()

	// Find and remove from Interfaces slice
	found := false
	for i, itf := range d.Interfaces {
		if itf.Name == ifaceName {
			d.Interfaces = append(d.Interfaces[:i], d.Interfaces[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	// Stop and remove services for this interface
	if services, ok := d.services[ifaceName]; ok {
		for _, service := range services {
			service.Stop()
		}
		delete(d.services, ifaceName)
	}

	return nil
}

// IsInterfaceEnabled checks if an interface is currently enabled
func (d *Discovery) IsInterfaceEnabled(ifaceName string) bool {
	d.servicesMu.RLock()
	defer d.servicesMu.RUnlock()

	for _, itf := range d.Interfaces {
		if itf.Name == ifaceName {
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
