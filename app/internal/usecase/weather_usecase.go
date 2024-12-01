package usecase

import (
	"context"
	"errors"
	"fmt"

	entity "github.com/willychavez/otel-weather-service/app/internal/entities"
)

type WeatherUseCase struct {
	cityRepo    entity.CityRepo
	weatherRepo entity.WeatherRepo
}

func NewWeatherUseCase(cityRepo entity.CityRepo, weatherRepo entity.WeatherRepo) *WeatherUseCase {
	return &WeatherUseCase{
		cityRepo:    cityRepo,
		weatherRepo: weatherRepo,
	}
}

func (w *WeatherUseCase) GetWeatherByZipCode(ctx context.Context, zipCode string) (*entity.WeatherResponse, error) {
	city, err := w.cityRepo.GetCity(ctx, zipCode)
	if err != nil {
		return nil, errors.New("can not find city")
	}

	weather, err := w.weatherRepo.GetWeather(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch the current temperature for city: %s", city)
	}

	return &entity.WeatherResponse{
		City:  city,
		TempC: weather,
		TempF: weather * 1.8,
		TempK: weather + 273.15,
	}, nil
}
