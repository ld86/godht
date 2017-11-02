FROM ubuntu:latest
MAINTAINER Bogdan Melnik "teh.ld86@gmail.com"

CMD ./godht
WORKDIR /home/dht

RUN add-apt-repository ppa:gophers/archive
RUN apt-get update
RUN apt-get intsall -y golang-1.9-go

COPY godht /home/dht/godht
