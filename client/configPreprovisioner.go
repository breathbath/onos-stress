//Package onos Wraps ONOS client communications
package onos

import (
	"encoding/json"
	"fmt"
	"github.com/breathbath/go_utils/utils/config"
	log "github.com/sirupsen/logrus"
)

const (
	lldpLinkProviderServiceName = "org.onosproject.provider.lldp.impl.LldpLinkProvider"
	oltFlowServiceName          = "org.opencord.olt.impl.OltFlowService"
)

//ConfigPreProvisioner preprovisions data in ONOS
type ConfigPreProvisioner struct {
	onosClient     Client
	opts           *options.ParameterBag
	configProvider ConfigProvider
}

//NewConfigPreProvisioner constructor
func NewConfigPreProvisioner(
	onosClient Client,
	opts *options.ParameterBag,
	configProvider ConfigProvider,
) *ConfigPreProvisioner {
	return &ConfigPreProvisioner{
		onosClient:     onosClient,
		opts:           opts,
		configProvider: configProvider,
	}
}

//PreProvision starting function
func (mbpp *ConfigPreProvisioner) PreProvision() error {
	err := mbpp.processConfiguration(lldpLinkProviderServiceName, "ONOS_COMPONENT_LLDPLINKPROVIDER_CONFIG_FILE")
	if err != nil {
		return err
	}
	err = mbpp.processConfiguration(oltFlowServiceName, "ONOS_COMPONENT_OLTFLOWSERVICE_CONFIG_FILE")
	if err != nil {
		return err
	}

	return nil
}

func (mbpp *ConfigPreProvisioner) processConfiguration(configurationName, locationName string) (err error) {
	log.Infof("incoming request for ONOS configuration %s", configurationName)
	dataLocation, err := mbpp.opts.ReadRequiredString(locationName)
	if err != nil {
		return err
	}

	dataToProvision, err := mbpp.configProvider.GetConfig(dataLocation)
	if err != nil {
		return err
	}

	if len(dataToProvision) == 0 {
		log.Infof("nothing to provision config %s from location %s", configurationName, dataLocation)
		return nil
	}
	log.Debugf("Successfully read configuration %s from data source, which should be provisioned: %s", configurationName, string(dataToProvision))

	var dataToProvisionMap map[string]interface{}
	err = json.Unmarshal(dataToProvision, &dataToProvisionMap)
	if err != nil {
		return fmt.Errorf("cannot convert the configuration to provision %s to a hash map, probably it has a wrong format: %w", string(dataToProvision), err)
	}
	if len(dataToProvisionMap) == 0 {
		log.Infof("nothing to provision config %s from location %s", configurationName, dataLocation)
		return nil
	}

	shouldProvision, err := mbpp.shouldProvision(dataToProvision, dataToProvisionMap, configurationName)
	if err != nil {
		return err
	}

	if !shouldProvision {
		log.Infof("configuration %s in ONOS is expected, no provisioning will be triggered", configurationName)
		return nil
	}

	err = mbpp.onosClient.PostConfiguration(configurationName, dataToProvision)
	if err != nil {
		return err
	}

	existingData, found, err := mbpp.onosClient.ReadConfiguration(configurationName)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("configuration %s was not found after a successful posting, check ONOS logs", configurationName)
	}
	log.Infof("ONOS configuration success")
	log.Debugf("config name: %s, ONOS confirmation response: %s", configurationName, string(existingData))

	return
}

func (mbpp *ConfigPreProvisioner) shouldProvision(dataToProvision []byte, dataToProvisionMap map[string]interface{}, configurationName string) (bool, error) {
	existingData, found, err := mbpp.onosClient.ReadConfiguration(configurationName)
	if err != nil {
		return false, err
	}

	if !found {
		return true, nil
	}

	if len(existingData) == 0 {
		return true, nil
	}

	var existingDataMap map[string]map[string]interface{}
	err = json.Unmarshal(existingData, &existingDataMap)
	if err != nil {
		return false, fmt.Errorf("cannot convert the configuration from ONOS %s to a hash map, probably it has a wrong format: %w", string(existingData), err)
	}

	isChanged := mbpp.isConfigurationChanged(dataToProvisionMap, existingDataMap, configurationName)

	if isChanged {
		log.Debugf("configuration %s from ONOS %s differs from the expected configuration %s", configurationName, string(existingData), string(dataToProvision))
	} else {
		log.Infof("configuration %s should not be provisioned as it already exists in ONOS", configurationName)
	}

	return isChanged, nil
}

func (mbpp *ConfigPreProvisioner) isConfigurationChanged(
	configCandidate map[string]interface{},
	configFromONOS map[string]map[string]interface{},
	configName string,
) bool {
	configItemFromONOS, ok := configFromONOS[configName]
	if !ok {
		return true
	}

	for configCandidateName, configCandidateValue := range configCandidate {
		existingVal, existingValOk := configItemFromONOS[configCandidateName]
		if !existingValOk {
			return true
		}
		existingValStr := fmt.Sprint(existingVal)
		configCandidateValueStr := fmt.Sprint(configCandidateValue)
		if existingValStr != configCandidateValueStr {
			return true
		}
	}

	return false
}
