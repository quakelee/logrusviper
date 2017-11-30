# logrusviper
A Module to help load [Logrus](https://github.com/sirupsen/logrus) configurations from [spf13/Viper](https://github.com/spf13/viper).
Supply a simple way to set logrus by viper. Logrus configurations can set with other config entries, no need to have a standalone
configuration for it

## Why?
[Logrus](https://github.com/sirupsen/logrus) is great, but can't parse configuration from file or other sources, have to code it.
Nowaday tons of apps are coded with [spf13/Viper](https://github.com/spf13/viper), get configurations from file or other sources,
but there is no helper can parse out [Logrus](https://github.com/sirupsen/logrus) from source and set up it.

## Howto use it
```go
import (
	"github.com/quakelee/logrusviper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".dbutil" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dbutil")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
		logrusviper.SetLogrusConfig(log.StandardLogger(), viper.GetViper())
	} else {
		fmt.Println(err.Error())
	}
}

func main() {
	initConfig()
	logrus.WithFields(logrus.Fields{
		"animal": "walrus",
	}).Error("A walrus appears")
}
```

