1. 生成密钥，证书

    openssl genrsa -out server.key 2048

    openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

    openssl req -newkey rsa:2048 -nodes -keyout server.key -x509 -days 365 -out server.crt

2. 资料

    https://gist.github.com/denji/12b3a568f092ab951456

    https://posener.github.io/http2/