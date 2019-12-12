package logrusviper

import (
	"log/syslog"
	"os"
	"reflect"
	"strings"

	"github.com/johntdyer/slackrus"
	"github.com/sirupsen/logrus"
	syslog_hook "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/viper"
)

// Hook defined logrus hooks configurations
type Hook map[string]interface{}

// Hooks is abbreviation of Hook array
type Hooks []Hook

// SetLogrusConfig get all defines from viper, set to logger
func SetLogrusConfig(logger *logrus.Logger, viper *viper.Viper) {
	if viper.IsSet("logger.level") {
		switch strings.ToLower(viper.GetString("logger.level")) {
		case "debug":
			logger.SetLevel(logrus.DebugLevel)
		case "info":
			logger.SetLevel(logrus.InfoLevel)
		case "warn":
			logger.SetLevel(logrus.WarnLevel)
		case "error":
			logger.SetLevel(logrus.ErrorLevel)
		case "fatal":
			logger.SetLevel(logrus.FatalLevel)
		case "panic":
			logger.SetLevel(logrus.PanicLevel)
		default:
			logger.SetLevel(logrus.DebugLevel)
		}
	}
	if viper.IsSet("logger.out.name") {
		switch strings.ToLower(viper.GetString("logger.out.name")) {
		case "stdout":
			logger.Out = os.Stdout
		case "stderr":
			logger.Out = os.Stderr
		case "file":
			if options := viper.GetStringMapString("logger.out.options"); options != nil {
				fn := "dbutil.log"
				if filename, isSet := options["filename"]; isSet {
					fn = filename
				}
				f, err := os.OpenFile(fn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
				if err == nil {
					logger.Out = f
				} else {
					logger.Errorln("Failed to open log file, the reason is ", err.Error())
				}
			}
		}
	}
	if viper.IsSet("logger.formatter.name") {
		switch strings.ToLower(viper.GetString("logger.formatter.name")) {
		case "text":
			if viper.IsSet("logger.formatter.options") {
				var (
					fc, dc    string
					formatter logrus.TextFormatter
				)
				options := viper.GetStringMapString("logger.formatter.options")
				if fc, isSet := options["forcecolors"]; isSet {
					if strings.ToLower(fc) == "true" {
						formatter.ForceColors = true
					}
				}
				if dc, isSet := options["disablecolors"]; isSet {
					if strings.ToLower(dc) == "true" {
						formatter.DisableColors = true
					}
				}
				if strings.ToLower(fc) == "true" && strings.ToLower(dc) == "true" {
					logger.Errorln("forcecolors and disablecolors can't be true same time")
				}
				if dt, isSet := options["disabletimestamp"]; isSet {
					if strings.ToLower(dt) == "true" {
						formatter.DisableTimestamp = true
					}
				}
				if ft, isSet := options["fulltimestamp"]; isSet {
					if strings.ToLower(ft) == "true" {
						formatter.FullTimestamp = true
					}
				}
				if ds, isSet := options["disablesorting"]; isSet {
					if strings.ToLower(ds) == "true" {
						formatter.DisableSorting = true
					}
				}
				if qef, isSet := options["quoteemptyfields"]; isSet {
					if strings.ToLower(qef) == "true" {
						formatter.QuoteEmptyFields = true
					}
				}
				if tf, isSet := options["timestampformat"]; isSet {
					formatter.TimestampFormat = tf
				}
				logger.Formatter = &formatter
			} else {
				logger.Formatter = &logrus.TextFormatter{}
			}
		case "json":
			if viper.IsSet("logger.formatter.options") {
				var formatter logrus.JSONFormatter
				options := viper.GetStringMapString("logger.formatter.options")
				if dt, isSet := options["disabletimestamp"]; isSet {
					if strings.ToLower(dt) == "true" {
						formatter.DisableTimestamp = true
					}
				}
				if tf, isSet := options["timestampformat"]; isSet {
					formatter.TimestampFormat = tf
				}
				if fm := viper.GetStringMapString("logger.formatter.options.fieldmap"); fm != nil {
					cfm := make(logrus.FieldMap)
					if fkt, isSet := fm["fieldkeytime"]; isSet {
						cfm[logrus.FieldKeyTime] = fkt
					}
					if fkl, isSet := fm["fieldkeylevel"]; isSet {
						cfm[logrus.FieldKeyLevel] = fkl
					}
					if fkm, isSet := fm["fieldkeymsg"]; isSet {
						cfm[logrus.FieldKeyMsg] = fkm
					}
					formatter.FieldMap = cfm
				}
				logger.Formatter = &formatter
			} else {
				logger.Formatter = &logrus.JSONFormatter{}
			}
		}
	}
	if viper.IsSet("logger.hooks") {
		var (
			hooks                                     Hooks
			al                                        logrus.Level
			sl                                        syslog.Priority
			alstr, hookurl, channel, icon             string
			username, levelstr, protocol, target, tag string
		)
		validFormat := make(map[string]interface{})
		err := viper.UnmarshalKey("logger.hooks", &hooks)
		if err != nil {
			logger.Errorln("Failed to parse slackrus hooks settings:", err.Error())
		}
		for _, hook := range hooks {
			switch strings.ToLower(hook["name"].(string)) {
			case "slackrus":
				if reflect.TypeOf(hook["options"]) == reflect.TypeOf(validFormat) {
					options := hook["options"].(map[string]interface{})
					hookurl = options["hookurl"].(string)
					alstr = strings.ToLower(options["acceptedlevels"].(string))
					channel = options["channel"].(string)
					icon = options["iconemoji"].(string)
					username = options["username"].(string)
				} else {
					logger.Panicln("Unexpected slackrus options format")
				}
				switch alstr {
				case "debug":
					al = logrus.DebugLevel
				case "info":
					al = logrus.InfoLevel
				case "warn":
					al = logrus.WarnLevel
				case "error":
					al = logrus.ErrorLevel
				case "fatal":
					al = logrus.FatalLevel
				case "panic":
					al = logrus.PanicLevel
				default:
					al = logrus.DebugLevel
				}
				logger.AddHook(&slackrus.SlackrusHook{
					HookURL:        hookurl,
					AcceptedLevels: slackrus.LevelThreshold(al),
					Channel:        channel,
					IconEmoji:      icon,
					Username:       username,
				})
			case "syslog":
				if reflect.TypeOf(hook["options"]) == reflect.TypeOf(validFormat) {
					options := hook["options"].(map[string]interface{})
					protocol = strings.ToLower(options["protocol"].(string))
					target = options["target"].(string)
					levelstr = options["level"].(string)
					tag = options["tag"].(string)
				} else {
					logger.Panicln("Unexpected syslog options format")
				}

				switch levelstr {
				case "debug":
					sl = syslog.LOG_DEBUG
				case "info":
					sl = syslog.LOG_INFO
				case "warn":
					sl = syslog.LOG_WARNING
				case "error":
					sl = syslog.LOG_ERR
				case "fatal":
					sl = syslog.LOG_CRIT
				case "panic":
					sl = syslog.LOG_EMERG
				default:
					sl = syslog.LOG_DEBUG
				}
				hook, err := syslog_hook.NewSyslogHook(protocol, target, sl, tag)
				if err != nil {
					logger.Errorln("Failed to set syslog hook:", err.Error())
				}
				logger.AddHook(hook)
			}
		}
	}
}
