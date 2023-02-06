package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)


var (
	defaultConfig = []byte(`
mappings:
  kubernetes_state_container_memory_requested: kubernetes_state.container.memory_requested
  kube_pod_container_resource_requests: kubernetes_state.container.memory_requested
ignoreLabels:
  - resource
  - unit
addLabels:
  clustername=thebest
`)
)

func InitConfig() error {
	var err error
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("DRA")
	var cfg []byte = nil
	cfg = defaultConfig
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	err = viper.ReadConfig(bytes.NewBuffer(cfg))
	if err != nil {
		return err
	}

	if viper.GetString("custom.mapping.location") != "" {
		fileout, err := os.ReadFile(viper.GetString("custom.mapping.location"))
		if err != nil {
			return fmt.Errorf("error reading mapping file: %w ", err)
		}
		buffer := bytes.NewBuffer(fileout)
		err = viper.MergeConfig(buffer)
		if err != nil {
			return err
		}
	}


	return err
}

