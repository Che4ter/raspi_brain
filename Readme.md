 export PATH=$PATH:/home/philipp/SourceCodes/golang/bin
go run main.go -webgui



#crosscompile for rpi3
GOARM=7 GOARCH=arm GOOS=linux go build

#embd notice
##warning
embd has a bug
add new chip string in detect.go
strings.Contains(hardware, "BCM2835"