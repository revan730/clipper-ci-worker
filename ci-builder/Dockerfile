FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y git curl

ENV DOCKERVERSION=18.03.1-ce
RUN curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz \
  && tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 \
                 -C /usr/local/bin docker/docker \
  && rm docker-${DOCKERVERSION}.tgz

RUN mkdir /builder
WORKDIR /builder
ADD . .
RUN chmod +x /builder/builder.sh

ENTRYPOINT [ "/builder/builder.sh" ]
