GO  = GO111MODULE=on go
TARGET      =   doyo-recdatasvr
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ./recdatasvr/*.go ./proto/*.go ./doyorecdata/*.go)
VERSION     =   0.0.1
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "all=-N -l"
#GCFLAGS     =   
BINDIR      =   ./bin
GOTAGS      :=  -tags static

.PHONY: clean all vet setup deps

all : vet lint $(TARGET)

$(TARGET) : % : $(BINDIR)/%

$(BINDIR)/% : $(ALLSRCS)
	go build $(GOTAGS) -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	go clean -i 
	-rm $(BINDIR)/$(TARGET)

vet :
	@echo "Running vet..."
	@go vet -shadow -v $(go list ./...| grep -v /test/)

lint :
	@echo "Running golint..."
	@golint ./recdatasvr/... ./doyorecdata/... ./proto/...

setup :
	@go get -u go.uber.org/zap/zapcore
	@go get -u github.com/urfave/cli
	@rm -rf ${GOPATH}/src/golang.org/x/lint && cd ${GOPATH}/src/golang.org/x && git clone https://github.com/golang/lint.git && cd -
	@go get -u golang.org/x/lint/golint
	@go get -u github.com/gin-gonic/gin
	@go get -u github.com/go-redis/redis
	@go get -u github.com/hashicorp/consul/api
	@go get -u github.com/mitchellh/mapstructure
	@go get -u github.com/monnand/dhkx
	@go get -u github.com/mozhata/merr
	@go get -u github.com/pquerna/ffjson/ffjson
	@go get -u github.com/satori/go.uuid
	@go get -u github.com/spaolacci/murmur3
	@go get -u go.uber.org/zap
	@rm -rf ${GOPATH}/src/golang.org/x/time && cd ${GOPATH}/src/golang.org/x && git clone https://github.com/golang/time.git && cd -
	@go get -u gopkg.in/natefinch/lumberjack.v2
	@go get -u github.com/golang/protobuf/proto

deps:
	${GO} mod vendor
	${GO} mod download	