# Network Control (Decentralized ANO Management)

The purpose of this module is to manage the entries to the ano management chain, calculate results, and provide a voting interface for node operators. A proposal can be initiated by any node in the authority set (fed or audit) and ends when a consensus of >50% is reached or 1000 blocks pass, whichever comes first.

Decentralized ANO Management is planned for the Factom MainNet and the community testnet with an optional setting for other networks.


## Proposal Details

Proposals are created by adding a correctly formatted entry to the ANO Management chain. This proposal will be picked up by nodes and a corresponding proposal will be shown in the voting interface. ANOs can vote for a proposal by submitting a correctly formatted vote. Creating a proposal counts as a vote from that ANO. Each server can vote only once. There is no explicit "no" vote.

ANO Management is enabled via an activation height after which the system is responsible for auth set management, disabling the skeleton key and skeleton identity. To make the transition smoother for legacy nodes, Factom Inc. can reveal the skeleton key after the activation height has passed.

## Config File

The following new section is available in the `factomd.conf` config file:
```ini
[networkmanagement]
Port=8030
History=10000
EntryCreditKey=
SkeletonKey=
```

* `Port` specifies the port of the voting interface server. If set to 0, the server is disabled. Default = 8030.
* `History` specifies how long proposals should be displayed in the history section of the voting interface (in factom blocks)
* `EntryCreditKey` is the private key of EC address used to submit proposals and votes. There will be a communal EC address for this purpose.
* `SkeletonKey` is the current network skeleton key, which can be used to sign proper messages for legacy nodes


## Chain

The ANO Management chain is `888888165d185ba3342d8f0dcc331066f454196f1ad7060b00f856b6f483b619`, which exists on both [MainNet](https://explorer.factom.pro/chains/888888165d185ba3342d8f0dcc331066f454196f1ad7060b00f856b6f483b619) and the [community testnet](https://testnet.factoid.org/entry?hash=7033579b015ebeee3d5e146321978cc9640cf43eed32b0eb3b18a3c67f9b8649). It is generated via the ExtIDs "Factom ANO Management" and "1587229" (string).

## Actions

There are three actions: promote, remove, and vote. A proposal (promote, remove) is active starting the block it is included in and lasts for 1,000 blocks. If a proposal is started on height 12345, the last block you can vote in is 13344. The desired result is included in the next block after the passing vote. If a proposal reaches consensus with an entry that is included in block 12345, the appropriate message is included in the admin chain of block 12346.

### Promote

If the identity in the proposal is **not** part of the auth set, the identity must be correctly registered with the identity registration chain and possess all the necessary attributes to run as an auth server. If the identity is not valid, the proposal is considered invalid. New servers can only be promoted to **audit** node, not directly to fed.

If the identity in the proposal is already part of the auth set, this will toggle its status. A successful promotion for a federated node will make it an audit server. A successful promotion for an audit server will make it a federated node. Both of these actions will change the total number of feds the protocol has, decreasing or increasing it by one respectively.

### Remove

If the identity in the proposal is **not** part of the auth set, the proposal is considered invalid. Otherwise, a successful vote will remove that server from the auth set. If the node was a fed, this will decrease the total number of feds by one.

### Vote

A vote is for a specific proposal and counts as a single vote. Each identity in the auth set has one vote.

## Format

The entries follow a specific format that must be signed by the node's block signing key.

### Promote & Remove

