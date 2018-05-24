package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/philbrookes/artifact-proxy-operator/pkg/jenkins"
	"github.com/philbrookes/artifact-proxy-operator/pkg/openshift"
)

var osClient *openshift.OpenShiftClient
var jenkinsClient *jenkins.JenkinsClient

func main() {
	var err error
	jenkinsClient = jenkins.NewJenkinsClient()
	osClient, err = openshift.NewOpenShiftClient(jenkinsClient)
	if err != nil {
		log.Fatal("error instantiating OpenShiftClient - error " + err.Error())
	}
	go osClient.WatchBuilds()
	serveHttp()
}

func serveHttp() {
	http.HandleFunc("/", handler)
	listen := os.Getenv("ARTIFACT_PROXY_OPERATOR_SERVICE_PORT")
	if len(listen) == 0 {
		listen = ":8080"
	} else {
		listen = ":" + listen
	}
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal("error starting http server")
	}
	log.Printf("listening on %s", listen)
}

func handler(rw http.ResponseWriter, r *http.Request) {
	isValid, err := validateURLPath(r.URL)
	if err != nil {
		http.Error(rw, "error parsing request", http.StatusInternalServerError)
		return
	}
	if !isValid {
		http.Error(rw, "bad request. route should be called with /<build-id>/download?token=eg-token", http.StatusBadRequest)
		return
	}

	token, err := parseToken(r.URL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) < 2 {
		http.Error(rw, "unable to parse build name from path", http.StatusInternalServerError)
		return
	}
	build, err := osClient.GetBuild(splitPath[1])
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(rw, fmt.Sprintf("no resources found for build %s", splitPath[1]), http.StatusNotFound)
			return
		}
		http.Error(rw, fmt.Sprintf("error fetching build %s", build.Name), http.StatusInternalServerError)
		return
	}
	tokenAnnotationVal, ok := build.Annotations[osClient.GetTokenConst()]
	if tokenAnnotationVal != token || !ok {
		http.Error(rw, fmt.Sprintf("invalid token provided for build %s", build.Name), http.StatusForbidden)
		return
	}
	artifactUrl, ok := build.Annotations[osClient.GetDownloadConst()]
	if !ok || artifactUrl == "" {
		http.Error(rw, "missing annotation on build object", http.StatusInternalServerError)
		return
	}

	artifactStreamer, err := jenkinsClient.StreamArtifact(artifactUrl, osClient.AuthToken)
	if err != nil {
		http.Error(rw, "error when streaming atifact", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := artifactStreamer.Close(); err != nil {
			fmt.Printf("error. failed to close file handle. could be leaking resources %s", err)
		}
	}()
	rw.Header().Set("content-type", "octet/stream")
	rw.Header().Set("content-disposition", "attachment; filename=\"app.apk\"")
	if _, err := io.Copy(rw, artifactStreamer); err != nil {
		fmt.Println("error writing download of application binary")
		return
	}
}

func parseToken(url *url.URL) (string, error) {
	token, ok := url.Query()["token"]

	if !ok || len(token) != 1 {
		return "", errors.New("invalid request, missing token")
	}
	return token[0], nil
}

func validateURLPath(url *url.URL) (bool, error) {
	return regexp.MatchString("/.*/download", url.Path)
}
