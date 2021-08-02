module github.com/saveio/scan

go 1.16

replace (
	github.com/saveio/dsp-go-sdk => ../dsp-go-sdk
	github.com/saveio/themis => ../themis
	github.com/saveio/themis-go-sdk => ../themis-go-sdk
)

require (
	github.com/anacrolix/dht v0.0.0-20181123025733-9b0a8e862ccc
	github.com/anacrolix/dht/v2 v2.8.0 // indirect
	github.com/anacrolix/envpprof v1.1.1
	github.com/anacrolix/log v0.9.0 // indirect
	github.com/anacrolix/missinggo v1.2.1
	github.com/anacrolix/torrent v1.15.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.3.2
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/marten-seemann/qtls v0.10.0 // indirect
	github.com/ontio/ontology-eventbus v0.9.1
	github.com/saveio/carrier v0.0.0-20210519082359-9fc4d908c385
	github.com/saveio/dsp-go-sdk v0.0.0-20210527074313-4b3d6afed755
	github.com/saveio/edge v1.0.250
	github.com/saveio/max v0.0.0-20210728025229-47aa8d4cc8f9
	github.com/saveio/pylons v0.0.0-20210519083005-78a1ef20d8a0
	github.com/saveio/themis v1.0.121
	github.com/saveio/themis-go-sdk v0.0.0-20210702081903-52a40e927ed8
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210427231257-85d9c07bbe3a // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
)
