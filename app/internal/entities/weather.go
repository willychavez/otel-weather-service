package entity

import "context"

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type CityRepo interface {
	GetCity(ctx context.Context, zipCode string) (string, error)
}

type WeatherRepo interface {
	GetWeather(ctx context.Context, city string) (float64, error)
}
