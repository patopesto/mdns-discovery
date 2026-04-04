package network

import (
	"log"
	"net"
	"slices"
)

// Custom network Interface overlay
type Interface struct {
	*net.Interface
	IPv4 net.IP
}

func GetInterfaces() []*Interface {

	itfs := make([]*Interface, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return nil
	}

	for i := range ifaces {
		iface := &ifaces[i]
		switch {
		case iface.Flags&net.FlagUp == 0:
			continue // Ignore interfaces that are down
		case iface.Flags&net.FlagLoopback != 0:
			continue // Ignore loopback interfaces
		case iface.Flags&net.FlagMulticast == 0:
			continue // Ignore non-multicast interfaces
		case iface.Flags&net.FlagPointToPoint != 0:
			continue // Ignore point-to-point interfaces
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}

		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					itf := &Interface{
						Interface: iface,
						IPv4:      v.IP,
					}
					if !slices.Contains(itfs, itf) {
						itfs = append(itfs, itf)
					}
				}
			}
		}
	}

	log.Println("Interfaces found:")
	for _, itf := range itfs {
		log.Println(itf)
	}

	return itfs
}

func GetInterfacesByName(ifaces []string) []*Interface {

	itfs := make([]*Interface, 0)

	for _, iface := range ifaces {
		itf, err := net.InterfaceByName(iface)
		if err != nil {
			continue
		}
		// Get the IPv4 address
		addrs, err := itf.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}
		var ipv4 net.IP
		for _, a := range addrs {
			if v, ok := a.(*net.IPNet); ok {
				if v.IP.To4() != nil {
					ipv4 = v.IP
					break
				}
			}
		}
		itfs = append(itfs, &Interface{
			Interface: itf,
			IPv4:      ipv4,
		})
	}

	log.Println("Interfaces found:")
	for _, itf := range itfs {
		log.Println(itf)
	}

	return itfs
}
