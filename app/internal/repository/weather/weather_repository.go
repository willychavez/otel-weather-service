package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
	entity "github.com/willychavez/otel-weather-service/app/internal/entities"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type weatherRepository struct {
	client *http.Client
}

func NewWeatherRepository(client *http.Client) entity.WeatherRepo {
	return &weatherRepository{client: client}
}

func (repo *weatherRepository) GetWeather(ctx context.Context, city string) (float64, error) {
	trace := otel.GetTracerProvider().Tracer("weather-service")
	_, span := trace.Start(ctx, "GetWeather")
	defer span.End()

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", viper.GetString("WEATHER_API_KEY"), city)

	resp, err := repo.client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		span.RecordError(err)
		return 0, errors.New("failed to fetch weather data")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if current, ok := result["current"].(map[string]interface{}); ok {
		if tempC, ok := current["temp_c"].(float64); ok {
			span.SetAttributes(attribute.Float64("tempC", tempC))
			return tempC, nil
		}
	}

	span.SetAttributes(attribute.String("tempC", "not found"))
	return 0, errors.New("temperature not found")

}
