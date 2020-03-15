# AreYouIN Server

## Get started
```shell
$ docker build -t areyouin/server .
$ docker run --rm -it \  
    -p 1822:1822 \  
    -p 2022:2022 \  
    -p 40186:40186 \  
    areyouin/server`
```