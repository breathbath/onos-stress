//Package onos Wraps ONOS client communications
package onos

import (
	"io/ioutil"
	"path/filepath"
)

//ConfigProvider interface to provide data for ONOS configuration
type ConfigProvider interface {
	GetConfig(location string) ([]byte, error)
}

//ConfigFileSystemProvider fs implementation for the ConfigProvider
type ConfigFileSystemProvider struct{}

//GetConfig read downstream data from file implementing the ConfigProvider
func (cfsp ConfigFileSystemProvider) GetConfig(location string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Clean(location))
	return data, err
}
