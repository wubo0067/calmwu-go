TARGET      =   kp_agent
SRCS        =   $(wildcard *.go)
ALLSRCS     =   $(wildcard *.go ../kp_proto/*.go)
VERSION     =   1.0.0
BUILDTIME   =   `date +%FT%T%z`
LDFLAGS     =   -ldflags "-X main.kpagent_version=${VERSION} -X main.kpagent_buildtime=${BUILDTIME}"
GCFLAGS     =   -gcflags "-N -l"
BINDIR      =   ../../bin

all : $(TARGET)

$(TARGET) : % : $(BINDIR)/%

#加上$(ALLSRCS)才能感知代码的变化，重新编译
$(BINDIR)/% : $(ALLSRCS)
	go build $(GCFLAGS) $(LDFLAGS) -o $@ $(SRCS)

clean :
	rm $(BINDIR)/$(TARGET)
