package darksky

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Exclude some data blocks from the API response.

	// ExCurrently option to exclude the Currently DataBlock from the response.
	ExCurrently = "currently"
	// ExMinutely option to exclude the Minutely DataBlock from the response.
	ExMinutely = "minutely"
	// ExHourly option to exclude the Hourly DataBlock from the response.
	ExHourly = "hourly"
	// ExDaily option to exclude the Daily DataBlock from the response.
	ExDaily = "daily"
	// ExAlerts option to exclude Alerts from the response.
	ExAlerts = "alerts"
	// ExFlags option to exclude Flags from the response.
	ExFlags = "flags"

	// Return summary properties in the desired language. (Note that units in the summary will be
	// set according to the units parameter, so be sure to set both parameters appropriately.)
	// language may be:

	// LangAR => Arabic
	LangAR = "ar"
	// LangAZ => Azerbaijani
	LangAZ = "az"
	// LangBE => Belarusian
	LangBE = "be"
	// LangBG => Bulgarian
	LangBG = "bg"
	// LangBS => Bosnian
	LangBS = "bs"
	// LangCA => Catalan
	LangCA = "ca"
	// LangCS => Czech
	LangCS = "cs"
	// LangDA => Danish
	LangDA = "da"
	// LangDE => German
	LangDE = "de"
	// LangEL => Greek
	LangEL = "el"
	// LangEN => English (which is the default)
	LangEN = "en"
	// LangES : Spanish
	LangES = "es"
	// LangET => Estonian
	LangET = "et"
	// LangFI => Finnish
	LangFI = "fi"
	// LangFR => French
	LangFR = "fr"
	// LangHE => Hebrew
	LangHE = "he"
	// LangHR => Croatian
	LangHR = "hr"
	// LangHU => Hungarian
	LangHU = "hu"
	// LangID => Indonesian
	LangID = "id"
	// LangIS => Icelandic
	LangIS = "is"
	// LangIT => Italian
	LangIT = "it"
	// LangJA => Japanese
	LangJA = "ja"
	// LangKA => Georgian
	LangKA = "ka"
	// LangKO => Korean
	LangKO = "ko"
	// LangKW => Cornish
	LangKW = "kw"
	// LangLV => Latvian
	LangLV = "lv"
	// LangNB => Norwegian Bokmål
	LangNB = "nb"
	// LangNL => Dutch
	LangNL = "nl"
	// LangNO => Norwegian Bokmål (alias for nb)
	LangNO = "no"
	// LangPL => Polish
	LangPL = "pl"
	// LangPT => Portuguese
	LangPT = "pt"
	// LangRO => Romanian
	LangRO = "ro"
	// LangRU => Russian
	LangRU = "ru"
	// LangSK => Slovak
	LangSK = "sk"
	// LangSL => Slovenian
	LangSL = "sl"
	// LangSR => Serbian
	LangSR = "sr"
	// LangSV => Swedish
	LangSV = "sv"
	// LangTE =>  Tetum
	LangTE = "te"
	// LangTR => Turkish
	LangTR = "tr"
	// LangUK => Ukrainian
	LangUK = "uk"
	// LangXPIG => Igpay Atinlay
	LangXPIG = "x-pig-latin"
	// LangZH => simplified Chinese
	LangZH = "zh"
	// LangZHTW => traditional Chinese
	LangZHTW = "zh-tw"

	// Return weather conditions in the requested units. Should be one of the following:

	// UnitAuto automatically select units based on geographic location.
	UnitAuto = "auto"
	// UnitCA same as si, except that windSpeed and windGust are in kilometers per hour.
	UnitCA = "ca"
	// UnitUK2 same as si, except that nearestStormDistance and visibility are in miles, and windSpeed and windGust in miles per hour.
	UnitUK2 = "uk2"
	// UnitUS Imperial units (the default).
	UnitUS = "us"
	// UnitSI SI units.
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

