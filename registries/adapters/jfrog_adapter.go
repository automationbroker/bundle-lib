//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package adapters

import (
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const jfrogName = "jfrog.io"

// JFrogAdapter - JFrog Adapter
type JFrogAdapter struct {
	Config Configuration
}

// JFrogImageResponse - response for JFrog Registry.
type JFrogImageResponse struct {
	Repositories []string `json:"repositories"`
}

// RegistryName - Retrieve the registry name
func (r JFrogAdapter) RegistryName() string {
	return jfrogName
}

// GetImageNames - retrieve the images
func (r JFrogAdapter) GetImageNames() ([]string, error) {
	log.Debug("JFrogAdapter::GetImages")
	log.Debug("BundleSpecLabel: %s", BundleSpecLabel)
	log.Debug("Loading image list for URL: [ %v ]", r.Config.URL)

	// Basic Auth Base64 username:password
	token := b64.StdEncoding.EncodeToString([]byte(r.Config.User + ":" + r.Config.Pass))

	// Initial Image URL
	url := r.Config.URL.String() + "/v2/_catalog?n=100"

	// Get all APB Image Names until the 'Link' value is empty
	// https://docs.docker.com/registry/spec/api/#pagination
	var apbData []string
	for {
		images, linkStr, err := r.getNextImages(token, url)
		if err != nil {
			return nil, err
		}

		apbData = append(apbData, images.Repositories...)

		// no more to get..
		if len(linkStr) == 0 {
			break
		}

		url = r.getNextImageUrl(r.Config.URL.String(), linkStr)
		if len(url) == 0 {
			return nil, errors.New("Invalid Next Image URL")
		}
	}

	return apbData, nil
}

// getNextImages - will get the next URL.
func (r JFrogAdapter) getNextImages(token string, url string) (*JFrogImageResponse, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		return nil, "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		return nil, "", err
	}
	defer resp.Body.Close()

	// get the repolist
	iResp := JFrogImageResponse{}
	bodyText, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyText, &iResp)
	if err != nil {
		log.Errorf("unable to get next images for url: %v - %v", url, err)
		return nil, "", err
	}

	return &iResp, resp.Header.Get("Link"), nil
}

// getNextImageUrl - returns the next image URL created from the 'Link' in the header
func (r JFrogAdapter) getNextImageUrl(hostUrl string, link string) string {
	if len(hostUrl) == 0 {
		log.Errorf("'HOST URL' is empty")
		return ""
	}

	if len(link) == 0 {
		log.Errorf("'Link' value is empty")
		return ""
	}

	res := strings.Split(link, ";")
	if len(res[0]) == 0 {
		log.Errorf("Invalid Link value")
		return ""
	}

	var lvalue string
	lvalue = strings.TrimSpace(res[0])
	lvalue = strings.Trim(lvalue, "<>")
	return (hostUrl + lvalue)
}

// FetchSpecs - retrieve the spec for the image names.
func (r JFrogAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	specs := []*bundle.Spec{}
	for _, imageName := range imageNames {
		spec, err := r.loadSpec(imageName)
		if err != nil {
			log.Errorf("Failed to retrieve spec data for image %s - %v", imageName, err)
		}
		if spec != nil {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

func (r JFrogAdapter) loadSpec(imageName string) (*bundle.Spec, error) {
	// Basic Auth Base64 username:password
	token := b64.StdEncoding.EncodeToString([]byte(r.Config.User + ":" + r.Config.Pass))

	digest, err := r.getDigest(imageName, token)
	if err != nil {
		return nil, err
	}
	return r.digestToSpec(digest, imageName, token)
}

func (r JFrogAdapter) getDigest(imageName string, token string) (string, error) {
	if r.Config.Tag == "" {
		r.Config.Tag = "latest"
	}

	url := r.Config.URL.String() + "/v2/" + imageName + "/manifests/" + r.Config.Tag
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type config struct {
		Digest string `json:"digest"`
	}

	conf := struct {
		Config config `json:"config"`
	}{}

	// get the manifest
	bodyText, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyText, &conf)
	if err != nil {
		log.Errorf("unable to get digest for image [%s] with url: %v - %v", imageName, url, err)
		return "", err
	}

	return conf.Config.Digest, nil
}

func (r JFrogAdapter) digestToSpec(digest string, imageName string, token string) (*bundle.Spec, error) {
	spec := &bundle.Spec{}
	url := r.Config.URL.String() + "/v2/" + imageName + "/blobs/" + digest
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %v", token))
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Errorf("Unable to authenticate to the registry, registry credentials could be invalid.")
		return nil, nil
	}

	type label struct {
		Spec    string `json:"com.redhat.apb.spec"`
		Runtime string `json:"com.redhat.apb.runtime"`
	}

	type config struct {
		Label label `json:"Labels"`
	}

	conf := struct {
		Config *config `json:"config"`
	}{}

	bodyText, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyText, &conf)
	if err != nil {
		log.Errorf("Unable to get Spec for [%s]: - %v", imageName, err)
		return nil, err
	}

	if conf.Config == nil {
		log.Infof("Did not find config. Skipping image")
		return nil, nil
	}
	if conf.Config.Label.Spec == "" {
		log.Infof("Didn't find encoded Spec label. Assuming image is not APB and skiping")
		return nil, nil
	}

	encodedSpec := conf.Config.Label.Spec
	decodedSpecYaml, err := b64.StdEncoding.DecodeString(encodedSpec)
	if err != nil {
		log.Errorf("Something went wrong decoding spec from label")
		return nil, err
	}

	if err = yaml.Unmarshal(decodedSpecYaml, spec); err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}

	spec.Runtime, err = getAPBRuntimeVersion(conf.Config.Label.Runtime)
	if err != nil {
		return nil, err
	}

	imgTag := r.Config.Tag
	if len(imgTag) == 0 {
		imgTag = "latest"
	}

	spec.Image = r.Config.URL.RequestURI() + "/" + imageName + ":" + imgTag

	log.Debugf("adapter::imageToSpec -> Got plans %+v", spec.Plans)
	log.Debugf("Successfully converted Image %s into Spec", spec.Image)

	return spec, nil
}
