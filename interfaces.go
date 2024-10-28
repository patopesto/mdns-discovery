package main

import (
	"log"
	"net"
	"slices"
)

func GetInterfaces() []*net.Interface {

	itfs := make([]*net.Interface, 0)

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, iface := range ifaces {
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
			switch a.(type) {
			case *net.IPNet:
				if !slices.Contains(itfs, &iface) {
					itfs = append(itfs, &iface)
				}
			}
		}
	}

	// log.Println("Interfaces found:")
	// for _, itf := range itfs {
	//     log.Println(itf)
	// }

	return itfs
}

func GetInterfacesByName(ifaces []string) []*net.Interface {

	itfs := make([]*net.Interface, 0)

	for _, iface := range ifaces {
		itf, err := net.InterfaceByName(iface)
		if err != nil {
			continue
		}
		itfs = append(itfs, itf)
	}

	// log.Println("Interfaces found:")
	// for _, itf := range itfs {
	//     log.Println(itf)
	// }

	return itfs
}
