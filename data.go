package main

import (
	"net"
)

// Fake data to populate table

var fakeData = []ServiceEntry{
	{
		Name: "Asterix._device_info._tcp._local.",
		Host: "Asterix.local.",
		AddrV4: net.IPv4(192,168,1,1),
		Port: 9356,
		Info: "abc=def|123=456",
	},
	{
		Name: "Obelix._device_info._tcp._local.",
		Host: "Obelix.local.",
		AddrV4: net.IPv4(192,168,1,145),
		Port: 876,
		Info: "abc=def|123=456",
	},
	{
		Name: "Obelix._ssh._tcp._local.",
		Host: "Obelix.local.",
		AddrV4: net.IPv4(192,168,1,145),
		Port: 22,
		Info: "encrypted",
	},
	{
		Name: "Idefix._esphomelib._tcp._local.",
		Host: "Idefix.local.",
		AddrV4: net.IPv4(192,168,1,34),
		Port: 6543,
		Info: "esphome=true|esp=c3|idf_version=v5.1.1",
	},
	{
		Name: "Panoramix._esphomelib._tcp._local.",
		Host: "Panoramix.local.",
		AddrV4: net.IPv4(192,168,1,34),
		Port: 6543,
		Info: "esphome=true|esp=wroom|idf_version=v5.2.0",
	},
	{
		Name: "Abraracourcix._printer._tcp._local.",
		Host: "Abraracourcix.local.",
		AddrV4: net.IPv4(192,168,1,13),
		Port: 1,
		Info: "printer_is_dead=true",
	},
	{
		Name: "Falbala._smb._tcp._local.",
		Host: "Falbala.local.",
		AddrV4: net.IPv4(192,168,1,14),
		Port: 445,
		Info: "",
	},
	{
		Name: "Assurancetourix._airplay._tcp._local.",
		Host: "Assurancetourix.local.",
		AddrV4: net.IPv4(192,168,1,254),
		Port: 7000,
		Info: "now_plating=TheBlaze-Territory",
	},
}
