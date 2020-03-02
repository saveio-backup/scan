package network

import (
	"github.com/saveio/carrier/network"
	"github.com/saveio/themis/common/log"
)

type NetComponent struct {
	*network.Component
	Net *Network
}

func (this *NetComponent) Startup(net *network.Network) {
}

func (this *NetComponent) Receive(ctx *network.ComponentContext) error {
	msg := ctx.Message()
	client := ctx.Client()
	log.Debugf("client ++++ %v, addr %s, peer %s", client, client.Address, client.ClientID())
	addr := ctx.Client().Address
	peerId := ctx.Client().ClientID()

	return this.Net.Receive(msg, addr, peerId)
}
