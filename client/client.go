//Package onos Wraps ONOS client communications
package onos

import (
	"context"
	"fmt"
	http2 "github.com/breathbath/go_utils/utils/http"
	"github.com/breathbath/go_utils/utils/rest"
	error2 "github.com/breathbath/onos-stress/error"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

//Client communication point to ONOS API
type Client struct {
	ONOSLogin        string
	ONOSPass         string
	ConfigurationURL string
	MetersURL        string
}

//NewClient ONOS client constructor
func NewClient(onosAddress, onosLogin, onosPass string) (Client, error) {
	metersURL, err := buildEndpointURL(onosAddress, "/onos/v1/meters")
	if err != nil {
		return Client{}, err
	}

	configurationURL, err := buildEndpointURL(onosAddress, "/onos/v1/configuration")
	if err != nil {
		return Client{}, err
	}

	return Client{
		ONOSLogin:        onosLogin,
		ONOSPass:         onosPass,
		MetersURL:        metersURL,
		ConfigurationURL: configurationURL,
	}, nil
}

func buildEndpointURL(onosAddress, uri string) (string, error) {
	oltAppURL, err := url.Parse(onosAddress)
	if err != nil {
		return "", err
	}

	oltAppURL.Path = oltAppURL.Path + "/" + uri

	return oltAppURL.String(), nil
}

//GetHealth checks health of ONOS API
func (c Client) GetHealth(ctx context.Context) error {
	log.Debugf("Will check ONOS health at %s", c.MetersURL)
	_, err, _ := c.callONOSAPI(c.MetersURL, "GET", "")
	if err == nil {
		log.Debugf("ONOS health is OK")
	}
	return err
}

//PostConfiguration post new configuration to ONOS
func (c Client) PostConfiguration(configName string, dataIn []byte) (err error) {
	log.Infof("Will post configuration %s to ONOS", configName)
	log.Debugf("Configuration data %s to be provisioned: %s", configName, string(dataIn))

	_, err, _ = c.callONOSAPI(c.ConfigurationURL + "/" + configName+"?preset=true", "POST", string(dataIn))
	if err != nil {
		return err
	}

	log.Infof("Configuration %s successfully posted to ONOS", configName)

	return
}

//ReadConfiguration reads configuration from ONOS API
func (c Client) ReadConfiguration(configName string) (dataOut []byte, found bool, err error) {
	log.Debugf("Will read configuration %s from ONOS", configName)

	found = true
	dataOut, err, _ = c.callONOSAPI(c.ConfigurationURL + "/" + configName, "GET", "")
	if err != nil {
		return []byte{}, false, err
	}

	log.Debugf("Got configuration data for %s from ONOS: %s", configName, string(dataOut))

	return
}

func (c Client) callONOSAPI(url, method, body string) ([]byte, error, *http.Response) {
	client := rest.NewJsonClient()
	log.Debugf("Will call %s for ONOS API at [%s]", method, url)

	reqContext := rest.RequestContext{
		TargetUrl: url,
		Method:    method,
		Body:      body,
		Headers:   map[string]string{},
	}
	if c.ONOSLogin != "" && c.ONOSPass != "" {
		reqContext.Headers["Authorization"] = "Basic " + http2.BuildBasicAuthString(c.ONOSLogin, c.ONOSPass)
	}
	data, err, resp := client.CallApi(reqContext)
	if err != nil {
		err = error2.ApiError{
			Msg: fmt.Sprintf("ONOS API request failed: %v", err),
		}
	}

	return data, err, resp
}
