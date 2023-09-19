package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

var Win = "0"

type PacketInfo struct {
	probing.Packet
	TimeStr string
	Dup     bool
}

func NewPacketInfo(pkt *probing.Packet) *PacketInfo {
	return &PacketInfo{
		Packet:  *pkt,
		TimeStr: time.Now().Format("01-02 15:04:05"),
	}
}

var pinger *probing.Pinger

func Ping(ctx context.Context, count int, intervalMills time.Duration, dest string) string {
	p, err := probing.NewPinger(dest)
	if err != nil {
		rt.MessageDialog(ctx, rt.MessageDialogOptions{
			Message: err.Error(),
			Title:   "出错啦",
			Type:    rt.ErrorDialog,
		})
		return ""
	}

	pinger = p

	go func() {
		<-ctx.Done()
		fmt.Println("app stop....")
	}()

	if Global.IsWindows {
		pinger.SetPrivileged(true)
	}

	rt.EventsOn(ctx, "PING_STOP", func(optionData ...any) {
		fmt.Println("stop ping from frontend")
		pinger.Stop()
	})

	if count > 1000 {
		count = 1000
	}
	if count > 0 {
		// 指定 ping 的次数，指定负数表示一直 ping
		pinger.Count = count
	}
	pinger.Interval = time.Millisecond * intervalMills

	pinger.OnRecv = func(pkt *probing.Packet) {
		bytes, _ := json.Marshal(NewPacketInfo(pkt))
		emitPing(ctx, string(bytes))
	}

	pinger.OnDuplicateRecv = func(pkt *probing.Packet) {
		pf := NewPacketInfo(pkt)
		pf.Dup = true
		bytes, _ := json.Marshal(pf)
		emitPing(ctx, string(bytes))
	}

	pinger.OnFinish = func(stats *probing.Statistics) {
		bytes, _ := json.Marshal(stats)
		rt.EventsEmit(ctx, "PING_STAT", string(bytes))
	}

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	err = pinger.RunWithContext(ctx)
	if err != nil {
		panic(err)
	}
	return dest
}

func emitPing(ctx context.Context, str ...string) {
	for i := 0; i < len(str); i++ {
		rt.EventsEmit(ctx, "PING", str[i])
	}
}
