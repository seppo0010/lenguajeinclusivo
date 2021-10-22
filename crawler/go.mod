module github.com/seppo0010/juscaba/crawler

replace github.com/seppo0010/juscaba/shared => ../shared

require (
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/seppo0010/juscaba/shared v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/streadway/amqp v1.0.0
)

require (
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/compress v1.13.5 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/minio-go/v7 v7.0.15 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20201216223049-8b5274cf687f // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
)

go 1.17
