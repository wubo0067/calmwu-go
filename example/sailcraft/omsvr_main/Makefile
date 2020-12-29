TARGET      =   omsvr_main
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ./root/*.go \
                                ./web/*.go 	\
                                ./activemgr/*.go \
                                ./common/*.go \
                                ./proto/*go \
                                ./db/*.go )
VERSION     =   1.0.0
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
#GCFLAGS     =   
BINDIR      =   ./bin

.PHONY: clean all

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

$(BINDIR)/% : $(ALLSRCS)
	go build $(GOTAGS) -i -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	go clean -i -cache
	-rm $(BINDIR)/$(TARGET)
