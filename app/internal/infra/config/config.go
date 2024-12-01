package config

import "github.com/spf13/viper"

type config struct {
	ExternalCallURL    string `env:"EXTERNAL_CALL_URL"`
	ExternalCallMethod string `env:"EXTERNAL_CALL_METHOD"`
	OTLPExporteEndpont string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELServiceName    string `env:"OTEL_SERVICE_NAME"`
	OTELResquestName   string `env:"OTEL_REQUEST_NAME"`
	ZipkinEndpoint     string `env:"ZIPKIN_ENDPOINT"`
	HttpPort           string `env:"HTTP_PORT"`
	WeatherApiKey      string `env:"WEATHER_API_KEY"`
}

func LoadConfig(path string) (*config, error) {
	var cfg *config
	// viper.SetConfigName("config")
	// viper.SetConfigType("env")
	// viper.AddConfigPath(path)
	// viper.SetConfigFile("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func Init() {
	viper.AutomaticEnv()
}
