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

	// ExCurrently currently
	ExCurrently = "currently"
	// ExMinutely minutely
	ExMinutely = "minutely"
	// ExHourly hourly
	ExHourly = "hourly"
	// ExDaily daily
	ExDaily = "daily"
	// ExAlerts alerts
	ExAlerts = "alerts"
	// ExFlags flags
	ExFlags = "flags"

	// Supported Languages

	// LangAR ...
	LangAR = "ar"
	// LangAZ ...
	LangAZ = "az"
	// LangBE ...
	LangBE = "be"
	// LangBG ...
	LangBG = "bg"
	// LangBS ...
	LangBS = "bs"
	// LangCA ...
	LangCA = "ca"
	// LangCS ...
	LangCS = "cs"
	// LangDA ...
	LangDA = "da"
	// LangDE ...
	LangDE = "de"
	// LangEL ...
	LangEL = "el"
	// LangEN ...
	LangEN = "en"
	// LangES ...
	LangES = "es"
	// LangET ...
	LangET = "et"
	// LangFI ...
	LangFI = "fi"
	// LangFR ...
	LangFR = "fr"
	// LangHE ...
	LangHE = "he"
	// LangHR ...
	LangHR = "hr"
	// LangHU ...
	LangHU = "hu"
	// LangID ...
	LangID = "id"
	// LangIS ...
	LangIS = "is"
	// LangIT ...
	LangIT = "it"
	// LangJA ...
	LangJA = "ja"
	// LangKA ...
	LangKA = "ka"
	// LangKO ...
	LangKO = "ko"
	// LangKW ...
	LangKW = "kw"
	// LangLV ...
	LangLV = "lv"
	// LangNB ...
	LangNB = "nb"
	// LangNL ...
	LangNL = "nl"
	// LangNO ...
	LangNO = "no"
	// LangPL ...
	LangPL = "pl"
	// LangPT ...
	LangPT = "pt"
	// LangRO ...
	LangRO = "ro"
	// LangRU ...
	LangRU = "ru"
	// LangSK ...
	LangSK = "sk"
	// LangSL ...
	LangSL = "sl"
	// LangSR ...
	LangSR = "sr"
	// LangSV ...
	LangSV = "sv"
	// LangTE ...
	LangTE = "te"
	// LangTR ...
	LangTR = "tr"
	// LangUK ...
	LangUK = "uk"
	// LangXPIG ...
	LangXPIG = "x-pig-latin"
	// LangZH ...
	LangZH = "zh"
	// LangZHTW ...
	LangZHTW = "zh-tw"

	// Units

	// UnitAuto automatic units
	UnitAuto = "auto"
	// UnitCA Canada units
	UnitCA = "ca"
	// UnitUK2 United Kingdom units
	UnitUK2 = "uk2"
	// UnitUS USA units
	UnitUS = "us"
	// UnitSI International System of Units units
	UnitSI = "si"
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

// Alert builds out the alert message object
type Alert struct {
	Description string   `json:"description"`
	Expires     int64    `json:"expires"`
	Regions     []string `json:"regions"`
	Severity    string   `json:"severity"`
	Time        int64    `json:"time"`
	Title       string   `json:"title"`
	URI         string   `json:"uri"`
}

// DataBlock object representation from the API.
type DataBlock struct {
	Data    []DataPoint `json:"data"`
	Summary string      `json:"summary,omitempty"`
	Icon    string      `json:"icon,omitempty"`
}

// DataPoint object from the API.
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

// Flags object from the API
type Flags struct {
	DarkskyUnavailable string   `json:"darksky-unavailable,omitempty"`
	Sources            []string `json:"sources"`
	NearestStation     float64  `json:"nearest-station"`
	Units              string   `json:"units"`
}

// HTTPClient the HTTP client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// API is the API client and secret
type API struct {
	secret string
	client HTTPClient
}

// APIOption is an error
type APIOption func(*API) error

var (
	// ErrEmptySecret error message
	ErrEmptySecret = errors.New("secret cannot be empty")
	// ErrNilHTTPCLient error message
	ErrNilHTTPCLient = errors.New("HTTP client provided cannot be nil")
)

// HTTPClientOption is for when you need a custom client instead of the http.DefaultCLient
func HTTPClientOption(c HTTPClient) APIOption {
	return func(api *API) error {
		if c == nil {
			return ErrNilHTTPCLient
		}

		api.client = c

		return nil
	}
}

// NewAPI is a helper function to create a new API.
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

// Forecast query to the API.
func (api API) Forecast(lat, lng float64, opts ...Option) (wd *APIData, err error) {
	r, err := newForecastRequest(api.secret, lat, lng, opts)

	if err != nil {
		return nil, err
	}

	return api.handleRequest(r)
}

// TimeMachine query to the API.
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

	return wd, err
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

	if err != nil {
		return nil, err
	}
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
