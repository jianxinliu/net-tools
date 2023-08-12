package main

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func Hoplist(host string, maxHops int) []net.IP {
	ips, err := runTraceCmd(host, maxHops)
	if err != nil {
		fmt.Printf("trace router failed, %v", err)
		return []net.IP{}
	}
	return ips

	// ips, err := net.LookupIP(host)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// var dst net.IPAddr
	// for _, ip := range ips {
	// 	if ip.To4() != nil {
	// 		dst.IP = ip
	// 		fmt.Printf("using %v for tracing an IP packet route to %s\n", dst.IP, host)
	// 		break
	// 	}
	// }
	// if dst.IP == nil {
	// 	fmt.Println("no A record found")
	// }

	// c, err := net.ListenPacket("ip4:1", "0.0.0.0") // ICMP for IPv4
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer c.Close()
	// p := ipv4.NewPacketConn(c)

	// if err := p.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
	// 	fmt.Println(err)
	// }
	// wm := icmp.Message{
	// 	Type: ipv4.ICMPTypeEcho, Code: 0,
	// 	Body: &icmp.Echo{
	// 		ID:   os.Getpid() & 0xffff,
	// 		Data: []byte("HELLO-R-U-THERE"),
	// 	},
	// }

	// ret := []net.IP{}
	// rb := make([]byte, 1500)
	// for i := 1; i <= 30; i++ { // up to 64 hops
	// 	wm.Body.(*icmp.Echo).Seq = i
	// 	wb, err := wm.Marshal(nil)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	if err := p.SetTTL(i); err != nil {
	// 		fmt.Println(err)
	// 	}

	// 	// In the real world usually there are several
	// 	// multiple traffic-engineered paths for each hop.
	// 	// You may need to probe a few times to each hop.
	// 	begin := time.Now()
	// 	if _, err := p.WriteTo(wb, nil, &dst); err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	if err := p.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	n, cm, peer, err := p.ReadFrom(rb)
	// 	if err != nil {
	// 		if err, ok := err.(net.Error); ok && err.Timeout() {
	// 			fmt.Printf("%v\t*\n", i)
	// 			continue
	// 		}
	// 		fmt.Println(err)
	// 	}
	// 	rm, err := icmp.ParseMessage(1, rb[:n])
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	rtt := time.Since(begin)

	// 	// In the real world you need to determine whether the
	// 	// received message is yours using ControlMessage.Src,
	// 	// ControlMessage.Dst, icmp.Echo.ID and icmp.Echo.Seq.
	// 	switch rm.Type {
	// 	case ipv4.ICMPTypeTimeExceeded:
	// 		names, _ := net.LookupAddr(peer.String())
	// 		fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
	// 		ret = append(ret, net.ParseIP(peer.String()))
	// 	case ipv4.ICMPTypeEchoReply:
	// 		names, _ := net.LookupAddr(peer.String())
	// 		fmt.Printf("%d\t%v %+v %v\n\t%+v\n", i, peer, names, rtt, cm)
	// 		break
	// 	default:
	// 		fmt.Printf("unknown ICMP message: %+v\n", rm)
	// 	}
	// }

	// return ret
}

const winRtPattern = "(\\d{1,3})\\s.*?((\\d{1,3}\\.){3}\\d{1,3})"
const darwinRtPattern = "(\\d{1,3})\\s.*?((\\d{1,3}\\.){3}\\d{1,3}).*$"

func getTraceInfo(host string, maxHops int) (*exec.Cmd, string) {
	platform := runtime.GOOS
	max := strconv.Itoa(maxHops)
	// no name resolve, max hop is max
	if Global.IsWindows {
		return exec.Command("tracert", "-d", "-h", max, host), winRtPattern
	} else if Global.IsDarwin {
		return exec.Command("traceroute", "-n", "-m", max, host), darwinRtPattern
	} else {
		return nil, fmt.Sprintf("not supported system %s", platform)
	}
}

func runTraceCmd(host string, maxHops int) ([]net.IP, error) {
	cmd, regPattern := getTraceInfo(host, maxHops)

	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("run trace route failed: %w", err)
	}

	return parseTraceRouteResult(buf.String(), regPattern), nil
}

func parseTraceRouteResult(line, pattern string) []net.IP {
	ret := []net.IP{}

	lines := strings.Split(line, "\n")
	for _, l := range lines {
		L := strings.TrimSpace(l)
		if L == "" {
			continue
		}
		reg, _ := regexp.Compile(pattern)
		if reg.Match([]byte(L)) {
			matches := reg.FindAllStringSubmatch(L, -1)
			grp := matches[0]
			_, ip := grp[1], grp[2]
			ret = append(ret, net.ParseIP(ip))
		}
	}

	return ret
}
