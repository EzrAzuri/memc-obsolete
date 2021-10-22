## Memc Library

Accept a large file (somewhere between 5MiB to 50MiB) as an input and store it in Memcache server.Once stored, the library can be used to retrieve the file from Memcached server.

Accept and store file posted from an HTTP Client. Once stored, the HTTP API can be used to retrieve the file from Memcached server.

This library use gomemcache and heavily based on memza library.

## Requirements

1. Go
2. Docker
3. Memcached server

## Setup

1. Clone required repo

```
// change directory (eg: $GOPATH = /home/user/go/src) and create folder if needed

cd $GOPATH/src/github.com/EzrAzuri

git clone https://github.com/EzrAzuri/memc.git

cd $GOPATH/src/github.com/bradfitz

git clone https://github.com/bradfitz/gomemcache.git

cd $GOPATH/src/github.com/EzrAzuri/memc
```

2. Memcached server

```
docker container run -d -p 11211:11211 memcached -m 1000
```

3. Memcached server URL (). If not specified, the local host will be used.

```
export MEMCACHED_SERVER_URL=192.168.1.128:11211
```

4. OPTIONAL: use this command to generate a 50MiB file of random data if needed

```
dd if=/dev/urandom of=50MiB.dat bs=1048576 count=50
```

## Memc Local CLI

CLI help

```
go run memc_local/main.go -h
```

Store a file on local Memcached (without build binary)

```
go run memc_local/main.go -m 192.168.1.128:11211 -s testfile.dat
```

Retrieve a file

```
go run memc_local/main.go -r testfile.dat
```

Build a binary:

```
cd memc_local

go build -o memc_local

./memc_local -h
```

## Memc API

Run API service without building binary

```
go run memc_api/main.go
```

To upload, go to

```
http://localhost:8080/upload
```

To download, go to

```
http://localhost:8080/download
```

Curl command to post file

```
curl -F 'file=@/path/to/testfile.dat' http://localhost:8080/receive
```

Curl command to get file

```
curl "http://localhost:8080/retrieve?key=/path/to/testfile.dat" --output testfile.dat
```

Build API binary with and run with server specifications
```
cd memc_api

go build -o memc_api

./memza -d -s 192.168.1.128:11211
```
