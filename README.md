# DASLink
DASLink is a simple tool to link ipfs content from [.bit](https://www.did.id/).

## How does it work?
Dependent on [DNSLink](https://docs.ipfs.io/concepts/dnslink/), [Cloudflare ipfs gateway](https://developers.cloudflare.com/distributed-web/ipfs-gateway), [Cloudflare DNS](https://api.cloudflare.com/#dns-records-for-a-zone-properties) and [das-database](https://github.com/dotbitHQ/das-database).

```
┌───────────┐               ┌───────────┐               ┌────────────┐
│           │               │           │               │            │
│   Alice   │               │    DNS    │               │ipfs gateway│
│           │               │           │               │            │
└─────┬─────┘               └─────┬─────┘               └──────┬─────┘
      │                           │                            │
      │    visit alice.bit.cc     │                            │
      ├──────────────────────────►│                            │
      │                           │       CNAME point to       │
      │                           ├───────────────────────────►│
      │                           │                            │
      │                           │◄───────────────────────────┤
      │                           │ looking up the TXT record  │
      │                           ├───────────────────────────►│
      │                           │                            ├───────────┐
      │                           │                            │           │
      │                           │                            │    get the│ipfs content
      │                           │                            │           │
      │    return ipfs content    │                            │◄──────────┘
      │◄──────────────────────────┼────────────────────────────┤
      │                           │                            │
      │                           │                            │
┌─────┴─────┐               ┌─────┴─────┐               ┌──────┴─────┐
│           │               │           │               │            │
│   Alice   │               │    DNS    │               │ipfs gateway│
│           │               │           │               │            │
└───────────┘               └───────────┘               └────────────┘
```

## Install
```
# run das-database and keep it synchronized with the latest data
https://github.com/dotbitHQ/das-database

# get the code
git clone https://github.com/paicha/daslink.git

# get your Cloudflare api tokens
https://dash.cloudflare.com/profile/api-tokens

# edit config.yaml
cd config
cp config.yaml.sample config.yaml
vi config.yaml

# compile and run
go build
./daslink
```
