# Evepraisal
Evepraisal is a bulk-price estimator for Eve Online.

## Docker Instructions (production)
The following was tested on Ubuntu Server 18.10
- Install docker.io
```
  $ sudo apt install docker.io
```
- Install docker-compose
```
  $ sudo curl -L "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  $ sudo chmod +x /usr/local/bin/docker-compose
```
- Download Dockerfile, docker-compose.yml, and evepraisal.toml to a directory
```
  $ wget https://github.com/evepraisal/go-evepraisal/blob/master/Dockerfile
  $ wget https://github.com/evepraisal/go-evepraisal/blob/master/docker-compose.yml
  $ wget https://github.com/evepraisal/go-evepraisal/blob/master/evepraisal.toml
```
- build, and bring the container up
```
  $ docker-compose up
```

## Instructions (development)
The following was tested on Ubuntu Server 18.10
- Install golang 1.11
```
  ~$ curl https://dl.google.com/go/go1.11.10.linux-amd64.tar.gz | tar xz
  ~$ sudo mv go /usr/local
  ~$ echo 'export GOROOT=/usr/local/go' >>~/.profile
  ~$ echo 'export GOPATH=$HOME/go' >>~/.profile
  ~$ echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >>~/.profile
  ~$ echo 'export GO111MODULE=on' >>~/.profile
  ~$ source ~/.profile
```
- Install build requirements
```
  ~$ sudo apt install git gcc musl-dev make
```
- Download and build evepraisal
```
  ~$ mkdir -p $GOPATH/src/github.com/evepraisal/go-evepraisal
  ~$ cd $GOPATH/src/github.com/evepraisal/go-evepraisal
  ~/go/src/github.com/evepraisal/go-evepraisal$ git clone https://github.com/evepraisal/go-evepraisal.git .
  ~/go/src/github.com/evepraisal/go-evepraisal$ make setup
  ~/go/src/github.com/evepraisal/go-evepraisal$ make build
```
- Run evepraisal
```
  ~/go/src/github.com/evepraisal/go-evepraisal$ ./target/evepraisal-linux-amd64
```
