package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)


func InterfaceMarshall(jsonObj interface{}, targetStruct interface{}) error {
	b, err := json.Marshal(jsonObj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &targetStruct)
	return err
}


func GetZapReleases(executable string) (*ZapReleases, error) {
	// declare the stuff which we are going to return
	zapReleases := &ZapReleases{}

	// get the target URL based on the executable name
	targetUrl := fmt.Sprintf(ZapDefaultConfig.mirror, executable)
	logger.Infof("Fetching %s", targetUrl)

	// create http client and fetch JSON
	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		SetHeader("Accept", "application/json").
		Get(targetUrl)
	if err != nil {
		return zapReleases, err
	}


	var rawResponse interface{}
	err = json.Unmarshal(resp.Body(), &rawResponse)
	if err != nil {
		return zapReleases, err
	}

	rawResponseMap := rawResponse.(map[string]interface{})

	var zapReleasesArray = make([]ZapRelease, 0)
	for k, v := range rawResponseMap {
		switch jsonObj := v.(type) {
		case string:
			if k == "owner" {
				// handle cases when the key defines the owner
				zapReleases.Author = v.(string)
				break
			} else if k == "source" {
				// handle the sources attribute in the json file
				zapSource := &ZapSource{}
				err = InterfaceMarshall(v, zapSource)
				if err != nil {
					return zapReleases, err
				}
				zapReleases.Source = *zapSource
				break
			} else {
				return zapReleases, errors.New("invalid data received")
			}
		case interface{}:
			// handles all other cases
			zapRelease := &ZapRelease{}

			err = InterfaceMarshall(jsonObj, zapRelease)
			if err != nil {
				return zapReleases, nil
			}
			logger.Debug(zapRelease)
			if zapRelease.Id != "" {
				zapRelease.Roll, err = strconv.Atoi(k)
				if err != nil {
					return zapReleases, err
				}

				zapReleasesArray = append(zapReleasesArray, *zapRelease)
			}

		default:
			return zapReleases, errors.New("was expecting JSON, got something else instead")
		}
	}

	// return the array of zap releases
	zapReleases.Releases = zapReleasesArray

	logger.Infof("Found %d releases", len(zapReleases.Releases))
	return zapReleases, nil
}