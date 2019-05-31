# Evepraisal
Evepraisal is a bulk-price estimator for Eve Online.

## Docker Instructions
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
