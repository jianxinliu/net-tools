.PHONY: dev build-mac build-windows

dev:
	wails dev

build-mac:
	wails build -clean --platform darwin/arm64 -o Pathlive-Net-Tool

build-windows-arm64:
	wails build -clean -platform windows/arm64 -o Pathlive-Net-Tool-arm64.exe

build-windows-x86:
	wails build -clean -platform windows/386 -o Pathlive-Net-Tool-x86.exe

build-windows-amd64:
	wails build -clean -platform windows/amd64 -o Pathlive-Net-Tool-amd64.exe