module gitlab.com/akita/mem/v2

require (
	github.com/golang/mock v1.4.4
	github.com/google/btree v1.0.0
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/onsi/ginkgo v1.15.1
	github.com/onsi/gomega v1.10.5
	github.com/rs/xid v1.2.1
	gitlab.com/akita/akita v1.3.2
	gitlab.com/akita/akita/v2 v2.0.0-alpha.3
	gitlab.com/akita/util/v2 v2.0.0-alpha.3
	golang.org/x/sys v0.0.0-20210313110737-8e9fff1a3a18 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	rsc.io/quote/v3 v3.1.0 // indirect
)

// replace gitlab.com/akita/akita => ../akita

// replace gitlab.com/akita/util => ../util

go 1.16
