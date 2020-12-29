#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <errno.h>
#include <net/if.h>
#include <arpa/inet.h>

#include <string>
#include <sstream>

#include <cryptopp/osrng.h>
#include <cryptopp/osrng.h>
#include <cryptopp/integer.h>
#include <cryptopp/nbtheory.h>
#include <cryptopp/dh.h>
#include <cryptopp/secblock.h>

#include <gflags/gflags.h>

#include "proto_hs.pb.h"

using CryptoPP::AutoSeededRandomPool;
using CryptoPP::DH;
using CryptoPP::Integer;
using CryptoPP::ModularExponentiation;
using CryptoPP::SecByteBlock;

DH  gDh;
Integer p( "0x8b79f180cbd3f282de92e8b8f2d092674ffda61f01ed961f8ef04a1b7a3709ff748c2abf6226cf0c4538e48838193da456e92ee53"
           "0ef7aa703e741585e475b26cd64fa97819181cef27de2449cd385c49c9b030f89873b5b7eaf063a788f00db3cb670c73846bc4f76af"
           "062d672bde8f29806b81548411ab48b99aebfd9c2d09" );
Integer g( "0x029843c81d0ea285c41a49b1a2f8e11a56a4b39040dfbc5ec040150c16f72f874152f9c44c659d86f7717b2425b62597e9a453b13"
           "da327a31cde2cced600915252d30262d1e54f4f864ace0e484f98abdbb37ebb0ba4106af5f0935b744677fa2f7f3826dcef3a158695"
           "6105ebea805d871f34c46c25bc30fc66b2db26cb0a93" );

DEFINE_string( svr_ip, "", "server listen ip address ");
DEFINE_int32( svr_port, 0, "server listen port, default 8003");

#pragma pack(push, 1)
typedef struct NetHead_s {
    uint32_t magicValue;
    uint32_t msgID;
    int16_t  cmd;
    int32_t  dataLen;
}NetHead_t;
#pragma pack(pop)

static uint32_t gMsgID = 0;

static bool ValidatePort(const char* flagname, int32_t value) {
   if (value > 0 && value < 32768)   // value is ok
     return true;
   printf("Invalid value for --%s: port(%d)\n", flagname, (int)value);
   return false;
}

static bool ValidateAddr(const char * flagname, const std::string & ip) {
    int32_t ret = inet_addr(ip.c_str());
    if( INADDR_NONE != ret )
        return true;
    printf("Invalid value for --%s: ip[%s]\n", flagname, ip.c_str());
    return false;
}

DEFINE_validator( svr_port, &ValidatePort );
DEFINE_validator( svr_ip, &ValidateAddr );

static int32_t connect_serv() {

    struct sockaddr_in svr_addr;
    svr_addr.sin_family = AF_INET;
    svr_addr.sin_port = htons( FLAGS_svr_port );
    svr_addr.sin_addr.s_addr = inet_addr( FLAGS_svr_ip.c_str() );

    int32_t sock_fd = socket( AF_INET, SOCK_STREAM, 0 );
    if( sock_fd < 0 ) {
        fprintf(stderr, "socket failed! errno[%d] reason[%s]\n", errno, strerror(errno));
        return -1;
    }

    int32_t ret = connect( sock_fd, (struct sockaddr*)&svr_addr, sizeof(svr_addr) );
    if( ret < 0 ) {
        fprintf(stderr, "connect failed! errno[%d] reason[%s]\n", errno, strerror(errno));
        return -2;
    }

    return sock_fd;
}

static int32_t read_byte4( const char * netBuf, uint32_t & n )
{
#if _BYTE_ORDER == _LITTLE_ENDIAN
    // 本地是小端
    uint32_t b1, b2, b3, b4;
    b1 = (uint8_t)netBuf[0];
    b2 = (uint8_t)netBuf[1];
    b3 = (uint8_t)netBuf[2];
    b4 = (uint8_t)netBuf[3];
    n = ( b1 << 24 | b2 << 16 | b3 << 8 | b4 );

#elif _BYTE_ORDER == _BIG_ENDIAN
    // 不用改变
    uint32_t b1, b2, b3, b4;
    b1 = (uint8_t)netBuf[0];
    b2 = (uint8_t)netBuf[1];
    b3 = (uint8_t)netBuf[2];
    b4 = (uint8_t)netBuf[3];
    n = ( b4 << 24 | b3 << 16 | b2 << 8 | b1 );

#endif
    return 0;
}

static int32_t read_byte2( const char * netBuf, uint16_t & n )
{
#if _BYTE_ORDER == _LITTLE_ENDIAN
    // 网络字节序转为主机字节序
    uint16_t b1, b2;
    b1 = (uint8_t)netBuf[0];
    b2 = (uint8_t)netBuf[1];
    n = ( b1 << 8 | b2 );

#elif _BYTE_ORDER == _BIG_ENDIAN
    // 本机就是网络字节序，不用改变
    uint16_t b1, b2;
    b1 = (uint8_t)netBuf[0];
    b2 = (uint8_t)netBuf[1];
    n = ( b1 | b2 << 8 );

#endif
    return 0;
}

