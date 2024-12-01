package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/viper"
	"github.com/willychavez/otel-weather-service/app/internal/infra/config"
	"github.com/willychavez/otel-weather-service/app/internal/infra/httpclient"
	"github.com/willychavez/otel-weather-service/app/internal/repository/viacep"
	"github.com/willychavez/otel-weather-service/app/internal/repository/weather"
	"github.com/willychavez/otel-weather-service/app/internal/usecase"
	"github.com/willychavez/otel-weather-service/app/internal/web"
	traicing "github.com/willychavez/otel-weather-service/app/pkg"
	"go.opentelemetry.io/otel"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config.Init()

	shutdown, err := traicing.InitProvider(viper.GetString("OTEL_SERVICE_NAME"), viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := shutdown(ctx)
		if err != nil {
			log.Fatal("error shutting down jeager provider: %w", err)
		}
	}()

	httpClient := httpclient.NewHttpClient()
	viacepRepo := viacep.NewViacepRepository(httpClient)
	weatherRepo := weather.NewWeatherRepository(httpClient)

	weatherUseCase := usecase.NewWeatherUseCase(viacepRepo, weatherRepo)

	tracer := otel.Tracer("weather-service")

	templateDta := &web.TemplateData{
		ExternalCallURL: viper.GetString("EXTERNAL_CALL_URL"),
		RequestNameOTEL: viper.GetString("REQUEST_NAME_OTEL"),
		OTELTracer:      tracer,
	}

	server := web.NewWebServer(templateDta, weatherUseCase)
	r := server.CreateServer()

	go func() {
		log.Printf("server listening on port %s", viper.GetString("HTTP_PORT"))
		err = http.ListenAndServe(":"+viper.GetString("HTTP_PORT"), r)
		if err != nil {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+C pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due to other reason...")
	}

	// Create a timeout context for the graceful shutdown
	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}
