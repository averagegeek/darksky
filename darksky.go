package darksky

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// Exclude options
	ExCurrently = "currently"
	ExMinutely  = "minutely"
	ExHourly    = "hourly"
	ExDaily     = "daily"
	ExAlerts    = "alerts"
	ExFlags     = "flags"

	// Languages
	LangAR   = "ar"
	LangAZ   = "az"
	LangBE   = "be"
	LangBG   = "bg"
	LangBS   = "bs"
	LangCA   = "ca"
	LangCS   = "cs"
	LangDA   = "da"
	LangDE   = "de"
	LangEL   = "el"
	LangEN   = "en"
	LangES   = "es"
	LangET   = "et"
	LangFI   = "fi"
	LangFR   = "fr"
	LangHE   = "he"
	LangHR   = "hr"
	LangHU   = "hu"
	LangID   = "id"
	LangIS   = "is"
	LangIT   = "it"
	LangJA   = "ja"
	LangKA   = "ka"
	LangKO   = "ko"
	LangKW   = "kw"
	LangLV   = "lv"
	LangNB   = "nb"
	LangNL   = "nl"
	LangNO   = "no"
	LangPL   = "pl"
	LangPT   = "pt"
	LangRO   = "ro"
	LangRU   = "ru"
	LangSK   = "sk"
	LangSL   = "sl"
	LangSR   = "sr"
	LangSV   = "sv"
	LangTE   = "te"
	LangTR   = "tr"
	LangUK   = "uk"
	LangXPIG = "x-pig-latin"
	LangZH   = "zh"
	LangZHTW = "zh-tw"

	// Units
	UnitAuto = "auto"
	UnitCA   = "ca"
	UnitUK2  = "uk2"
	UnitUS   = "us"
	UnitSI   = "si"
)

// APIData represents the whole payload for both request type.
type APIData struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
	Currently DataPoint `json:"currently,omitempty"`
	Minutely  DataBlock `json:"minutely,omitempty"`
	Hourly    DataBlock `json:"hourly,omitempty"`
	Daily     DataBlock `json:"daily,omitempty"`
	Alerts    []Alert   `json:"alerts,omitempty"`
	Flags     Flags     `json:"flags,omitempty"`
}

type Alert struct {
	Description string   `json:"description"`
	Expires     int64    `json:"expires"`
	Regions     []string `json:"regions"`
	Severity    string   `json:"severity"`
	Time        int64    `json:"time"`
	Title       string   `json:"title"`
	URI         string   `json:"uri"`
}

type DataBlock struct {
	Data    []DataPoint `json:"data"`
	Summary string      `json:"summary,omitempty"`
	Icon    string      `json:"icon,omitempty"`
}

type DataPoint struct {
	ApparentTemperature         float64 `json:"apparentTemperature,omitempty"`
	ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh,omitempty"`
	ApparentTemperatureHighTime int64   `json:"apparentTemperatureHighTime,omitempty"`
	ApparentTemperatureLow      float64 `json:"apparentTemperatureLow,omitempty"`
	ApparentTemperatureLowTime  int64   `json:"apparentTemperatureLowTime,omitempty"`
	CloudCover                  float64 `json:"cloudCover,omitempty"`
	DewPoint                    float64 `json:"dewPoint,omitempty"`
	Humidity                    float64 `json:"humidity,omitempty"`
	Icon                        string  `json:"icon,omitempty"`
	MoonPhase                   float64 `json:"moonPhase,omitempty"`
	NearestStormBearing         int64   `json:"nearestStormBearing,omitempty"`
	NearestStormDistance        int64   `json:"nearestStormDistance,omitempty"`
	Ozone                       float64 `json:"ozone,omitempty"`
	PrecipAccumulation          float64 `json:"precipAccumulation,omitempty"`
	PrecipIntensity             float64 `json:"precipIntensity,omitempty"`
	PrecipIntensityError        float64 `json:"precipIntensityError,omitempty"`
	PrecipIntensityMax          float64 `json:"precipIntensityMax,omitempty"`
	PrecipIntensityMaxTime      int64   `json:"precipIntensityMaxTime,omitempty"`
	PrecipProbability           float64 `json:"precipProbability,omitempty"`
	PrecipType                  string  `json:"precipType,omitempty"`
	Pressure                    float64 `json:"pressure,omitempty"`
	Summary                     string  `json:"summary,omitempty"`
	SunriseTime                 int64   `json:"sunriseTime,omitempty"`
	SunsetTime                  int64   `json:"sunsetTime,omitempty"`
	Temperature                 float64 `json:"temperature,omitempty"`
	TemperatureHigh             float64 `json:"temperatureHigh,omitempty"`
	TemperatureHighTime         int64   `json:"temperatureHighTime,omitempty"`
	TemperatureLow              float64 `json:"temperatureLow,omitempty"`
	TemperatureLowTime          int64   `json:"temperatureLowTime,omitempty"`
	Time                        int64   `json:"time"`
	UvIndex                     int64   `json:"uvIndex,omitempty"`
	UvIndexTime                 int64   `json:"uvIndexTime,omitempty"`
	Visibility                  float64 `json:"visibility,omitempty"`
	WindBearing                 float64 `json:"windBearing,omitempty"`
	WindGust                    float64 `json:"windGust,omitempty"`
	WindGustTime                int64   `json:"windGustTime,omitempty"`
	WindSpeed                   float64 `json:"windSpeed,omitempty"`
}

type Flags struct {
	DarkskyUnavailable string   `json:"darksky-unavailable,omitempty"`
	Sources            []string `json:"sources"`
	NearestStation     float64  `json:"nearest-station"`
	Units              string   `json:"units"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type API struct {
	secret string
	client HTTPClient
}

type APIOption func(*API) error

var (
	ErrEmptySecret   = errors.New("Secret cannot be empty.")
	ErrNilHTTPCLient = errors.New("HTTP client provided cannot be nil")
)

func HTTPClientOption(c HTTPClient) APIOption {
	return func(api *API) error {
		if c == nil {
			return ErrNilHTTPCLient
		}

		api.client = c

		return nil
	}
}

func NewAPI(secret string, opts ...APIOption) (*API, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	api := &API{
		secret,
		http.DefaultClient,
	}

	for _, opt := range opts {
		err := opt(api)

		if err != nil {
			return nil, err
		}
	}

	return api, nil
}

func (api API) Forecast(lat, lng float64, opts ...Option) (wd *APIData, err error) {
	r, err := newForecastRequest(api.secret, lat, lng, opts)

	if err != nil {
		return nil, err
	}

	return api.handleRequest(r)
}

func (api API) TimeMachine(lat, lng float64, time time.Time, opts ...Option) (*APIData, error) {
	r, err := newTimeMachineRequest(api.secret, lat, lng, time, opts)

	if err != nil {
		return nil, err
	}

	return api.handleRequest(r)
}

func (api *API) handleRequest(r *http.Request) (wd *APIData, err error) {
	resp, err := api.client.Do(r)

	if err != nil {
		return
	}

	content, err := extractContent(resp)

	if err != nil {
		return
	}

	err = json.Unmarshal(content, &wd)

	return
}

func extractContent(resp *http.Response) ([]byte, error) {
	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		return uncompressGzip(content)
	default:
		return content, err
	}
}

func uncompressGzip(body []byte) ([]byte, error) {
	buf := bytes.NewBuffer(body)
	gr, err := gzip.NewReader(buf)
	defer gr.Close()

	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(gr)

	if err != nil {
		return nil, err
	}

	return b, err
}
