package mangad

import (
	"os"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Title   string `json:"title"`
	Url     string `json:"url"`
	Source  int64  `json:"source"`
	PkgName string `json:"pkgName"`
}

func LoadConfig(filename string) (map[int64][]Config, error) {
	flatConfig, err := loadFromFile(filename)
	if err != nil {
		return nil, err
	}
	return groupBySource(flatConfig), nil
}

func loadFromFile(filename string) ([]Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := []Config{}
	if err := yaml.UnmarshalStrict(file, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func groupBySource(flatConfig []Config) map[int64][]Config {
	config := map[int64][]Config{}

	for _, conf := range flatConfig {
		c, ok := config[conf.Source]
		if !ok {
			c = []Config{}
		}

		config[conf.Source] = append(c, conf)
	}

	return config
}
