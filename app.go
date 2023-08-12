package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

type Env struct {
	IsWindows bool
	IsDarwin  bool
}

var Global = Env{}

// NewApp creates a new App application struct
func NewApp() *App {
	if runtime.GOOS == "windows" {
		Global.IsWindows = true
	}
	if runtime.GOOS == "darwin" {
		Global.IsDarwin = true
	}
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) beforeStop(ctx context.Context) bool {

	return false
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) Ping(count int, interval int, dest string) string {
	return Ping(a.ctx, count, time.Duration(interval), dest)
}

func (a *App) Mtr(count int, interval int, dest string) {
	Mtr(a.ctx, dest, count, time.Duration(interval))
}
