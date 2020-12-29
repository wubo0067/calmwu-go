编译中遇到的问题
它需要依赖这个module：k8s.io/api/admissionregistration/v1alpha1，报错，github上同样有人问https://github.com/kubernetes/client-go/issues/551。
修改go.mod文件，将k8s.io/api kubernetes-1.13.2改为如此。再执行go mod vendor。这样编译通过。

编译命令：go build -x -v -mod=vendor k8sclient.go

./k8sclent 

mklink /J D:\develope\gopath\src\k8s.io\klog D:\develope\gopath\src\k8sclient\vendor\k8s.io\klog
mklink /J D:\develope\gopath\src\k8s.io\client-go D:\develope\gopath\src\k8sclient\vendor\k8s.io\client-go
mklink /J D:\develope\gopath\src\k8s.io\api D:\develope\gopath\src\k8sclient\vendor\k8s.io\api
mklink /J D:\develope\gopath\src\k8s.io\apimachinery D:\develope\gopath\src\k8sclient\vendor\k8s.io\apimachinery