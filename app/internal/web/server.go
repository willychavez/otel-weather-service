package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/willychavez/otel-weather-service/app/internal/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type TemplateData struct {
	ExternalCallMethod string
	ExternalCallURL    string
	RequestNameOTEL    string
	OTELTracer         trace.Tracer
}

type zipCode struct {
	ZipCode string `json:"cep"`
}

type WebServer struct {
	TemplateData *TemplateData
	UseCase      *usecase.WeatherUseCase
}

func NewWebServer(templateData *TemplateData, useCase *usecase.WeatherUseCase) *WebServer {
	return &WebServer{
		TemplateData: templateData,
		UseCase:      useCase,
	}
}

func (w *WebServer) CreateServer() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Post("/", w.ValidationZipcode)
	r.Get("/{zipcode}", w.GetWeather)

	return r
}

func (h *WebServer) ValidationZipcode(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)
	ctx, span := h.TemplateData.OTELTracer.Start(ctx, h.TemplateData.RequestNameOTEL)
	defer span.End()

	var zipcode zipCode
	var err error

	err = json.NewDecoder(r.Body).Decode(&zipcode)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to decode request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isValidZipcode(zipcode.ZipCode) {
		span.SetStatus(codes.Error, "Invalid zipcode")
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	req, err := createExternalRequest(ctx, "GET", h.TemplateData.ExternalCallURL+"/"+zipcode.ZipCode, nil)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to create request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		span.SetStatus(codes.Error, "External call failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, fmt.Sprintf("External call returned status %d", resp.StatusCode))

		externalError, _ := io.ReadAll(resp.Body)
		http.Error(w, string(externalError), resp.StatusCode)
		return
	}

	response, err := handleExternalResponse(resp)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to read response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

func (h *WebServer) GetWeather(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), carrier)
	ctx, span := h.TemplateData.OTELTracer.Start(ctx, h.TemplateData.RequestNameOTEL)
	defer span.End()

	zipcode := chi.URLParam(r, "zipcode")

	if !isValidZipcode(zipcode) {
		span.SetStatus(codes.Error, "Invalid zipcode")
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	weather, err := h.UseCase.GetWeatherByZipCode(ctx, zipcode)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to get weather")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}

func isValidZipcode(zipcode string) bool {
	validate := regexp.MustCompile(`^[0-9]{8}$`)
	return validate.MatchString(zipcode)
}

func createExternalRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	return req, nil
}

func handleExternalResponse(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}
