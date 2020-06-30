package main

import (
	"context"
	options "github.com/breathbath/go_utils/utils/config"
	onos "github.com/breathbath/onos-stress/client"
	error2 "github.com/breathbath/onos-stress/error"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

func init() {
	log.SetLevel(log.DebugLevel)

	customFormatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(customFormatter)
}

func main() {
	params := options.New(options.EnvValuesProvider{})
	onosLogin := params.ReadString("ONOS_API_LOGIN", "")
	onosPass := params.ReadString("ONOS_API_PASS", "")
	onosAddr, err := params.ReadRequiredString("ONOS_API_ADDRESS")
	if err != nil {
		log.Fatal(err)
	}

	onosClient, err := onos.NewClient(onosAddr, onosLogin, onosPass)
	if err != nil {
		log.Fatal(err)
	}

	configProvider := onos.ConfigFileSystemProvider{}

	configPreprovisioner := onos.NewConfigPreProvisioner(
		onosClient,
		&params,
		configProvider,
	)
	ctx, cancel := context.WithCancel(context.Background())

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for {
			sleepDurAfterFailure, err1 := time.ParseDuration(params.ReadString("SLEEP_AFTER_FAILURE", "1s"))
			sleepDurAfterSuccess, err2 := time.ParseDuration(params.ReadString("SLEEP_AFTER_SUCCESS", "10s"))
			if err1 != nil {
				log.Error(err1)
				return
			}
			if err2 != nil {
				log.Error(err2)
				return
			}

			err = configPreprovisioner.PreProvision()
			if err != nil {
				apiErr, ok := err.(error2.ApiError)
				if !ok {
					return
				}
				log.Error(apiErr)
			}

			select {
			case <-ctx.Done():
				return
			default:
				if err != nil {
					log.Infof("will sleep after failure %v", sleepDurAfterFailure)
					time.Sleep(sleepDurAfterFailure)
				} else {
					log.Infof("will sleep after success %v", sleepDurAfterSuccess)
					time.Sleep(sleepDurAfterSuccess)
				}
				log.Info("woke up after")
				continue
			}
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		<-stopChan
		cancel()
	}()

	<-doneChan
	log.Info("Exit due to the subprocess exit")
}
