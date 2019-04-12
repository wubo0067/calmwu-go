TARGET		= cppclient

CPP_SRCS	:= $(wildcard *.cpp)
CPP_OBJS	:= $(patsubst %.cpp, %.o, $(CPP_SRCS))
TARGET_CPP_OBJS = $(addprefix $(TMP_DIR)/, $(CPP_OBJS))

CC_SRCS	:= $(wildcard *.cc)
CC_OBJS	:= $(patsubst %.cc, %.o, $(CC_SRCS))
TARGET_CC_OBJS = $(addprefix $(TMP_DIR)/, $(CC_OBJS))

TMP_DIR     := obj
SRCS		:= $(CPP_SRCS) $(CC_SRCS)
TARGET_OBJS	= $(TARGET_CPP_OBJS) $(TARGET_CC_OBJS)
LIBS 		= /usr/local/lib/libcryptopp.a /usr/local/lib/libprotobuf.a /usr/local/lib/libgflags.a -lpthread

CPPFLAGS	=
CFLAGS		= -ggdb3 -g3 -Wall -fPIC
LDFLAGS		= 

.PHONY : all clean depend

all : depend $(TARGET)

-include .depend

$(TARGET) : $(TARGET_OBJS)
	$(CXX) $(CPPFLAGS) $(LDFLAGS) -o $(TARGET) $(TARGET_OBJS) $(LIBS)

$(TARGET_CPP_OBJS) : $(TMP_DIR)/%.o : %.cpp
	$(CXX) $(CPPFLAGS) $(CFLAGS) -c $< -o $@	

$(TARGET_CC_OBJS) : $(TMP_DIR)/%.o : %.cc
	$(CXX) $(CPPFLAGS) $(CFLAGS) -c $< -o $@

depend :
	$(CXX) -MM $(CFLAGS) $(CPPFLAGS) $(SRCS) > .depend
	@-(./replace.sh)

clean :
	@-(rm $(TMP_DIR)/*.o ./$(TARGET) .depend)
 