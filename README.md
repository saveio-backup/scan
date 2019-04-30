# SCAN

Save Content Addressing Node

## Overview
> #### Module Design
* Network：Use oniP2P for foundenmental network comunication and try tobe a complete connected graph.
* Tracker：Take the idea in BT, use tracker services to help find the oniFS service nodes.
* Dsp.DNS：Manage dns nodes, work with oniChain native smartcontract.
* Dsp.Channel：Settle done layer2 channels for routing and management.
* API：Support rpc & restful interface.
* Actor：Use actor concurrent model for decoupling different module jobs.
* CMD：Desc: Works as a full node, inherit oniChain commands to have all management features.
    * Dev Schedule
        * Channel cmd features
            * Openchannel
            * Closechannel
            * Payment
            * Listchannels
            * CloseAllChannels
        * DNS cmd features



