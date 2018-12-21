package darksky

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

var (
	defaultLat    = 37.8267
	defaultLng    = -122.4233
	defaultSecret = "test-secret"

	defaultForecastURL    = "https://api.darksky.net/forecast/test-secret/37.8267,-122.4233"
	defaultTimeMachineURL = "https://api.darksky.net/forecast/test-secret/37.8267,-122.4233,%d"

	ClientMock = &HTTPClientMock{}
)

type HTTPClientMock struct{}

func (c HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	urlSplit := strings.Split(req.URL.String(), "/")
	params := strings.Split(urlSplit[5], ",")

	var body string

	if len(params) == 3 {
		body = timeMachineResponseStub
	} else {
		body = forecastResponseStub
	}

	resp, err := formatResponse(body, 200, req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

type HTTPClientErrorMock struct {
	code        int
	body        string
	contentType string
}

func (unc *HTTPClientErrorMock) Do(req *http.Request) (*http.Response, error) {
	resp, err := formatResponse(unc.body, unc.code, req)

	if err != nil {
		return nil, err
	}

	resp.Header.Add("Content-Type", unc.contentType)

	return resp, nil
}

func newErrorClient(code int, body, contentType string) *HTTPClientErrorMock {
	return &HTTPClientErrorMock{code, body, contentType}
}

func formatResponse(body string, statusCode int, req *http.Request) (*http.Response, error) {
	resp := new(http.Response)
	resp.Header = make(map[string][]string)
	resp.StatusCode = statusCode

	if req.Header.Get("Accept-Encoding") == "gzip" {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)

		_, err := zw.Write([]byte(body))

		if err != nil {
			return nil, err
		}

		defer func() {
			err := zw.Close()

			if err != nil {
				panic(err)
			}
		}()

		resp.Body = nopCloser{&buf, nil}
		resp.Header.Set("Content-Encoding", "gzip")
	} else {
		resp.Body = nopCloser{bytes.NewBufferString(body), nil}
	}

	return resp, nil
}

type nopCloser struct {
	io.Reader
	err error
}

func (np nopCloser) Close() error { return np.err }

type logWriter struct {
	res []byte
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.res = p

	return len(p), nil
}

func TestGetForecast(t *testing.T) {
	api, err := NewAPI("test-secret", HTTPClientOption(ClientMock))

	if err != nil {
		t.Error(err)
	}

	d, err := api.Forecast(defaultLat, defaultLng)

	if err != nil {
		t.Error(err)
	}

	validateForecast(t, d)
}

func TestGetForecastWithInvalidOption(t *testing.T) {
	api, err := NewAPI("test-secret", HTTPClientOption(ClientMock))

	if err != nil {
		t.Error(err)
	}

	_, err = api.Forecast(defaultLat, defaultLng, LanguageOption("test"))

	if err != ErrLanguageNotSupported {
		t.Error("Should have return ErrLanguageNotSupported error")
	}
}

func TestGetTimeMachine(t *testing.T) {
	api, err := NewAPI("test-secret", HTTPClientOption(ClientMock))

	if err != nil {
		t.Error(err)
	}

	d, err := api.TimeMachine(defaultLat, defaultLng, time.Now())

	if err != nil {
		t.Error(err)
	}

	validateTimeMachine(t, d)
}

func TestGetTimeMachineWithInvalidOption(t *testing.T) {
	api, err := NewAPI("test-secret", HTTPClientOption(ClientMock))

	if err != nil {
		t.Error(err)
	}

	_, err = api.TimeMachine(defaultLat, defaultLng, time.Now(), ExcludeOption("test"))

	if err.Error() != newOptionError("test").Error() {
		t.Error("Should have return an excludeOptionError")
	}
}

func TestRequestWithoutGzipEncoding(t *testing.T) {
	api, err := NewAPI("test-secret", HTTPClientOption(ClientMock))

	if err != nil {
		t.Error(err)
	}

	r, err := newForecastRequest(api.secret, defaultLat, defaultLng, []Option{})

	if err != nil {
		t.Error(err)
	}

	r.Header.Del("Accept-Encoding")
	d, err := api.handleRequest(r)

	if err != nil {
		t.Error(err)
	}

	validateForecast(t, d)
}

func TestErrOnEmptySecret(t *testing.T) {
	_, err := NewAPI("")

	if err != ErrEmptySecret {
		t.Error("Empty secret should return ErrEmptySecret")
	}
}

func TestErrNilHTTPClient(t *testing.T) {
	_, err := NewAPI("secret", HTTPClientOption(nil))

	if err != ErrNilHTTPCLient {
		t.Error("Nil http client should return ErrNilHTTPClient")
	}
}

func TestErrNilLogger(t *testing.T) {
	_, err := NewAPI("secret", LoggerOption(nil))

	if err != ErrNilLogger {
		t.Error("Nil logger should return ErrNilLogger")
	}
}

func TestLoggerOption(t *testing.T) {
	w := &logWriter{}
	logger := log.New(w, "darksky test - ", log.LstdFlags)

	api, err := NewAPI("secret", LoggerOption(logger))

	if err != nil {
		t.Error(err)
	}

	api.logger.Print("Hello logger")

	if !strings.Contains(string(w.res), "darksky test - ") && !strings.Contains(string(w.res), "Hello logger") {
		t.Error("Custom logger should have been used")
	}
}

func TestCloseDefer(t *testing.T) {
	err := errors.New("Test error")
	closer := &nopCloser{nil, err}
	writer := &logWriter{}
	logger := log.New(writer, "darksky test - ", log.LstdFlags)

	close(closer, logger)

	if !strings.Contains(string(writer.res), "darksky test - ") && !strings.Contains(string(writer.res), "Test error") {
		t.Error("Closer is returning an error, should have been logged in logger from function parameter.")
	}
}

func TestForecastHttpErrorResponse(t *testing.T) {
	clients := []struct {
		code        int
		message     string
		contentType string
	}{
		{403, "Unauthorized", "text/plain"},
		{400, "Location out of bounds", "application/json"},
		{500, "Server Error", "application/json; charset=utf8"},
	}

	testErr := func(code int, message string, err error) {
		expectedError := HTTPError(code, message)

		if err.Error() != expectedError.Error() {
			t.Errorf("Expected error should read : %s got %s", expectedError, err)
		}
	}

	for _, c := range clients {
		var body string

		if strings.Contains(c.contentType, "json") {
			body = fmt.Sprintf(`{"code":%d,"error":"%s"}`, c.code, c.message)
		} else {
			body = c.message
		}

		errClient := newErrorClient(c.code, body, c.contentType)
		api, err := NewAPI("test-secret", HTTPClientOption(errClient))

		if err != nil {
			t.Error(err)
		}

		_, err = api.Forecast(defaultLat, defaultLng)
		testErr(c.code, c.message, err)

		_, err = api.TimeMachine(defaultLat, defaultLng, time.Now())
		testErr(c.code, c.message, err)
	}
}

func ExampleAPI_Forecast() {
	// Using a mock http client to prevent calling the real api.
	api, err := NewAPI("SECRET", HTTPClientOption(ClientMock))

	if err != nil {
		panic(err)
	}

	data, err := api.Forecast(37.8267, -122.4233,
		LanguageOption(LangEN),
		ExcludeOption(ExMinutely, ExHourly),
		UnitOption(UnitUS),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Current Weather: %s\n", data.Currently.Summary)
	fmt.Printf("Current Temperature: %2.2f\n", data.Currently.Temperature)

	// Output:
	// Current Weather: Overcast
	// Current Temperature: 48.42
}

func ExampleAPI_TimeMachine() {
	// Using a mock http client to prevent calling the real api.
	api, err := NewAPI("SECRET", HTTPClientOption(ClientMock))

	if err != nil {
		panic(err)
	}

	t, err := time.Parse(time.RFC822, "06 Feb 78 19:00 EST")

	if err != nil {
		panic(err)
	}

	data, err := api.TimeMachine(37.8267, -122.4233, t,
		LanguageOption(LangEN),
		ExcludeOption(ExMinutely, ExHourly),
		UnitOption(UnitUS),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Current Weather: %s\n", data.Currently.Summary)
	fmt.Printf("Current Temperature: %2.2f\n", data.Currently.Temperature)
	fmt.Printf("Time: %s", time.Unix(data.Currently.Time, 0).Format(time.RFC822))

	// Output:
	// Current Weather: Mostly Cloudy
	// Current Temperature: 60.46
	// Time: 06 Feb 78 19:00 EST
}

func validateForecast(t *testing.T, d *APIData) {
	assertFloat(t, "Latitude", d.Latitude, 37.8267)
	assertFloat(t, "Longitude", d.Longitude, -122.4233)
	assertString(t, "Timezone", d.Timezone, "America/Los_Angeles")

	assertInt(t, "Currently.Time", d.Currently.Time, 1544378256)
	assertString(t, "Currently.Summary", d.Currently.Summary, "Overcast")
	assertString(t, "Currently.Icon", d.Currently.Icon, "cloudy")
	assertInt(t, "Currently.NearestStormDistance", d.Currently.NearestStormDistance, 12)
	assertInt(t, "Currently.NearestStormBearing", d.Currently.NearestStormBearing, 83)
	assertFloat(t, "Currently.PrecipIntensity", d.Currently.PrecipIntensity, 0)
	assertFloat(t, "Currently.PrecipProbability", d.Currently.PrecipProbability, 0)
	assertFloat(t, "currently.Temperature", d.Currently.Temperature, 48.42)
	assertFloat(t, "currently.ApparentTemperature", d.Currently.ApparentTemperature, 47.57)
	assertFloat(t, "currently.DewPoint", d.Currently.DewPoint, 44.15)
	assertFloat(t, "currently.Humidity", d.Currently.Humidity, 0.85)
	assertFloat(t, "currently.Pressure", d.Currently.Pressure, 1027.1)
	assertFloat(t, "currently.WindSpeed", d.Currently.WindSpeed, 3.35)
	assertFloat(t, "currently.WindGust", d.Currently.WindGust, 8.04)
	assertFloat(t, "Currently.WindBearing", d.Currently.WindBearing, 47)
	assertFloat(t, "currently.CloudCover", d.Currently.CloudCover, 0.97)
	assertInt(t, "Currently.UvIndex", d.Currently.UvIndex, 1)
	assertFloat(t, "currently.Visibility", d.Currently.Visibility, 5.72)
	assertFloat(t, "currently.Ozone", d.Currently.Ozone, 272.39)

	assertString(t, "Minutely.Summary", d.Minutely.Summary, "Overcast for the hour.")
	assertString(t, "Minutely.Icon", d.Minutely.Icon, "cloudy")
	assertInt(t, "Minutely.Data[0].Time", d.Minutely.Data[0].Time, 1544378220)
	assertFloat(t, "Minutely.Data[0].PrecipIntensity", d.Minutely.Data[0].PrecipIntensity, 0)
	assertFloat(t, "Minutely.Data[0].PrecipProbability", d.Minutely.Data[0].PrecipProbability, 0)

	assertString(t, "Hourly.Summary", d.Hourly.Summary, "Mostly cloudy until tomorrow morning.")
	assertString(t, "Hourly.Icon", d.Hourly.Icon, "partly-cloudy-night")
	assertInt(t, "Hourly.Data[0].Time", d.Hourly.Data[0].Time, 1544374800)
	assertString(t, "Hourly.Data[0].Summary", d.Hourly.Data[0].Summary, "Mostly Cloudy")
	assertString(t, "Hourly.Data[0].Icon", d.Hourly.Data[0].Icon, "partly-cloudy-day")
	assertFloat(t, "Hourly.Data[0].PrecipIntensity", d.Hourly.Data[0].PrecipIntensity, 0)
	assertFloat(t, "Hourly.Data[0].PrecipProbability", d.Hourly.Data[0].PrecipProbability, 0)
	assertFloat(t, "Hourly.Data[0].Temperature", d.Hourly.Data[0].Temperature, 47.7)
	assertFloat(t, "Hourly.Data[0].ApparentTemperature", d.Hourly.Data[0].ApparentTemperature, 47.7)
	assertFloat(t, "Hourly.Data[0].DewPoint", d.Hourly.Data[0].DewPoint, 43.55)
	assertFloat(t, "Hourly.Data[0].Humidity", d.Hourly.Data[0].Humidity, 0.85)
	assertFloat(t, "Hourly.Data[0].Pressure", d.Hourly.Data[0].Pressure, 1026.85)
	assertFloat(t, "Hourly.Data[0].WindSpeed", d.Hourly.Data[0].WindSpeed, 2.76)
	assertFloat(t, "Hourly.Data[0].WindGust", d.Hourly.Data[0].WindGust, 8.04)
	assertFloat(t, "Hourly.Data[0].WindBearing", d.Hourly.Data[0].WindBearing, 46)
	assertFloat(t, "Hourly.Data[0].CloudCover", d.Hourly.Data[0].CloudCover, 0.86)
	assertInt(t, "Hourly.Data[0].UvIndex", d.Hourly.Data[0].UvIndex, 1)
	assertFloat(t, "Hourly.Data[0].Visibility", d.Hourly.Data[0].Visibility, 3.76)
	assertFloat(t, "Hourly.Data[0].Ozone", d.Hourly.Data[0].Ozone, 270.82)

	assertString(t, "Daily.Summary", d.Daily.Summary, "Rain tomorrow and next Sunday, with high temperatures peaking at 60°F on Wednesday.")
	assertString(t, "Daily.Icon", d.Daily.Icon, "rain")
	assertInt(t, "Daily.Data[0].Time", d.Daily.Data[0].Time, 1544342400)
	assertString(t, "Daily.Data[0].Summary", d.Daily.Data[0].Summary, "Mostly cloudy throughout the day.")
	assertString(t, "Daily.Data[0].Icon", d.Daily.Data[0].Icon, "partly-cloudy-day")
	assertInt(t, "Daily.Data[0].SunriseTime", d.Daily.Data[0].SunriseTime, 1544368517)
	assertInt(t, "Daily.Data[0].SunsetTime", d.Daily.Data[0].SunsetTime, 1544403121)
	assertFloat(t, "Daily.Data[0].MoonPhase", d.Daily.Data[0].MoonPhase, 0.08)
	assertFloat(t, "Daily.Data[0].PrecipIntensity", d.Daily.Data[0].PrecipIntensity, 0.0002)
	assertFloat(t, "Daily.Data[0].PrecipIntensityMax", d.Daily.Data[0].PrecipIntensityMax, 0.0018)
	assertInt(t, "Daily.Data[0].PrecipIntensityMaxTime", d.Daily.Data[0].PrecipIntensityMaxTime, 1544407200)
	assertFloat(t, "Daily.Data[0].PrecipProbability", d.Daily.Data[0].PrecipProbability, 0.19)
	assertString(t, "Daily.Data[0].PrecipType", d.Daily.Data[0].PrecipType, "rain")
	assertFloat(t, "Daily.Data[0].TemperatureHigh", d.Daily.Data[0].TemperatureHigh, 54.7)
	assertInt(t, "Daily.Data[0].TemperatureHighTime", d.Daily.Data[0].TemperatureHighTime, 1544400000)
	assertFloat(t, "Daily.Data[0].TemperatureLow", d.Daily.Data[0].TemperatureLow, 48.78)
	assertInt(t, "Daily.Data[0].TemperatureLowTime", d.Daily.Data[0].TemperatureLowTime, 1544454000)
	assertFloat(t, "Daily.Data[0].ApparentTemperatureHigh", d.Daily.Data[0].ApparentTemperatureHigh, 54.7)
	assertInt(t, "Daily.Data[0].ApparentTemperatureHighTime", d.Daily.Data[0].ApparentTemperatureHighTime, 1544400000)
	assertFloat(t, "Daily.Data[0].ApparentTemperatureLow", d.Daily.Data[0].ApparentTemperatureLow, 46.57)
	assertInt(t, "Daily.Data[0].ApparentTemperatureLowTime", d.Daily.Data[0].ApparentTemperatureLowTime, 1544454000)
	assertFloat(t, "Daily.Data[0].DewPoint", d.Daily.Data[0].DewPoint, 45.04)
	assertFloat(t, "Daily.Data[0].Humidity", d.Daily.Data[0].Humidity, 0.81)
	assertFloat(t, "Daily.Data[0].Pressure", d.Daily.Data[0].Pressure, 1025.69)
	assertFloat(t, "Daily.Data[0].WindSpeed", d.Daily.Data[0].WindSpeed, 3.02)
	assertFloat(t, "Daily.Data[0].WindGust", d.Daily.Data[0].WindGust, 8.04)
	assertInt(t, "Daily.Data[0].WindGustTime", d.Daily.Data[0].WindGustTime, 1544374800)
	assertFloat(t, "Daily.Data[0].WindBearing", d.Daily.Data[0].WindBearing, 53)
	assertFloat(t, "Daily.Data[0].CloudCover", d.Daily.Data[0].CloudCover, 0.48)
	assertInt(t, "Daily.Data[0].UvIndex", d.Daily.Data[0].UvIndex, 2)
	assertInt(t, "Daily.Data[0].UvIndexTime", d.Daily.Data[0].UvIndexTime, 1544382000)
	assertFloat(t, "Daily.Data[0].Visibility", d.Daily.Data[0].Visibility, 9.62)
	assertFloat(t, "Daily.Data[0].Ozone", d.Daily.Data[0].Ozone, 275.05)

	assertString(t, "Alerts[0].Description", d.Alerts[0].Description, "Test description")
	assertInt(t, "Alerts[0].Expires", d.Alerts[0].Expires, 1544371200)
	assertString(t, "Alerts[0].Regions[0]", d.Alerts[0].Regions[0], "ca")
	assertString(t, "Alerts[0].Regions[1]", d.Alerts[0].Regions[1], "us")
	assertString(t, "Alerts[0].Severity", d.Alerts[0].Severity, "watch")
	assertInt(t, "Alerts[0].Time", d.Alerts[0].Time, 1544371200)
	assertString(t, "Alerts[0].Title", d.Alerts[0].Title, "Alert title")
	assertString(t, "Alerts[0].URI", d.Alerts[0].URI, "https://www.darksky.net")

	assertString(t, "Flags.Sources[0]", d.Flags.Sources[0], "nearest-precip")
	assertString(t, "Flags.Sources[1]", d.Flags.Sources[1], "nwspa")
	assertString(t, "Flags.Sources[2]", d.Flags.Sources[2], "cmc")
	assertString(t, "Flags.Sources[3]", d.Flags.Sources[3], "gfs")
	assertString(t, "Flags.Sources[4]", d.Flags.Sources[4], "hrrr")
	assertString(t, "Flags.Sources[5]", d.Flags.Sources[5], "icon")
	assertString(t, "Flags.Sources[6]", d.Flags.Sources[6], "isd")
	assertString(t, "Flags.Sources[7]", d.Flags.Sources[7], "madis")
	assertString(t, "Flags.Sources[8]", d.Flags.Sources[8], "nam")
	assertString(t, "Flags.Sources[9]", d.Flags.Sources[9], "sref")
	assertString(t, "Flags.Sources[10]", d.Flags.Sources[10], "darksky")
	assertFloat(t, "Flags.NearestStation", d.Flags.NearestStation, 1.839)
	assertString(t, "Flags.Units", d.Flags.Units, "us")
}

func validateTimeMachine(t *testing.T, d *APIData) {
	assertFloat(t, "Latitude", d.Latitude, 37.8267)
	assertFloat(t, "Longitude", d.Longitude, -122.4233)
	assertString(t, "Timezone", d.Timezone, "America/Los_Angeles")

	assertInt(t, "Currently.Time", d.Currently.Time, 255657600)
	assertString(t, "Currently.Summary", d.Currently.Summary, "Mostly Cloudy")
	assertString(t, "Currently.Icon", d.Currently.Icon, "partly-cloudy-day")
	assertFloat(t, "Currently.PrecipIntensity", d.Currently.PrecipIntensity, 0)
	assertFloat(t, "Currently.PrecipProbability", d.Currently.PrecipProbability, 0)
	assertFloat(t, "currently.Temperature", d.Currently.Temperature, 60.46)
	assertFloat(t, "currently.ApparentTemperature", d.Currently.ApparentTemperature, 60.46)
	assertFloat(t, "currently.DewPoint", d.Currently.DewPoint, 53.98)
	assertFloat(t, "currently.Humidity", d.Currently.Humidity, 0.79)
	assertFloat(t, "currently.Pressure", d.Currently.Pressure, 1008.86)
	assertFloat(t, "currently.WindSpeed", d.Currently.WindSpeed, 12.68)
	assertFloat(t, "Currently.WindBearing", d.Currently.WindBearing, 231)
	assertFloat(t, "currently.CloudCover", d.Currently.CloudCover, 0.9)
	assertInt(t, "Currently.UvIndex", d.Currently.UvIndex, 1)
	assertFloat(t, "currently.Visibility", d.Currently.Visibility, 7)

	assertString(t, "Hourly.Summary", d.Hourly.Summary, "Rain overnight and in the morning and breezy in the morning.")
	assertString(t, "Hourly.Icon", d.Hourly.Icon, "rain")
	assertInt(t, "Hourly.Data[0].Time", d.Hourly.Data[0].Time, 255600000)
	assertString(t, "Hourly.Data[0].Summary", d.Hourly.Data[0].Summary, "Overcast")
	assertString(t, "Hourly.Data[0].Icon", d.Hourly.Data[0].Icon, "cloudy")
	assertFloat(t, "Hourly.Data[0].PrecipIntensity", d.Hourly.Data[0].PrecipIntensity, 0)
	assertFloat(t, "Hourly.Data[0].PrecipProbability", d.Hourly.Data[0].PrecipProbability, 0)
	assertFloat(t, "Hourly.Data[0].Temperature", d.Hourly.Data[0].Temperature, 55.34)
	assertFloat(t, "Hourly.Data[0].ApparentTemperature", d.Hourly.Data[0].ApparentTemperature, 55.34)
	assertFloat(t, "Hourly.Data[0].DewPoint", d.Hourly.Data[0].DewPoint, 50.77)
	assertFloat(t, "Hourly.Data[0].Humidity", d.Hourly.Data[0].Humidity, 0.85)
	assertFloat(t, "Hourly.Data[0].Pressure", d.Hourly.Data[0].Pressure, 1011.1)
	assertFloat(t, "Hourly.Data[0].WindSpeed", d.Hourly.Data[0].WindSpeed, 11.19)
	assertFloat(t, "Hourly.Data[0].WindBearing", d.Hourly.Data[0].WindBearing, 149)
	assertFloat(t, "Hourly.Data[0].CloudCover", d.Hourly.Data[0].CloudCover, 0.96)
	assertInt(t, "Hourly.Data[0].UvIndex", d.Hourly.Data[0].UvIndex, 0)
	assertFloat(t, "Hourly.Data[0].Visibility", d.Hourly.Data[0].Visibility, 10)

	assertInt(t, "Daily.Data[0].Time", d.Daily.Data[0].Time, 255600000)
	assertString(t, "Daily.Data[0].Summary", d.Daily.Data[0].Summary, "Rain and breezy in the morning.")
	assertString(t, "Daily.Data[0].Icon", d.Daily.Data[0].Icon, "rain")
	assertInt(t, "Daily.Data[0].SunriseTime", d.Daily.Data[0].SunriseTime, 255625832)
	assertInt(t, "Daily.Data[0].SunsetTime", d.Daily.Data[0].SunsetTime, 255663586)
	assertFloat(t, "Daily.Data[0].MoonPhase", d.Daily.Data[0].MoonPhase, 0.97)
	assertFloat(t, "Daily.Data[0].PrecipIntensity", d.Daily.Data[0].PrecipIntensity, 0.0164)
	assertFloat(t, "Daily.Data[0].PrecipIntensityMax", d.Daily.Data[0].PrecipIntensityMax, 0.1692)
	assertInt(t, "Daily.Data[0].PrecipIntensityMaxTime", d.Daily.Data[0].PrecipIntensityMaxTime, 255625200)
	assertFloat(t, "Daily.Data[0].PrecipProbability", d.Daily.Data[0].PrecipProbability, 1)
	assertString(t, "Daily.Data[0].PrecipType", d.Daily.Data[0].PrecipType, "rain")
	assertFloat(t, "Daily.Data[0].TemperatureHigh", d.Daily.Data[0].TemperatureHigh, 60.75)
	assertInt(t, "Daily.Data[0].TemperatureHighTime", d.Daily.Data[0].TemperatureHighTime, 255650400)
	assertFloat(t, "Daily.Data[0].TemperatureLow", d.Daily.Data[0].TemperatureLow, 54.78)
	assertInt(t, "Daily.Data[0].TemperatureLowTime", d.Daily.Data[0].TemperatureLowTime, 255708000)
	assertFloat(t, "Daily.Data[0].ApparentTemperatureHigh", d.Daily.Data[0].ApparentTemperatureHigh, 60.75)
	assertInt(t, "Daily.Data[0].ApparentTemperatureHighTime", d.Daily.Data[0].ApparentTemperatureHighTime, 255650400)
	assertFloat(t, "Daily.Data[0].ApparentTemperatureLow", d.Daily.Data[0].ApparentTemperatureLow, 54.78)
	assertInt(t, "Daily.Data[0].ApparentTemperatureLowTime", d.Daily.Data[0].ApparentTemperatureLowTime, 255708000)
	assertFloat(t, "Daily.Data[0].DewPoint", d.Daily.Data[0].DewPoint, 52)
	assertFloat(t, "Daily.Data[0].Humidity", d.Daily.Data[0].Humidity, 0.83)
	assertFloat(t, "Daily.Data[0].Pressure", d.Daily.Data[0].Pressure, 1009.02)
	assertFloat(t, "Daily.Data[0].WindSpeed", d.Daily.Data[0].WindSpeed, 9.07)
	assertFloat(t, "Daily.Data[0].WindBearing", d.Daily.Data[0].WindBearing, 187)
	assertFloat(t, "Daily.Data[0].CloudCover", d.Daily.Data[0].CloudCover, 0.83)
	assertInt(t, "Daily.Data[0].UvIndex", d.Daily.Data[0].UvIndex, 3)
	assertInt(t, "Daily.Data[0].UvIndexTime", d.Daily.Data[0].UvIndexTime, 255643200)
	assertFloat(t, "Daily.Data[0].Visibility", d.Daily.Data[0].Visibility, 8.81)
	assertFloat(t, "Daily.Data[0].Visibility", d.Daily.Data[0].Visibility, 8.81)

	assertString(t, "Flags.Sources[2]", d.Flags.Sources[0], "cmc")
	assertString(t, "Flags.Sources[3]", d.Flags.Sources[1], "gfs")
	assertString(t, "Flags.Sources[4]", d.Flags.Sources[2], "hrrr")
	assertString(t, "Flags.Sources[5]", d.Flags.Sources[3], "icon")
	assertString(t, "Flags.Sources[6]", d.Flags.Sources[4], "isd")
	assertString(t, "Flags.Sources[7]", d.Flags.Sources[5], "madis")
	assertString(t, "Flags.Sources[8]", d.Flags.Sources[6], "nam")
	assertString(t, "Flags.Sources[9]", d.Flags.Sources[7], "sref")
	assertFloat(t, "Flags.NearestStation", d.Flags.NearestStation, 2.583)
	assertString(t, "Flags.Units", d.Flags.Units, "us")
}

func assertInt(t *testing.T, name string, value, expected int64) {
	if value != expected {
		t.Errorf("Field %s expected to be %d, got %d", name, expected, value)
	}
}

func assertFloat(t *testing.T, name string, value, expected float64) {
	if value != expected {
		t.Errorf("Field %s expected to be %3.4f, got %3.4f", name, expected, value)
	}
}

func assertString(t *testing.T, name, value, expected string) {
	if value != expected {
		t.Errorf("Field %s expected to be %s, got %s", name, expected, value)
	}
}

var forecastResponseStub = `
{
	"latitude": 37.8267,
    "longitude": -122.4233,
    "timezone": "America/Los_Angeles",
    "currently": {
        "time": 1544378256,
        "summary": "Overcast",
        "icon": "cloudy",
        "nearestStormDistance": 12,
        "nearestStormBearing": 83,
        "precipIntensity": 0,
        "precipProbability": 0,
        "temperature": 48.42,
        "apparentTemperature": 47.57,
        "dewPoint": 44.15,
        "humidity": 0.85,
        "pressure": 1027.1,
        "windSpeed": 3.35,
        "windGust": 8.04,
        "windBearing": 47,
        "cloudCover": 0.97,
        "uvIndex": 1,
        "visibility": 5.72,
        "ozone": 272.39
	},
	"minutely": {
        "summary": "Overcast for the hour.",
        "icon": "cloudy",
        "data": [
            {
                "time": 1544378220,
                "precipIntensity": 0,
                "precipProbability": 0
			}
		]
	},
	"hourly": {
        "summary": "Mostly cloudy until tomorrow morning.",
        "icon": "partly-cloudy-night",
        "data": [
            {
                "time": 1544374800,
                "summary": "Mostly Cloudy",
                "icon": "partly-cloudy-day",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 47.7,
                "apparentTemperature": 47.7,
                "dewPoint": 43.55,
                "humidity": 0.85,
                "pressure": 1026.85,
                "windSpeed": 2.76,
                "windGust": 8.04,
                "windBearing": 46,
                "cloudCover": 0.86,
                "uvIndex": 1,
                "visibility": 3.76,
                "ozone": 270.82
			}
		]
	},
	"daily": {
        "summary": "Rain tomorrow and next Sunday, with high temperatures peaking at 60°F on Wednesday.",
        "icon": "rain",
        "data": [
            {
                "time": 1544342400,
                "summary": "Mostly cloudy throughout the day.",
                "icon": "partly-cloudy-day",
                "sunriseTime": 1544368517,
                "sunsetTime": 1544403121,
                "moonPhase": 0.08,
                "precipIntensity": 0.0002,
                "precipIntensityMax": 0.0018,
                "precipIntensityMaxTime": 1544407200,
                "precipProbability": 0.19,
                "precipType": "rain",
                "temperatureHigh": 54.7,
                "temperatureHighTime": 1544400000,
                "temperatureLow": 48.78,
                "temperatureLowTime": 1544454000,
                "apparentTemperatureHigh": 54.7,
                "apparentTemperatureHighTime": 1544400000,
                "apparentTemperatureLow": 46.57,
                "apparentTemperatureLowTime": 1544454000,
                "dewPoint": 45.04,
                "humidity": 0.81,
                "pressure": 1025.69,
                "windSpeed": 3.02,
                "windGust": 8.04,
                "windGustTime": 1544374800,
                "windBearing": 53,
                "cloudCover": 0.48,
                "uvIndex": 2,
                "uvIndexTime": 1544382000,
                "visibility": 9.62,
                "ozone": 275.05,
                "temperatureMin": 47.17,
                "temperatureMinTime": 1544371200,
                "temperatureMax": 54.7,
                "temperatureMaxTime": 1544400000,
                "apparentTemperatureMin": 47.17,
                "apparentTemperatureMinTime": 1544371200,
                "apparentTemperatureMax": 54.7,
                "apparentTemperatureMaxTime": 1544400000
			}
		]
    },
    "alerts": [
        {
            "description": "Test description",
            "expires": 1544371200,
            "regions": [
                "ca",
                "us"
            ],
            "severity": "watch",
            "time": 1544371200,
            "title": "Alert title",
            "uri": "https://www.darksky.net"
        }
    ],
	"flags": {
        "sources": [
            "nearest-precip",
            "nwspa",
            "cmc",
            "gfs",
            "hrrr",
            "icon",
            "isd",
            "madis",
            "nam",
            "sref",
            "darksky"
        ],
        "nearest-station": 1.839,
        "units": "us"
    }
}
`

var timeMachineResponseStub = `
{
    "latitude": 37.8267,
    "longitude": -122.4233,
    "timezone": "America/Los_Angeles",
    "currently": {
        "time": 255657600,
        "summary": "Mostly Cloudy",
        "icon": "partly-cloudy-day",
        "precipIntensity": 0,
        "precipProbability": 0,
        "temperature": 60.46,
        "apparentTemperature": 60.46,
        "dewPoint": 53.98,
        "humidity": 0.79,
        "pressure": 1008.86,
        "windSpeed": 12.68,
        "windBearing": 231,
        "cloudCover": 0.9,
        "uvIndex": 1,
        "visibility": 7
    },
    "hourly": {
        "summary": "Rain overnight and in the morning and breezy in the morning.",
        "icon": "rain",
        "data": [
            {
                "time": 255600000,
                "summary": "Overcast",
                "icon": "cloudy",
                "precipIntensity": 0,
                "precipProbability": 0,
                "temperature": 55.34,
                "apparentTemperature": 55.34,
                "dewPoint": 50.77,
                "humidity": 0.85,
                "pressure": 1011.1,
                "windSpeed": 11.19,
                "windBearing": 149,
                "cloudCover": 0.96,
                "uvIndex": 0,
                "visibility": 10
			}
		]
	},
	"daily": {
        "data": [
            {
                "time": 255600000,
                "summary": "Rain and breezy in the morning.",
                "icon": "rain",
                "sunriseTime": 255625832,
                "sunsetTime": 255663586,
                "moonPhase": 0.97,
                "precipIntensity": 0.0164,
                "precipIntensityMax": 0.1692,
                "precipIntensityMaxTime": 255625200,
                "precipProbability": 1,
                "precipType": "rain",
                "temperatureHigh": 60.75,
                "temperatureHighTime": 255650400,
                "temperatureLow": 54.78,
                "temperatureLowTime": 255708000,
                "apparentTemperatureHigh": 60.75,
                "apparentTemperatureHighTime": 255650400,
                "apparentTemperatureLow": 54.78,
                "apparentTemperatureLowTime": 255708000,
                "dewPoint": 52,
                "humidity": 0.83,
                "pressure": 1009.02,
                "windSpeed": 9.07,
                "windBearing": 187,
                "cloudCover": 0.83,
                "uvIndex": 3,
                "uvIndexTime": 255643200,
                "visibility": 8.81,
                "temperatureMin": 54.58,
                "temperatureMinTime": 255621600,
                "temperatureMax": 60.75,
                "temperatureMaxTime": 255650400,
                "apparentTemperatureMin": 54.58,
                "apparentTemperatureMinTime": 255621600,
                "apparentTemperatureMax": 60.75,
                "apparentTemperatureMaxTime": 255650400
            }
        ]
    },
    "flags": {
        "sources": [
            "cmc",
            "gfs",
            "hrrr",
            "icon",
            "isd",
            "madis",
            "nam",
            "sref"
        ],
        "nearest-station": 2.583,
        "units": "us"
	}
}
`
