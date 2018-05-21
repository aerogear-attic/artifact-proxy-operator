package main

import (
	"encoding/json"
	"log"
	"os"

	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apibuildv1 "github.com/openshift/api/build/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
)

const (
	requestURLAnnotation   = "MOBILE_ARTIFACT_DOWNLOAD"
	provideURLAnnotation   = "MOBILE_ARTIFACT_URL"
	provideTokenAnnotation = "MOBILE_ARTIFACT_TOKEN"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	buildV1Client, err := buildv1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	events, err := buildV1Client.Builds(os.Getenv("NAMESPACE")).Watch(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for update := range events.ResultChan() {
		raw, _ := json.Marshal(update.Object)
		var build = apibuildv1.Build{}
		json.Unmarshal(raw, &build)
		//artifact download url requested
		if val, ok := build.Annotations[requestURLAnnotation]; ok && val == "true" {
			//and not provided yet
			if _, ok := build.Annotations[provideURLAnnotation]; !ok {
				addURL(&build, buildV1Client)
				log.Printf("Download requested for %v\n", build.ObjectMeta.Name)
			} else {
				log.Printf("Download already provided for %v\n", build.ObjectMeta.Name)
			}
		} else {
			log.Printf("Download not requested for %v\n", build.ObjectMeta.Name)
		}
	}
}

func addURL(build *apibuildv1.Build, client *buildv1.BuildV1Client) {
	build.Annotations[provideURLAnnotation] = generateURL()
	build.Annotations[provideTokenAnnotation] = generateToken()
	build, err := client.Builds(os.Getenv("NAMESPACE")).Update(build)
	if err != nil {
		log.Printf("err: %+v\n", err)
	}

}

func generateToken() string {
	return "the token"
}

func generateURL() string {
	return "the url"
}
