module github.com/saveio/scan

go 1.14

replace (
	github.com/saveio/carrier => ../carrier
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/edge => ../edge
	github.com/saveio/max => ../max
	github.com/saveio/pylons => ../pylons
	github.com/saveio/themis => ../themis
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
)

require (
	github.com/anacrolix/dht v0.0.0-20181123025733-9b0a8e862ccc
	github.com/anacrolix/envpprof v1.1.1
	github.com/anacrolix/missinggo v1.2.1
	github.com/anacrolix/torrent v1.26.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/saveio/carrier v0.0.0-00010101000000-000000000000
	github.com/saveio/dsp-go-sdk v0.0.0-00010101000000-000000000000
	github.com/saveio/edge v0.0.0-00010101000000-000000000000
	github.com/saveio/max v0.0.0-00010101000000-000000000000
	github.com/saveio/pylons v0.0.0-00010101000000-000000000000
	github.com/saveio/themis v0.0.0-00010101000000-000000000000
	github.com/saveio/themis-go-sdk v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	github.com/urfave/cli v1.22.5
)
