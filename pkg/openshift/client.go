package openshift

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/aerogear/artifact-proxy-operator/pkg/jenkins"
	apibuildv1 "github.com/openshift/api/build/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	JenkinsBuildUri         = "openshift.io/jenkins-build-uri"
	WatchResourceAnnotation = "aerogear.org/download-mobile-artifact"
	JenkinsArtifactUri      = "aerogear.org/jenkins-mobile-artifact-url"
	DownloadProxyUri        = "aerogear.org/download-mobile-artifact-url"
	ArtifactDownloadToken   = "aerogear.org/mobile-artifact-token"
)

type OpenShiftClient struct {
	AuthToken     string
	BuildClient   *buildv1.BuildV1Client
	JenkinsClient *jenkins.JenkinsClient
}

func (c *OpenShiftClient) GetBuild(build string) (*apibuildv1.Build, error) {
	log.Printf("getting build info for build - " + build)
	b, err := c.BuildClient.Builds(os.Getenv("NAMESPACE")).Get(build, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return b, err
}

func (c *OpenShiftClient) GetDownloadConst() string {
	return JenkinsArtifactUri
}

func (c *OpenShiftClient) GetTokenConst() string {
	return ArtifactDownloadToken
}

func (c *OpenShiftClient) WatchBuilds() {
	events, err := c.BuildClient.Builds(os.Getenv("NAMESPACE")).Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for update := range events.ResultChan() {
		raw, _ := json.Marshal(update.Object)
		var build = apibuildv1.Build{}
		json.Unmarshal(raw, &build)
		//artifact download url requested
		if val, ok := build.Annotations[WatchResourceAnnotation]; ok && val == "true" {
			//and not provided yet
			if _, ok := build.Annotations[JenkinsArtifactUri]; !ok {
				c.addAnnotations(&build)
				log.Printf("Download requested for %v\n", build.ObjectMeta.Name)
			} else {
				log.Printf("Download already provided for %v\n", build.ObjectMeta.Name)
			}
		} else {
			log.Printf("Download not requested for %v\n", build.ObjectMeta.Name)
		}
	}
}

func (c *OpenShiftClient) addAnnotations(build *apibuildv1.Build) {
	buildDetails, err := c.JenkinsClient.GetBuildInfo(build.Annotations[JenkinsBuildUri], c.AuthToken)
	if err != nil {
		log.Println("error " + err.Error() + " fetching build details for build " + build.Name)
		return
	}
	if len(buildDetails.Artifacts) < 1 {
		log.Println("no artifact information available for build " + build.Name)
		return
	}
	build.Annotations[JenkinsArtifactUri] = build.Annotations[JenkinsBuildUri] + "artifact/" + buildDetails.Artifacts[0].RelativePath
	token := build.Name + "-" + strconv.FormatInt(buildDetails.Timestamp, 10)
	build.Annotations[ArtifactDownloadToken] = token
	operatorHost := os.Getenv("OPERATOR_HOSTNAME")
	if operatorHost == "" {
		log.Println("no hostname available to set annotation")
	}
	build.Annotations[DownloadProxyUri] = "http://" + operatorHost + "/" + build.Name + "/download?token=" + token
	_, err = c.BuildClient.Builds(os.Getenv("NAMESPACE")).Update(build)
	if err != nil {
		log.Println("error " + err.Error() + " while updating build annotations for build " + build.Name)
	}

}

func GetBuildClient() (*buildv1.BuildV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	return buildv1.NewForConfig(config)
}

func NewOpenShiftClient(jc *jenkins.JenkinsClient) (*OpenShiftClient, error) {
	token, err := getAuthToken()
	if err != nil {
		return nil, err
	}

	buildClient, err := GetBuildClient()
	if err != nil {
		return nil, err
	}
	return &OpenShiftClient{token, buildClient, jc}, nil
}

func getAuthToken() (string, error) {
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token") // just pass the file name
	if err != nil {
		return "", errors.New("error reading service account token " + err.Error())
	}
	return string(b), nil // convert content to a 'string'
}
