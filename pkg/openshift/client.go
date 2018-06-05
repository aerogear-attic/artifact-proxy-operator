package openshift

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aerogear/artifact-proxy-operator/pkg/jenkins"
	apibuildv1 "github.com/openshift/api/build/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	BuildConfig             = "openshift.io/build-config.name"
	JenkinsBuildUri         = "openshift.io/jenkins-build-uri"
	WatchResourceAnnotation = "aerogear.org/download-mobile-artifact"
	JenkinsArtifactUri      = "aerogear.org/jenkins-mobile-artifact-url"
	DownloadProxyUri        = "aerogear.org/download-mobile-artifact-url"
	ArtifactDownloadToken   = "aerogear.org/mobile-artifact-token"
	BuildType               = "mobile-client-type"
	AndroidExtension        = ".apk"
	IosExtenstion           = ".ipa"
)

type OpenShiftClient struct {
	AuthToken     string
	BuildClient   *buildv1.BuildV1Client
	JenkinsClient *jenkins.JenkinsClient
	namespace     string
	operatorHost  string
}

func (c *OpenShiftClient) GenerateArtifactUrl(buildName string, token string, artifact bool) string {
	url := "https://" + c.operatorHost + "/" + buildName + "/download?token=" + token
	if artifact {
		url += "&amp;artifact=true"
	}
	return url
}

func (c *OpenShiftClient) GetBuild(build string) (*apibuildv1.Build, error) {
	log.Printf("getting build info for build - " + build)
	b, err := c.BuildClient.Builds(c.namespace).Get(build, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return b, err
}

func (c *OpenShiftClient) GetBuildType(build *apibuildv1.Build) (string, error) {
	bc, ok := build.Annotations[BuildConfig]
	if !ok {
		return "", errors.New("unable to get build config info for " + build.Name)
	}
	b, err := c.BuildClient.BuildConfigs(c.namespace).Get(bc, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	buildType, ok := b.Labels[BuildType]
	if !ok {
		return "", errors.New("unable to get type for build config " + build.Name)
	}
	return buildType, nil
}

func (c *OpenShiftClient) GetDownloadConst() string {
	return JenkinsArtifactUri
}

func (c *OpenShiftClient) GetTokenConst() string {
	return ArtifactDownloadToken
}

func (c *OpenShiftClient) WatchBuilds() {
	events, err := c.BuildClient.Builds(c.namespace).Watch(metav1.ListOptions{})
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
	log.Printf("watch exited")
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

	var buildType string
	var binArtifact jenkins.Artifact
	for _, artifact := range buildDetails.Artifacts {
		if strings.Contains(artifact.RelativePath, AndroidExtension) {
			buildType = "android"
			binArtifact = artifact
			break
		}
		if strings.Contains(artifact.RelativePath, IosExtenstion) {
			buildType = "ios"
			binArtifact = artifact
			break
		}
	}
	if buildType == "" {
		if len(buildDetails.Artifacts) != 1 {
			log.Printf("can not accurately determine artifact for build %s\n", build.Name)
			return
		}
		binArtifact = buildDetails.Artifacts[0]
		log.Printf("unable to determine build type for build %s from artifact\n", build.Name)
		buildType, err = c.GetBuildType(build)
		if err != nil {
			log.Printf("no build type found for %s. required annotations can't be added\n", build.Name)
			return
		}
	}

	build.Annotations[JenkinsArtifactUri] = build.Annotations[JenkinsBuildUri] + "artifact/" + binArtifact.RelativePath
	token := build.Name + "-" + strconv.FormatInt(buildDetails.Timestamp, 10)
	build.Annotations[ArtifactDownloadToken] = token
	build.Annotations[DownloadProxyUri] = c.GenerateArtifactUrl(build.Name, token, buildType == "android")

	_, err = c.BuildClient.Builds(c.namespace).Update(build)
	if err != nil {
		log.Println("error " + err.Error() + " while updating build annotations for build " + build.Name)
	}

}

func NewOpenShiftClient(jc *jenkins.JenkinsClient) (*OpenShiftClient, error) {
	token, err := getAuthToken()
	if err != nil {
		return nil, err
	}

	buildClient, err := getBuildClient()
	if err != nil {
		return nil, err
	}

	ns := os.Getenv("NAMESPACE")
	if ns == "" {
		return nil, errors.New("cannot create OpenShift client. no namespace present")
	}

	operatorHost := os.Getenv("OPERATOR_HOSTNAME")
	if operatorHost == "" {
		return nil, errors.New("no hostname available to set required annotations")

	}
	return &OpenShiftClient{token, buildClient, jc, ns, operatorHost}, nil
}

func getAuthToken() (string, error) {
	b, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token") // just pass the file name
	if err != nil {
		return "", errors.New("error reading service account token " + err.Error())
	}
	return string(b), nil // convert content to a 'string'
}

func getBuildClient() (*buildv1.BuildV1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	return buildv1.NewForConfig(config)
}
