build_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o packing/linux_amd64/devserver .
	rm -f packing/distro/devserver_linux_amd64.zip
	zip -r packing/distro/devserver_linux_amd64.zip packing/linux_amd64
build_linux_arm64:
	GOOS=linux GOARCH=arm64 go build -o packing/linux_arm64/devserver .
	rm -f packing/distro/devserver_linux_arm64.zip
	zip -r packing/distro/devserver_linux_arm64.zip packing/linux_arm64	
build_windows_amd64:
	GOOS=windows GOARCH=amd64 go build -o packing/windows_amd64/devserver.exe .
	rm -f packing/distro/devserver_windows_amd64.zip
	zip -r packing/distro/devserver_windows_amd64.zip packing/windows_amd64
build_darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o packing/_mac/DevServer.app/Contents/MacOS/devserver .
	rm -f packing/distro/devserver_darwin_amd64.zip
	zip -r packing/distro/devserver_darwin_amd64.zip packing/_mac/DevServer.app
	rm packing/_mac/DevServer.app/Contents/MacOS/devserver
build_darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -o packing/_mac/DevServer.app/Contents/MacOS/devserver .
	rm -f packing/distro/devserver_darwin_arm.zip
	zip -r packing/distro/devserver_darwin_arm64.zip packing/_mac/DevServer.app
	rm packing/_mac/DevServer.app/Contents/MacOS/devserver	
build_all: build_windows_amd64