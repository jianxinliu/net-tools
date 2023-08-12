package main

import (
	"fmt"
	"testing"
)

func TestDarwinRtPattern(t *testing.T) {

	// traceroute -n -m 6 www.baidu.com
	ret := parseTraceRouteResult(`traceroute: Warning: www.baidu.com has multiple addresses; using 180.101.50.188
	traceroute to www.a.shifen.com (180.101.50.188), 6 hops max, 52 byte packets
	 1  192.168.0.1  36.551 ms  4.195 ms  16.766 ms
	 2  192.168.1.1  7.840 ms  12.329 ms  3.741 ms
	 3  222.67.192.1  11.558 ms  10.115 ms  11.890 ms
	 4  61.152.53.129  17.263 ms  7.873 ms  19.598 ms
	 5  101.95.88.134  8.995 ms
		61.152.25.14  11.266 ms
		61.152.26.26  10.508 ms
	 6  202.97.29.106  14.124 ms
		202.97.52.62  12.223 ms
		202.97.29.106  14.933 ms`, darwinRtPattern)

	for _, ip := range ret {
		fmt.Println(ip)
	}
}

func TestWinRtPattern(t *testing.T) {

	// tracert -d www.baidu.com
	ret := parseTraceRouteResult(`
	Tracing route to www.a.shifen.com [180.101.50.242]
over a maximum of 30 hops:

  1    14 ms    13 ms    13 ms  180.101.50.242
  2    14 ms    13 ms    13 ms  110.101.50.242
  3    14 ms    13 ms    13 ms  1.101.5.242
  4    14 ms    13 ms    13 ms  18.101.50.242

Trace complete.
	`, winRtPattern)

	for _, ip := range ret {
		fmt.Println(ip)
	}
}
