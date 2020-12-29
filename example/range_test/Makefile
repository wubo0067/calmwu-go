# This is how we want to name the binary output
OUTPUT=test
# These are the values we want to pass for Version and BuildTime
#GITTAG=`git describe --tags`
VERSION_TAG=$(tag)
BUILD_TIME=`date +%FT%T%z`
# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.VersionTag=${VERSION_TAG} -X main.BuildTime=${BUILD_TIME}"
all:
	go build ${LDFLAGS} -o ${OUTPUT} test.go
