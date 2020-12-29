TARGET      =   doyo-routersvr
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ./routersvr/*.go ./proto/*.go)
VERSION     =   0.0.1
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
#GCFLAGS     =   
BINDIR      =   ./bin
GOTAGS      :=  -tags static

.PHONY: clean all

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

$(BINDIR)/% : $(ALLSRCS)
	go build -mod vendor $(GOTAGS) -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS) 
	#这里使用go mod的vendor模式，在上级目录中已经有了vendor

clean :
	go clean -i 
	-rm $(BINDIR)/$(TARGET)