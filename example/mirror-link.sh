#!/bin/bash
git clone https://github.com/golang/image.git
git clone https://github.com/golang/text.git
git clone https://github.com/golang/net.git
git clone https://github.com/golang/tools.git
git clone https://github.com/golang/crypto.git
git clone https://github.com/golang/oauth2.git
git clone https://github.com/golang/exp.git
git clone https://github.com/golang/sys.git

ln -s /Users/calmwu/Program/Cestbon/src/mirror/tools /Users/calmwu/Program/Cestbon/src/golang.org/x/tools
ln -s /Users/calmwu/Program/Cestbon/src/mirror/net /Users/calmwu/Program/Cestbon/src/golang.org/x/net
ln -s /Users/calmwu/Program/Cestbon/src/mirror/crypto /Users/calmwu/Program/Cestbon/src/golang.org/x/crypto
ln -s /Users/calmwu/Program/Cestbon/src/mirror/oauth2 /Users/calmwu/Program/Cestbon/src/golang.org/x/oauth2
ln -s /Users/calmwu/Program/Cestbon/src/mirror/sys /Users/calmwu/Program/Cestbon/src/golang.org/x/sys
ln -s /Users/calmwu/Program/Cestbon/src/mirror/exp /Users/calmwu/Program/Cestbon/src/golang.org/x/exp
ln -s /Users/calmwu/Program/Cestbon/src/mirror/image /Users/calmwu/Program/Cestbon/src/golang.org/x/image
ln -s /Users/calmwu/Program/Cestbon/src/mirror/text /Users/calmwu/Program/Cestbon/src/golang.org/x/text


go get -u -v github.com/mdempsky/gocode
go get -u -v github.com/rogpeppe/godef
go get -u -v github.com/golang/lint/golint
go get -u -v github.com/lukehoban/go-outline
go get -u -v sourcegraph.com/sqs/goreturns
go get -u -v golang.org/x/tools/cmd/gorename
go get -u -v github.com/tpng/gopkgs
go get -u -v github.com/newhook/go-symbols
go get -u -v golang.org/x/tools/cmd/guru

cd D:\develope\gopath\src
go get -u -v golang.org/x/tools/gopls

go build github.com/mdempsky/gocode
go build github.com/rogpeppe/godef
go build github.com/golang/lint/golint
go build github.com/lukehoban/go-outline
go build sourcegraph.com/sqs/goreturns
go build golang.org/x/tools/cmd/gorename
go build github.com/tpng/gopkgs
go build github.com/newhook/go-symbols
go build golang.org/x/tools/cmd/guru

go build -o D:\develope\gopath\bin\gocode-gomod.exe github.com/stamblerre/gocode
go build -o D:\develope\gopath\bin\godef-gomod.exe github.com/ianthehat/godef

go get -u -v github.com/nsf/gocode go get -u -v github.com/uudashr/gopkgs/cmd/gopkgs go get -u -v github.com/ramya-rao-a/go-outline go get -u -v github.com/acroca/go-symbols go get -u -v golang.org/x/tools/cmd/guru go get -u -v golang.org/x/tools/cmd/gorename go get -u -v github.com/rogpeppe/godef go get -u -v golang.org/x/tools/cmd/godoc go get -u -v github.com/zmb3/gogetdoc go get -u -v github.com/sqs/goreturns go get -u -v golang.org/x/tools/cmd/goimports go get -u -v github.com/golang/lint/golint go get -u -v github.com/alecthomas/gometalinter go get -u -v honnef.co/go/tools/... go get -u -v github.com/derekparker/delve/cmd/dlv
go get -u -v github.com/haya14busa/goplay/cmd/goplay go get -u -v github.com/josharian/impl go get -u -v github.com/tylerb/gotype-live go get -u -v github.com/cweill/gotests/... go get -u -v github.com/sourcegraph/go-langserver go get -u -v github.com/davidrjenni/reftools/cmd/fillstruct


gocode-gomod exit
gocode-gomod -s -debug

go list -json -compiled

Download the vsix file from https://github.com/Microsoft/vscode-go/releases/tag/latest
Run code --install-extension Go-latest.vsix
Reload VS Code
https://github.com/Microsoft/vscode-go/wiki/Use-the-beta-version-of-the-latest-Go-extension

cd /tmp
curl -O https://github.com/Microsoft/vscode-go/releases/download/latest/Go-0.6.92-beta.1.vsix
code --install-extension Go-latest.vsix
go get -u -d github.com/stamblerre/gocode
go build -o $GOPATH/bin/gocode-gomod github.com/stamblerre/gocode
gocode close
gocode-gomod close

//"http.proxy": "http://127.0.0.1:1080",
//"http.proxyStrictSSL": false,  

安装
https://github.com/Microsoft/vscode-go/wiki/Go-modules-support-in-Visual-Studio-Code
https://github.com/Microsoft/vscode-go/releases/tag/latest
https://github.com/golang/go/wiki/gopls

有时候用这个https://github.com/sourcegraph/go-langserver更兼容cgo