TARGET      =   calcPreserve
SRCS        =   $(wildcard *.go)
VERSION     =   1.0.0
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
#GCFLAGS     =   
BINDIR      =   ../../bin

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

#加上$(ALLSRCS)才能感知代码的变化，重新编译
$(BINDIR)/% : $(SRCS)
	go build -v -x $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	go clean -i 
	rm $(BINDIR)/$(TARGET)