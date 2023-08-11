package main

import (
	"fmt"
	"net"
)

func Mtr(host string) error {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) < 1 {
		return fmt.Errorf("host [%s] dns resolve failed", host)
	}
	ip := ips[0]

	fmt.Printf("ip: %s", ip)
	// hops, err := traceroute.Trace(ip)

	// if err != nil {
	// 	return err
	// }
	// for i := 0; i < len(hops); i++ {
	// 	fmt.Printf("hop: dis: %d %v \n", hops[i].Distance, hops[i].Nodes)
	// }
	return nil
}
