package initializers

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"quote/config"
	"quote/workers/sneakerWorkers"
)

func InitWorkers() {
	pathStr, _ := filepath.Abs("config/workers.yml")
	content, err := ioutil.ReadFile(pathStr)
	if err != nil {
		log.Fatal(err)
	}
	yaml.Unmarshal(content, &config.AllWorkers)
	sneakerWorkers.InitializeKLineBuildWorker()
	sneakerWorkers.InitializeSubQuoteBuildWorker()
}
