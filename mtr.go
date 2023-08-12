package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

type MtrRow struct {
	prob *probing.Pinger
	IP   string
	Loss float32
	Min  int64
	Max  int64
	Last int64

	Sent int64
	Recv int64
}

func Mtr(ctx context.Context, host string, count int, interval time.Duration) {
	hops := Hoplist(host)

	fmt.Println("trace start.....")

	localIps := []net.IP{}
	for i := 0; i < len(hops); i++ {
		// if ip.IsPrivate() {
		localIps = append(localIps, hops[i])
		// }
	}

	breakSignal := make(chan struct{}, 1)
	go func() {
		<-ctx.Done()
		fmt.Println("app stop....")
		breakSignal <- struct{}{}
	}()

	rt.EventsOn(ctx, "MTR_STOP", func(optionData ...any) {
		fmt.Println("stop ping from frontend")
		breakSignal <- struct{}{}
	})

	fmt.Printf("locals ip len %d \n", len(localIps))

	mtrTable := make(map[string]*MtrRow)
	for i := 0; i < len(localIps); i++ {
		ip := localIps[i].String()
		row := &MtrRow{IP: ip}
		row.Min = math.MaxInt64

		prob, err := probing.NewPinger(ip)
		if err != nil {
			continue
		}

		if count > 0 {
			prob.Count = count
		}
		if Global.IsWindows {
			prob.SetPrivileged(true)
		}
		prob.Interval = time.Millisecond * interval
		prob.OnSend = func(p *probing.Packet) {
			row.Sent++
		}
		prob.OnRecv = func(p *probing.Packet) {
			rtt := p.Rtt.Milliseconds()

			row.Recv++

			row.Last = rtt
			if rtt > row.Max {
				row.Max = rtt
			}
			if rtt < row.Min {
				row.Min = rtt
			}
			row.Loss = (float32(row.Sent-row.Recv) / float32(row.Sent)) * 100

			// fmt.Printf("f: rtt %d %v \n", rtt, row)

			bytes, _ := json.Marshal(row)
			rt.EventsEmit(ctx, "MTR", string(bytes))
		}

		// // test
		// prob.OnFinish = func(s *probing.Statistics) {
		// 	bytes, _ := json.Marshal(s)
		// 	fmt.Println(string(bytes))
		// }

		row.prob = prob

		mtrTable[ip] = row
	}

	go func() {
		<-breakSignal
		for _, row := range mtrTable {
			row.prob.Stop()
		}
	}()

	bytes, _ := json.Marshal(toTable(localIps, &mtrTable))
	rt.EventsEmit(ctx, "MTR_INIT", string(bytes))

	fmt.Println("prob init done.....")

	for _, row := range mtrTable {
		// 可能有某些 ip 是超时的，需要分开探测
		go func(r *MtrRow) {
			r.prob.Run()
		}(row)
	}
}

func toTable(localIps []net.IP, table *map[string]*MtrRow) []MtrRow {
	ret := []MtrRow{}
	for i := 0; i < len(localIps); i++ {
		ip := localIps[i].String()
		row, ok := (*table)[ip]
		if ok {
			ret = append(ret, *row)
		} else {
			ret = append(ret, MtrRow{IP: ip})
		}
	}
	return ret
}
