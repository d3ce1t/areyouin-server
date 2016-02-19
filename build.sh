#/usr/bin/sh
cd common
go build && go install
cd ../dao
go build && go install
cd ../protocol
go build && go install
cd ../facebook
go build && go install
cd ../webhook
go build && go install
cd ../images_server
go build && go install
cd ../server
go build
