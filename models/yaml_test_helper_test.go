package models

import "gopkg.in/yaml.v2"

func unmarshalYaml(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
