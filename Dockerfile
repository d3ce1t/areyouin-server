FROM golang:1.10 as build

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/peeple/areyouin
COPY . .

RUN go get -u github.com/gocql/gocql \
    && go get -u github.com/golang/protobuf/proto \
    && go get -u github.com/golang/protobuf/protoc-gen-go \
    && go get -u github.com/imkira/go-observer \
    && go get -u github.com/twinj/uuid \
    && go get -u github.com/huandu/facebook \
    && go get -u github.com/google/go-gcm \
    && go get -u github.com/disintegration/imaging \
    && go get -u golang.org/x/crypto/ssh \
    && go get -u gopkg.in/yaml.v2
    
RUN ./build.sh \
    && mkdir -p dist/cert \
    && cp server/server dist/server \
    && cp server/extra/areyouin.example.yaml dist/areyouin.yaml

FROM golang:1.10
# RUN apk --no-cache add ca-certificates bash
WORKDIR /root/
COPY --from=build /go/src/peeple/areyouin/dist .
CMD ["./server"]

EXPOSE 1822 2022 40186