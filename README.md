# Clipper-CI-Worker

CI Worker microservice of Clipper CI\CD.

## Overview

This microservice is used to consume CI job messages from rabbitMQ queue, execute CI process in Docker (git clone, docker build, docker push), and write build info to PostgreSQL DB. It requires GCR hostname to push images to, and json auth file for [it](https://cloud.google.com/container-registry/docs/advanced-authentication) .

## Installation

1. Before using worker itself, you must make builder image from `ci-builder` directory or use default.
2. Go to GCP console and create service account credentials, save json file to convinient location and remember it.
3. Also, you must find out your gcr url. It consists of gcr hostname, ex. `eu.gcr.io` (European region gcr, check gcloud docs for others) and your project name in form of `gcr_hostname/project-name/` ex. `eu.gcr.io/wild-gophers-1488228/`.
4. Make sure you have API microservice, PostgreSQL and rabbitMQ running and configurated.
5. Build worker binary using `go build` or use docker image.

## Command line arguments
Call worker with `help` argument to see built-in help.
Call worker binary with `start` argument and next parameters.

| Parameter (short)     | Default                           | Usage                                                     |
|-----------------------|-----------------------------------|-----------------------------------------------------------|
| --rabbitmq (-r)       | amqp://guest:guest@localhost:5672 | rabbitmq connection URL                                   |
| --queue (-q)          | ciJobs                            | Name of rabbitmq queue to get ci jobs from                |
| --gcr (-g)            | <not set>                         | Your GCR URL                                              |
| --json (-j)           | secrets                           | Path to your GCP service account json credentials file    |
| --postgresAddr (-a)   | postgres:5432                     | PostgreSQL address                                        |
| --db (-d)             | clipper                           | PostgreSQL database to use                                |
| --user (-u)           | clipper                           | PostgreSQL database user                                  |
| --pass (-c)           | clipper                           | PostgreSQL user's password                                |
| --cdqueue (-k)        | cdJobs                            | Name of rabbitmq queue to put cd jobs to                  |
| --builder (-b)        | ci-builder                        | Name of builder docker image                              |
| --verbose (-v)        | false                             | Show debug level logs                                     |