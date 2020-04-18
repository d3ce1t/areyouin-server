# AreYouIN Server

## Quick start
```shell
$ cp areyouin.example.yaml areyouin.yaml
$ docker-compose up -d
$ docker-compose exec cassandra bash
$ sleep 30 && cqlsh < /tmp/schema.cql
$ exit
$ docker-compose restart
$ docker-compose logs -f
```