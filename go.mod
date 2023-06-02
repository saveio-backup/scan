module github.com/saveio/scan

go 1.16

require (
	github.com/anacrolix/dht v0.0.0-20181123025733-9b0a8e862ccc
	github.com/anacrolix/dht/v2 v2.8.0 // indirect
	github.com/anacrolix/envpprof v1.1.1
	github.com/anacrolix/log v0.9.0 // indirect
	github.com/anacrolix/missinggo v1.2.1
	github.com/anacrolix/torrent v1.15.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/saveio/carrier v0.0.0-20230322093539-24eaadd546b5
	github.com/saveio/dsp-go-sdk v0.0.0-20221129101740-684879440b10
	github.com/saveio/max v0.0.0-20230324091118-889c76b11561
	github.com/saveio/pylons v0.0.0-20230322094600-b5981ca8ed91
	github.com/saveio/themis v1.0.175
	github.com/saveio/themis-go-sdk v0.0.0-20230314033227-3033a22d3bcd
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
)

replace (
	github.com/saveio/carrier => ../carrier
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/max => ../max
	github.com/saveio/pylons => ../pylons
	github.com/saveio/scan => ../scan
	github.com/saveio/themis => ../themis
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
)
