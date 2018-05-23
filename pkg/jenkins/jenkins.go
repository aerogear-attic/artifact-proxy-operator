package jenkins

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Artifact struct {
	RelativePath string `json:"relativePath"`
}

type JenkinsBuildInfo struct {
	Result    string     `json:"result"`
	Artifacts []Artifact `json:"artifacts"`
	Timestamp int64      `json:"timestamp"`
}

type JenkinsClient struct {
	client *http.Client
}

func (c *JenkinsClient) GetBuildInfo(buildUrl string, authToken string) (*JenkinsBuildInfo, error) {
	var buildStatus *JenkinsBuildInfo
	api := buildUrl + "api/json"
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, errors.New("request failed to Jenkins build api " + err.Error())
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	res, err := c.client.Do(req)
	defer res.Body.Close()

	if err != nil {
		return nil, errors.New("error parsing response from Jenkins for build " + err.Error())
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&buildStatus); err != nil {
		return nil, errors.New("failed to build info response from Jenkins - " + err.Error())
	}
	return buildStatus, nil
}

func (c *JenkinsClient) StreamArtifact(location string, token string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to create jenkins download request %s", err.Error()))
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unexpected error making GET request to Jenkins %s", err.Error()))
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected response code from Jenkins download " + res.Status)
	}
	// hand body back to caller to be closed
	return res.Body, nil
}

func NewJenkinsClient() *JenkinsClient {
	return &JenkinsClient{generateClient()}
}

func generateClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	client.Timeout = time.Second * time.Duration(5)
	return client
}
