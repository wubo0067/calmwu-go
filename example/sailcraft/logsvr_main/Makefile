TARGET      =   logsvr_main
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ./logsvr/*.go 		\
                                ./proto/*.go)
VERSION     =   1.0.0
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
#GCFLAGS     =   
BINDIR      =   ./bin
GOTAGS      :=  

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

#加上$(ALLSRCS)才能感知代码的变化，重新编译
$(BINDIR)/% : $(ALLSRCS)
	go build $(GOTAGS) -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	go clean -i 
	-rm $(BINDIR)/$(TARGET)
