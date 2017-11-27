/*
   Copyright 2017 Atos

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package main

import (
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	configPrefix          string        = "sla"
	defaultPort           string        = "8090"
	defaultCheckPeriod    time.Duration = 60
	defaultRepositoryType string        = "memory"

	portPropertyName           = "port"
	checkPeriodPropertyName    = "checkPeriod"
	repositoryTypePropertyName = "repository"
	singleFilePropertyName     = "singlefile"

	unixConfigPath = "/etc/slalite:."
	configName     = "slalite"
)

func main() {

	// TODO: Add windows path
	configPath := flag.String("d", unixConfigPath, "Directories where to search config files")
	configBasename := flag.String("b", configName, "Filename (w/o extension) of config file")
	configFile := flag.String("f", "", "Path of configuration file. Overrides -b and -d")
	flag.Parse()

	log.Println("Initializing")
	config := createMainConfig(configFile, configPath, configBasename)
	logMainConfig(config)

	singlefile := config.GetBool(singleFilePropertyName)
	port := config.GetString(portPropertyName)
	checkPeriod := config.GetDuration(checkPeriodPropertyName)
	repoType := config.GetString(repositoryTypePropertyName)

	var repo model.IRepository = nil
	var repoconfig *viper.Viper = nil
	if singlefile {
		repoconfig = config
	}
	switch repoType {
	case defaultRepositoryType:
		repo = memrepository.MemRepository{}
	case "mongodb":
		mongoRepo, errMongo := mongodb.New(repoconfig)
		if errMongo != nil {
			log.Fatal("Error creating mongo repository: ", errMongo.Error())
		}
		repo = mongoRepo
	}
	if repo != nil {
		a := App{}
		a.Initialize(repo)
		go createValidationThread(repo, checkPeriod)
		a.Run(":" + port)
	}
}

//
// Creates the main Viper configuration.
// file: if set, is the path to a configuration file. If not set, paths and basename will be used
// paths: colon separated paths where to search a config file
// basename: basename of a configuration file accepted by Viper (extension is automatic)
//
func createMainConfig(file *string, paths *string, basename *string) *viper.Viper {
	config := viper.New()

	config.SetEnvPrefix(configPrefix) // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.SetDefault(portPropertyName, defaultPort)
	config.SetDefault(checkPeriodPropertyName, defaultCheckPeriod)
	config.SetDefault(repositoryTypePropertyName, defaultRepositoryType)

	if *file != "" {
		config.SetConfigFile(*file)
	} else {
		config.SetConfigName(*basename)
		for _, path := range strings.Split(*paths, ":") {
			config.AddConfigPath(path)
		}
	}

	errConfig := config.ReadInConfig()
	if errConfig != nil {
		log.Println("Can't find configuration file: " + errConfig.Error())
		log.Println("Using defaults")
	}
	return config
}

func logMainConfig(config *viper.Viper) {

	port := config.GetString(portPropertyName)
	checkPeriod := config.GetDuration(checkPeriodPropertyName)
	repoType := config.GetString(repositoryTypePropertyName)

	log.Printf("SLALite initialization\n"+
		"\tConfigfile: %s\n"+
		"\tRepository type: %s\n"+
		"\tPort: %s\n"+
		"\tCheck period:%d\n",
		config.ConfigFileUsed(), repoType, port, checkPeriod)
}

func createValidationThread(repo model.IRepository, checkPeriod time.Duration) {
	ticker := time.NewTicker(checkPeriod * time.Second)

	for {
		<-ticker.C
		validateProviders(repo)
	}

}

func validateProviders(repo model.IRepository) {
	providers, err := repo.GetAllProviders()

	if err == nil {
		log.Println("There are " + strconv.Itoa(len(providers)) + " providers")
	} else {
		log.Println("Error: " + err.Error())
	}
}