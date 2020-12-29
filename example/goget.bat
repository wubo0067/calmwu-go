@echo off

set http_proxy=socks5://127.0.0.1:1091
set https_proxy=socks5://127.0.0.1:1091

go get -u -v %*

echo ...

pause