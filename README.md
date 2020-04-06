# relaybroker [![Build Status](https://travis-ci.org/gempir/relaybroker.svg?branch=master)](https://travis-ci.org/gempir/relaybroker)

#### What is this?
relaybroker is a piece of software that is supposed to act as a proxy between your bot and twitch.tv irc servers. 
It will handle ratelimiting so you don't have to worry about getting global banned or having connection issues.

#### How to use 
Loglevel can be changed via env var "LOGLEVEL". The options are debug, info and error. Info is default.

The ip/port relaybroker listens to can be changed via the env var "BROKERHOST". The default is "127.0.0.1:3333" which means it listens to port 3333 on local

### Docker

#### Warning this is old and not updated currently

Run relaybroker as a docker container like this:

    docker run -p 8080:3333 -e BROKERPASS="mypassword" gempir/relaybroker 