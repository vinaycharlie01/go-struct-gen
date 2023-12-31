package main

import "time"

type EventData2 struct {
	Type       string     `json:"type,omitempty"`
	Geometry   Geometry   `json:"geometry,omitempty"`
	Properties Properties `json:"properties,omitempty"`
}
type Geometry struct {
	Type        string    `json:"type,omitempty"`
	Coordinates []float64 `json:"coordinates,omitempty"`
}
type Units struct {
	AirPressureAtSeaLevel string `json:"air_pressure_at_sea_level,omitempty"`
	AirTemperature        string `json:"air_temperature,omitempty"`
	CloudAreaFraction     string `json:"cloud_area_fraction,omitempty"`
	PrecipitationAmount   string `json:"precipitation_amount,omitempty"`
	RelativeHumidity      string `json:"relative_humidity,omitempty"`
	WindFromDirection     string `json:"wind_from_direction,omitempty"`
	WindSpeed             string `json:"wind_speed,omitempty"`
}
type Meta struct {
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Units     Units     `json:"units,omitempty"`
}
type Details struct {
	AirPressureAtSeaLevel float64 `json:"air_pressure_at_sea_level,omitempty"`
	AirTemperature        float64 `json:"air_temperature,omitempty"`
	CloudAreaFraction     float64 `json:"cloud_area_fraction,omitempty"`
	RelativeHumidity      float64 `json:"relative_humidity,omitempty"`
	WindFromDirection     float64 `json:"wind_from_direction,omitempty"`
	WindSpeed             float64 `json:"wind_speed,omitempty"`
}
type Instant struct {
	Details Details `json:"details,omitempty"`
}
type Summary struct {
	SymbolCode string `json:"symbol_code,omitempty"`
}
type Next12Hours struct {
	Summary Summary `json:"summary,omitempty"`
}
type Details struct {
	PrecipitationAmount float64 `json:"precipitation_amount,omitempty"`
}
type Next1Hours struct {
	Summary Summary `json:"summary,omitempty"`
	Details Details `json:"details,omitempty"`
}
type Next6Hours struct {
	Summary Summary `json:"summary,omitempty"`
	Details Details `json:"details,omitempty"`
}
type Data struct {
	Instant     Instant     `json:"instant,omitempty"`
	Next12Hours Next12Hours `json:"next_12_hours,omitempty"`
	Next1Hours  Next1Hours  `json:"next_1_hours,omitempty"`
	Next6Hours  Next6Hours  `json:"next_6_hours,omitempty"`
}
type Timeseries struct {
	Time time.Time `json:"time,omitempty"`
	Data Data      `json:"data,omitempty"`
}
type Properties struct {
	Meta       Meta         `json:"meta,omitempty"`
	Timeseries []Timeseries `json:"timeseries,omitempty"`
}
