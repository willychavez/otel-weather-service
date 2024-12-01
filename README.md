# otel-weather-service
A Go-based system that retrieves the current weather by ZIP code, returning the city and temperature in Celsius, Fahrenheit, and Kelvin. It integrates viaCEP and WeatherAPI, supports OpenTelemetry for tracing, and uses Zipkin for trace visualization. 

### Service A

- Receive input via HTTP POST with the schema:
  ```json
  { "cep": "29902555" }
  ```
- Validate if the input is an 8-digit string.
  - If invalid:
    - **HTTP 422**
    - Response: `{ "message": "invalid zipcode" }`
  - If valid:
    - Forward the CEP to Service B via HTTP.

### Service B

- Receive a valid 8-digit CEP.
- Query the CEP location using the viaCEP API.
- Query the location temperatures using the WeatherAPI.
- Format and return the response:
  - **HTTP 200**
    ```json
    { "city": "São Paulo", "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5 }
    ```
  - **HTTP 422** (Invalid CEP with correct format):
    ```json
    { "message": "invalid zipcode" }
    ```
  - **HTTP 404** (CEP not found):
    ```json
    { "message": "can not find zipcode" }
    ```

### OTEL + Jeager + Zipkin Implementation

- Distributed tracing between Service A and Service B.
- Access tracing dashboards:
  - [Zipkin](http://localhost:9411)
  - [Jaeger](http://localhost:16686)
- Measure response time:
  - For CEP lookup.
  - For temperature lookup.


### How to Use

1. Copy the `docker-compose.yaml.dist` file to `docker-compose.yaml`.
2. Add your WeatherAPI key on line 50 of the `docker-compose.yaml` file:
   ```yaml
   WEATHER_API_KEY: your_api_key_here
   ```
3. Build and run the services with Docker Compose:
   ```bash
   docker compose up --build
   ```
4. Inside the `api` folder, there are test requests available to validate the services.