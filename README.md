# Operator service to provide public download URLS for Jenkins artifacts

[![Go Report Card](https://goreportcard.com/badge/github.com/philbrookes/artifact-proxy-operator)](https://goreportcard.com/report/github.com/philbrookes/artifact-proxy-operator)
[![CircleCI](https://circleci.com/gh/philbrookes/artifact-proxy-operator.svg?style=svg)](https://circleci.com/gh/philbrookes/artifact-proxy-operator)

*Note* Just a POC at the moment

## Permissions

Currently this needs to use a service account with admin permissions. Use:

```sh
$ kubectl create clusterrolebinding <your namespace>-admin-binding --clusterrole=admin --serviceaccount=<your namespace>:default
```

## Usage

Make sure that you are logged in with `oc` and use the right namespace.

```
$ make build_linux
$ docker build -t docker.io/<dockerorg>/artifact-proxy-operator:latest -f Dockerfile .
$ oc create -f operator.json
```
