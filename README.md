# AreYouIN Server

## Quick get started (Local)
```
$ go build -o ./server.bin ./server
$ ./server.bin
```


## Quick get started (Docker)
```shell
$ docker build -t d3ce1t/areyouin-server .
$ docker run --rm -it \  
    -p 1822:1822 \  
    -p 2022:2022 \  
    -p 40186:40186 \  
    d3ce1t/areyouin-server
```