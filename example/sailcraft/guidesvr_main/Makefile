TARGET      =   guidesvr_main
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ./root/*.go 				\
                                ./guidesvr/*.go 		    \
                                ./common/*.go 				\
                                ./proto/*.go)
VERSION     =   1.0.0
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
#GCFLAGS     =   
BINDIR      =   ./bin
GOTAGS      :=  
ifdef CADEBUG
GOTAGS      +=  -tags "gocql_debug"
endif

.PHONY: clean all

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

$(BINDIR)/% : $(ALLSRCS)
	go build $(GOTAGS) -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	go clean -i 
	-rm $(BINDIR)/$(TARGET)
