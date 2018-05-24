# Operator service to provide public download URLS for Jenkins artifacts

[![Go Report Card](https://goreportcard.com/badge/github.com/philbrookes/artifact-proxy-operator)](https://goreportcard.com/report/github.com/philbrookes/artifact-proxy-operator)
[![CircleCI](https://circleci.com/gh/philbrookes/artifact-proxy-operator.svg?style=svg)](https://circleci.com/gh/philbrookes/artifact-proxy-operator)

*Note* Just a POC at the moment
## Usage

Deploy the template to a namespace, with a parameter for the URL the artifact proxy should use to serve artifacts, like so:

```
oc new-app -f operator.json -p OPERATOR_HOSTNAME=artifact-proxy-operator-route
```

Add an annotation to a build, to trigger the creation of a proxy URL and token, the annotation should like so:
```
aerogear.org/download-mobile-artifact: "true"
```
Once the build object is saved with this annotation, reload the build object to see the new annotations created by this operator.