static int32_t readFromNet(int32_t sockfd, NetHead_t & head, ::google::protobuf::Message * msg) {
    char netHeadBuf[14] = {0};

    int32_t ret = read(sockfd, netHeadBuf, 14);
    if(ret < 0) {
        fprintf(stderr, "readFromNet NetHeadBuf Failed! errno[%d] reason[%s]\n", errno, strerror(errno));
        return ret;        
    }

    read_byte4(netHeadBuf, head.magicValue);
    read_byte4(netHeadBuf+4, head.msgID);

    uint16_t cmd;
    read_byte2(netHeadBuf+8, cmd);
    head.cmd = cmd;

    uint32_t dataLen;
    read_byte4(netHeadBuf+10, dataLen);
    head.dataLen = dataLen;

    std::cout << "magicValue:" << head.magicValue << std::endl;
    std::cout << "msgID:" << head.msgID << std::endl;
    std::cout << "cmd:" << head.cmd << std::endl;
    std::cout << "dataLen:" << head.dataLen << std::endl;

    char * netPayloadBuf = (char*)malloc(head.dataLen);

    ret = read(sockfd, netPayloadBuf, head.dataLen);
    if(ret < 0) {
        fprintf(stderr, "readFromNet netPayloadBuf Failed! errno[%d] reason[%s]\n", errno, strerror(errno));
        free(netPayloadBuf);
        return ret;        
    }
    fprintf(stdout, "readFromNet Successed! bytes[%d]\n", ret);

    msg->ParseFromArray(netPayloadBuf, head.dataLen);

    free(netPayloadBuf);
    return ret;
}

static int32_t sendToNet(int32_t sockfd, int16_t cmd, ::std::string & serial_buf) {
    NetHead_t head;
    head.magicValue = htonl( uint32_t(0x98651210) );
    head.cmd = htons(cmd);
    head.msgID = htonl(gMsgID++);
    head.dataLen = htonl(serial_buf.size());

    int32_t netBufSize = 14 + serial_buf.size();
    char * netBuf = (char*)malloc(netBufSize);
    memcpy(netBuf, &head, 14);
    memcpy(netBuf + 14, serial_buf.c_str(), serial_buf.size());

    int32_t ret = write(sockfd, netBuf, netBufSize);
    if( ret < 0 ) {
        fprintf(stderr, "sendToNet Failed! errno[%d] reason[%s]\n", errno, strerror(errno));
        free(netBuf);
        return ret;        
    }
    free(netBuf);
    fprintf(stdout, "sendToNet Successed! bytes[%d]\n", ret);
    return ret;
}

static int32_t sendSyn(int32_t sockfd, const SecByteBlock & pubA) {
    ::protocol::ProtoSyn synMsg;
    synMsg.set_verifybuf("Hello DH");
    synMsg.set_dhclientpubkey(pubA.BytePtr(), pubA.SizeInBytes());

    ::std::string serial_buf;
    synMsg.SerializeToString( &serial_buf );

    return sendToNet(sockfd, 0, serial_buf);
}

int32_t main(int32_t argc, char * argv[]) {

    std::string usage( "Sample usage:\n" );
    usage += std::string( argv[0] ) + " --svr_ip=192.168.2.109 --svr_port=8008";

    ::google::SetUsageMessage( usage );
    ::google::ParseCommandLineFlags( &argc, &argv, true );

    int32_t sockfd = connect_serv();
    if( sockfd < 0 ) {
        exit( -1 );
    }

    AutoSeededRandomPool rnd;
    gDh.AccessGroupParameters().Initialize( p, g );
    SecByteBlock privA( gDh.PrivateKeyLength() );
    SecByteBlock pubA( gDh.PublicKeyLength() ); 
    gDh.GenerateKeyPair( rnd, privA, pubA );

    int32_t ret = sendSyn(sockfd, pubA);
    if(ret < 0) {
        close(sockfd);
        exit( -1 );        
    }

    NetHead_t head;
    ::protocol::ProtoAsyn asynMsg;
    ret = readFromNet(sockfd, head, &asynMsg);
    if(ret < 0) {
        close(sockfd);
        exit( -1 );        
    }

    SecByteBlock secretKeyA( gDh.AgreedValueLength() );
    SecByteBlock pubB((const unsigned char*)asynMsg.dhserverpubkey().c_str(),
        asynMsg.dhserverpubkey().size());

    if ( !gDh.Agree( secretKeyA, privA, pubB ) ) {
        fprintf(stderr, "gDh.Agree Failed!\n");
        return -1;
    }  

    Integer key;
    key.Decode( secretKeyA.BytePtr(), secretKeyA.SizeInBytes() );
    std::cout << "cppclient secretKeyA: " << std::hex << key << std::endl;    

    return 0;
}