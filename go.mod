module github.com/saveio/scan

go 1.16

replace (
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/max => ../max
	github.com/saveio/themis => ../themis
)

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
	github.com/saveio/carrier v0.0.0-20210802055929-7567cc29dfc9
	github.com/saveio/dsp-go-sdk v0.0.0-20210826071407-3bf6970de27f
	github.com/saveio/max v0.0.0-20210825101853-a279f7982519
	github.com/saveio/pylons v0.0.0-20210802062637-12c41e6d9ba7
	github.com/saveio/themis v1.0.148-0.20210825062649-e5038bf25c91
	github.com/saveio/themis-go-sdk v0.0.0-20210802052239-10a9844e20d5
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
)