// Alert If present, contains any severe weather alerts pertinent to the requested location.
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

// DataPoint object contains various properties, each representing the average
// (unless otherwise specified) of a particular weather phenomenon occurring during
// a period of time: an instant in the case of currently, a minute for minutely,
// an hour for hourly, and a day for daily.
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

// Flags object contains miscellaneous metadata about the request.
type Flags struct {
	DarkskyUnavailable string   `json:"darksky-unavailable,omitempty"`
	Sources            []string `json:"sources"`
	NearestStation     float64  `json:"nearest-station"`
	Units              string   `json:"units"`
}

type apiError struct {
	Code int    `json:"code"`
	Err  string `json:"error"`
}

// HTTPClient let's you substitute the default http.Client for a custom one.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// API is used to make requests.
type API struct {
	secret string
	client HTTPClient
	logger *log.Logger
}

// APIOption to override defaults of the api, like the HTTP client.
type APIOption func(*API) error

var (
	// ErrEmptySecret occurs when passing an empty token on api creation.
	ErrEmptySecret = errors.New("secret cannot be empty")

	// ErrNilHTTPCLient occurs when passing a nil client to the HTTPClientOption.
	ErrNilHTTPCLient = errors.New("HTTP client provided cannot be nil")

	// ErrNilLogger occurs when passing a nil logger to the LoggerOption.
	ErrNilLogger = errors.New("logger provided cannot be null")
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

// LoggerOption to add custom logging on some error less relevant for the api to return, but still want to know about.
func LoggerOption(l *log.Logger) APIOption {
	return func(api *API) error {
		if l == nil {
			return ErrNilLogger
		}

		api.logger = l

		return nil
	}
}

// NewAPI is a helper function to create a new API.
func NewAPI(secret string, opts ...APIOption) (*API, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	api := &API{secret: secret}

	for _, opt := range opts {
		if err := opt(api); err != nil {
			return nil, err
		}
	}

	if api.client == nil {
		api.client = http.DefaultClient
	}

	if api.logger == nil {
		api.logger = log.New(os.Stderr, "Darksky API Client - ", log.LstdFlags)
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

func (api *API) handleRequest(r *http.Request) (*APIData, error) {
	resp, err := api.client.Do(r)

	if err != nil {
		return nil, err
	}

	content, err := extractContent(resp, api.logger)

	if err != nil {
		return nil, err
	}

	data, err := unmarshalContent(resp, content)

	if err != nil {
		return nil, err
	}

	return data, err
}

func extractContent(resp *http.Response, logger *log.Logger) ([]byte, error) {
	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	defer close(resp.Body, logger)

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		return uncompressGzip(content, logger)
	default:
		return content, err
	}
}

func unmarshalContent(resp *http.Response, content []byte) (*APIData, error) {
	if resp.StatusCode >= 400 {
		contentType := resp.Header.Get("Content-Type")

		if contentType == "text/plain" {
			return nil, HTTPError(resp.StatusCode, string(content))
		} else if strings.Contains(contentType, "application/json") {
			var data apiError

			if err := json.Unmarshal(content, &data); err != nil {
				return nil, err
			}

			return nil, HTTPError(resp.StatusCode, data.Err)
		}
	}

	var data *APIData

	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// HTTPError formats a txt error to inform it's an HTTP error and also include code.
func HTTPError(code int, txt string) error {
	return fmt.Errorf("HTTP %d Error - %s", code, txt)
}

func uncompressGzip(body []byte, logger *log.Logger) ([]byte, error) {
	buf := bytes.NewBuffer(body)
	gr, err := gzip.NewReader(buf)

	if err != nil {
		return nil, err
	}

	defer close(gr, logger)

	b, err := ioutil.ReadAll(gr)

	if err != nil {
		return nil, err
	}

	return b, err
}

func close(c io.Closer, l *log.Logger) {
	err := c.Close()

	if err != nil {
		l.Println(err)
	}
}
