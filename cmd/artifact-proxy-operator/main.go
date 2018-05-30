package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/aerogear/artifact-proxy-operator/pkg/jenkins"
	"github.com/aerogear/artifact-proxy-operator/pkg/openshift"
	"github.com/aerogear/artifact-proxy-operator/pkg/plist"
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
		log.Fatalf("error starting http server on %s, (%s)", listen, err.Error())
	}
	fmt.Printf("listening on %s", listen)
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

	buildType, err := osClient.GetBuildType(build)
	if err != nil {
		http.Error(rw, fmt.Sprintf("no build type found for build %s", build), http.StatusBadRequest)
		return
	}
	switch buildType {
	case "android":
		handleBinaryResponse(rw, artifactUrl, "apk")
	case "ios":
		if isArtifactRequest(r.URL) {
			handleBinaryResponse(rw, artifactUrl, "ipa")
		}
		handleXmlResponse(rw)
	default:
		http.Error(rw, fmt.Sprintf("invalid build type found for build %s", build), http.StatusBadRequest)
		return
	}

}

func handleBinaryResponse(rw http.ResponseWriter, artifactUrl string, extension string) {
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
	rw.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"app.%s\"", extension))
	if _, err := io.Copy(rw, artifactStreamer); err != nil {
		fmt.Println("error writing download of application binary")
		return
	}
}

func handleXmlResponse(rw http.ResponseWriter) {
	// TODO Get plist artifact from jenkins and read its contents
	// Get the proxy url for this build which will just be existing plist url via proxy
	// with ?artifact=true appended as query param
	plist.ModifyXML(bytes.NewBuffer([]byte{}), "url", "https://proxy-url-placeholder")
	rw.Header().Set("content-type", "text/xml")
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

func isArtifactRequest(url *url.URL) bool {
	return url.Query().Get("artifact") == "true"
}
