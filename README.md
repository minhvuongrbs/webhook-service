# Webhook Service
Handling webhook and notify to partner the corresponding events

## Documentation
Finding documentation in folder: ./docs

## Technical Stack
Programming Language: Golang\
Database: MySQL\
Message Broker: Kafka\
[Temporal](https://temporal.io) for handling retryable activities

## Project structure
```
┣ cmd: init system with cli
┣ docs: repository's documentation
┣ pkg: repository's package
┣ internal
┃ ┣ adapters: layer to define traffic going out of system
┃ ┣ app: appliation logic
┃ ┣ common: common definition share between multiple flows
┃ ┣ entities: domain entities
┃ ┣ ports: layer to define traffic comming to system
┃ ┃ ┣ kafka_consumer
┃ ┃ ┗ temporal_worker
┃ ┣ service: init dependencies of application
┃ ┃ ┗ application.go
```

## Installation
```bash
make docker-compose-up
```
 
