.PHONY: dev build-mac build-windows

dev:
	wails dev

build-mac:
	wails build -clean --platform darwin/arm64 -o Pathlive-Net-Tool -upx

build-windows:
	wails build -clean -platform windows/arm64 -o Pathlive-Net-Tool.exe -ldflags "-X 'main.Win=1'"