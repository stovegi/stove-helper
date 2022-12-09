package config

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	LoggingDir string     `mapstructure:"loggingDir"`
	DataConfig DataConfig `mapstructure:"dataConfig"`
	Device     string     `mapstructure:"device"`
	Seed       uint64     `mapstructure:"seed"`
	ServerSeed uint64     `mapstructure:"serverSeed"`
}

type DataConfig struct {
	CmdIDPath      string `mapstructure:"cmdIdPath"`
	ProtoPath      string `mapstructure:"protoPath"`
	OutputPath     string `mapstructure:"outputPath"`
	DispatchRegion string `mapstructure:"dispatchRegion"`
	PrivateKeyPath string `mapstructure:"privateKeyPath"`
}

var DefaultConfig = Config{
	LoggingDir: "log",
	DataConfig: DataConfig{
		CmdIDPath: "data/Sorapointa-Protos/cmdid.csv",
		ProtoPath: "data/Sorapointa-Protos/proto",
	},
}

func LoadConfig() (cfg Config) { return LoadConfigName("config") }

func LoadConfigName(name string) (cfg Config) {
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/stove-helper")
	viper.AddConfigPath("$HOME/.stove-helper")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("Config file not found, using the default config")
			cfg = DefaultConfig
			initLogger(cfg.LoggingDir)
			return
		} else {
			log.Panic().Err(err).Msg("Failed to read config file")
		}
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Panic().Err(err).Msg("Failed to decode config file")
	}
	if cfg.Device == "" {
		log.Panic().Msg("Device is not set")
	}
	if cfg.DataConfig.CmdIDPath == "" || cfg.DataConfig.ProtoPath == "" {
		log.Panic().Msg("CmdIDPath or ProtoPath is not set")
	}
	if cfg.DataConfig.OutputPath == "" {
		cfg.DataConfig.OutputPath = "data/output"
	}
	if cfg.DataConfig.DispatchRegion == "" {
		log.Panic().Msg("DispatchRegion is not set")
	}
	if cfg.DataConfig.PrivateKeyPath == "" {
		log.Panic().Msg("PrivateKeyPath is not set")
	}
	initLogger(cfg.LoggingDir)
	return
}

func initLogger(dir string) {
	if dir != "" {
		log.Logger = log.Output(io.MultiWriter(zerolog.ConsoleWriter{Out: os.Stderr}, newRollingFile(dir))).With().Caller().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	}
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
}

func newRollingFile(dir string) io.Writer {
	if err := os.MkdirAll(dir, 0744); err != nil {
		log.Error().Err(err).Str("path", "log").Msg("can't create log directory")
		return nil
	}
	return &lumberjack.Logger{
		Filename: path.Join(dir, fmt.Sprintf("stove-helper-%s.log", time.Now().Format("2006-01-02"))),
	}
}
