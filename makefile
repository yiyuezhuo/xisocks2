# use git bash to run make in Windows, requiring 7z command in path

all: build-windows build-linux build-android

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o bin/linux/client github.com/yiyuezhuo/xisocks2/client
	GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o bin/linux/server github.com/yiyuezhuo/xisocks2/server
	cp server/config-server.json bin/linux/
	cp client/config-client.json bin/linux/
	7z a -tzip bin/linux.zip bin/linux

build-android:
	GOOS=linux GOARCH=arm64 go build -ldflags "-w" -o bin/android/client github.com/yiyuezhuo/xisocks2/client
	GOOS=linux GOARCH=arm64 go build -ldflags "-w" -o bin/android/server github.com/yiyuezhuo/xisocks2/server
	cp server/config-server.json bin/android/
	cp client/config-client.json bin/android/
	7z a -tzip bin/android.zip bin/android

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-w" -o bin/windows/client.exe github.com/yiyuezhuo/xisocks2/client
	GOOS=windows GOARCH=amd64 go build -ldflags "-w" -o bin/windows/server.exe github.com/yiyuezhuo/xisocks2/server
	cp server/config-server.json bin/windows/
	cp client/config-client.json bin/windows/
	7z a -tzip bin/windows.zip bin/windows

clean:
	rm bin/*.zip
	rm bin/windows/*
	rm bin/linux/*
	rm bin/android/*