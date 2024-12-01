package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	entity "github.com/willychavez/otel-weather-service/app/internal/entities"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type ViacepRepository struct {
	client *http.Client
}

func NewViacepRepository(client *http.Client) entity.CityRepo {
	return &ViacepRepository{client: client}
}

func (v *ViacepRepository) GetCity(ctx context.Context, zipCode string) (string, error) {
	tracer := otel.GetTracerProvider().Tracer("weather-service")
	_, span := tracer.Start(ctx, "GetCity")
	defer span.End()

	resp, err := v.client.Get("https://viacep.com.br/ws/" + zipCode + "/json/")
	if err != nil || resp.StatusCode != http.StatusOK {
		span.RecordError(err)
		return "", errors.New("error getting city")
	}
	defer resp.Body.Close()

	Body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(Body, &result)

	if city, ok := result["localidade"].(string); ok {
		span.SetAttributes(attribute.String("city", city))
		return city, nil
	}
	span.SetAttributes(attribute.String("city", "not found"))
	return "", errors.New("city not found")
}
