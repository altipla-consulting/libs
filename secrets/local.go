package secrets

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"libs.altipla.consulting/errors"
)

type localSecrets struct {
	Secrets map[string]string `yaml:"secrets"`
}

func readLocalSecrets() (map[string]string, error) {
	content, err := ioutil.ReadFile("secrets.yml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Trace(err)
	}
	var local localSecrets
	if err := yaml.UnmarshalStrict(content, &local); err != nil {
		return nil, errors.Trace(err)
	}
	return local.Secrets, nil
}
