# Batch example

Web application example using [github.com/elgopher/batch](https://github.com/elgopher/batch) package.

## How to run?

* Install docker, docker-compose, curl

* Start web application with Postgres database

`$ docker-compose up`

* Buy a train ticket

`$ curl -v "http://localhost:8080/book?train=batchy&person=elgopher&seat=3"`