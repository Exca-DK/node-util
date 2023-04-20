# Node-Util
This repo includes usefull services and tools for monitoring, gathering and analyzing p2p network.

Main modules are:
### Crawler
Crawls the p2p network and streams data into the provided target. Additionaly it saves entries to DB.

### Indexer
Awaits for data from event queue and transforms it

### Scanner
Scans server ports

## Getting Started

You can decide to run each of the modules as either service or command.

### Prerequisites

It is recommended to have docker installed since most of the examples are based on it.

  ```sh
  [Docker](https://github.com/docker/docker-ce/releases) 
  [docker-compose](https://github.com/docker/compose/releases). 
  ```

## Usage

Either use single component in which instructions are placed inside the module or use service example provided in /infra/docker-compose/

Steps to run service:

1. Set env variables:
    - export NODE_UTIL_SCANNER_REDIS_PASSWORD=""
    - export NODE_UTIL_DB_USERNAME=""
    - export NODE_UTIL_DB_PASSWORD=""
    - export NODE_UTIL_BROKER_CRAWLER_PASSWORD=""
    - export NODE_UTIL_BROKER_SCANNER_PASSWORD=""
1. cd ./infra/docker-compose/
2. docker compose up -d

Default passwords for broker accounts are: admin_secret, crawler_secret, scanner_secret.
If you change them, remember to update the hashes in broker definitions. You can easly regenerate the hashes with gen_pass.sh script

## Metrics

All services output usefull metrics. You can access them at port 6061 of service.

Example at /infra/docker-compose/ includes prometheus scraper. You can access it at port 9090.

## Database

By default the postgres interface is exposed on localhost. You can access it with default credentials.

You can extend the database by simply editing sql tables in /service/db/migration and queries at /service/db/query and recompile everything by using  provided make file at /service

Some tools require access to db such as migration. In order to run them you need to set env variable `NODE_UTIL_DB_ENDPOINT`, `NODE_UTIL_DB_USERNAME` and `NODE_UTIL_DB_PASSWORD`




### Contribution
Contribution is welcome. If you'd like to contribute, please fork, change, commit and send a pull request.