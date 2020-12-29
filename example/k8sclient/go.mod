module k8sclient

go 1.12

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/huandu/xstrings v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.1 // indirect
	github.com/monnand/dhkx v0.0.0-20180522003156-9e5b033f1ac4 // indirect
	github.com/pkg/errors v0.8.1
	github.com/snwfdhmp/errlog v0.0.0-20191219134421-4c9e67f11ebc // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/urfave/cli v1.20.0
	github.com/wubo0067/calmwu-go v0.0.0-20200410083741-d348bac27c84
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553 // indirect
	google.golang.org/genproto v0.0.0-20191216205247-b31c10ee225f // indirect
	helm.sh/helm/v3 v3.0.2
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/helm v2.16.1+incompatible
	k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// Kubernetes imports github.com/miekg/dns at a newer version but it is used
	// by a package Helm does not need. Go modules resolves all packages rather
	// than just those in use (like Glide and dep do). This sets the version
	// to the one oras needs. If oras is updated the version should be updated
	// as well.
	github.com/miekg/dns => github.com/miekg/dns v0.0.0-20181005163659-0d29b283ac0f
	gopkg.in/inf.v0 v0.9.1 => github.com/go-inf/inf v0.9.1
	gopkg.in/square/go-jose.v2 v2.3.0 => github.com/square/go-jose v2.3.0+incompatible

	rsc.io/letsencrypt => github.com/dmcgowan/letsencrypt v0.0.0-20160928181947-1847a81d2087
)
