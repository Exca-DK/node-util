# Node Crawler

Crawls the crypto p2p network 

Features:
- Allows you to add filters to narrow down the node search across network
- Easly expendable due to modularity
- Can stream data externally
- Minimal network usage

### Starting the crawler
In order to start the crawler `atleast one` bootnode as target is required as `--nodes` flag

Additionaly by specyfying `--identities` you can narrow down the crawl range.
Some networks can filter your node if the caps don't match. You can specify them with flag `--fakecaps`. If none flag is provided the crawler will use default eth caps.

In some usecases bad node name can also result in getting restricted across the network. You can specify the name with flag `--fakename` eg. geth/v1.2.3/linux-amd64/go1.18.1
```
$ crawler.exe --nodes=<your bootnodes> 
```
### Extra commands
It is possible to run just specific protocol messages. 

Currently it is possible to run:
- Ping node
- Request node to send it's random peers
- Request Enr record
- Peeking at db of the peer by constantly querying random nodes

Help command provides description and usage for all flags and subcommands. 
```
$ go run ./cmd/disc/main.go --help
```



### Production
Build crawler and copy the binary to desired location. 
```
$ make crawler && mv ./build/bin/crawler
```
Create a systemd service that executes binary with crawl command.

### Docker setup
```
Make sure you have [Docker](https://github.com/docker/docker-ce/releases) and [docker-compose](https://github.com/docker/compose/releases) tools installed. 
```

execute docker: `docker compose up`