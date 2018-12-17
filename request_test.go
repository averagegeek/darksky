package darksky

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestForecastRequest(t *testing.T) {
	err := testForecastRequest()

	if err != nil {
		t.Error(err)
	}
}

func TestForecastRequestWithSingleOptions(t *testing.T) {
	opts := []Option{
		LanguageOption("fr"),
		ExcludeOption([]string{"minutely", "hourly"}),
		ExtendOption(),
		UnitOption("ca"),
	}

	for _, opt := range opts {
		err := testForecastRequest(opt)

		if err != nil {
			t.Error(err)
		}
	}
}

func TestForecastRequestWithMultipleOptions(t *testing.T) {
	opts := []Option{
		LanguageOption("fr"),
		ExcludeOption([]string{"minutely", "hourly"}),
		ExtendOption(),
		UnitOption("ca"),
	}

	err := testForecastRequest(opts...)

	if err != nil {
		t.Error(err)
	}
}

func TestTimeMachineRequest(t *testing.T) {
	err := testTimeMachineRequest()

	if err != nil {
		t.Error(err)
	}
}

func TestUnsupportedLanguageOption(t *testing.T) {
	testOptionError(t, ErrLanguageNotSupported, LanguageOption("zzz"))
}

func TestUnSupportedExcludeOption(t *testing.T) {
	ex := []string{"zzz"}
	testOptionError(t, newOptionError("zzz"), ExcludeOption(ex))
}

func TestNotUniqueExcludeOption(t *testing.T) {
	ex := []string{"minutely", "minutely"}
	testOptionError(t, ErrExcludeOptionNotUnique, ExcludeOption(ex))
}

func TestUnsupportedUnit(t *testing.T) {
	testOptionError(t, ErrUnitNotSupported, UnitOption("zzz"))
}

func testForecastRequest(opts ...Option) error {
	r, err := newForecastRequest(defaultToken, defaultLat, defaultLng, opts)

	if err != nil {
		return err
	}

	return validateURL(r, defaultForecastURL, opts)
}

func testTimeMachineRequest(opts ...Option) error {
	t := time.Now()
	ts := int32(t.Unix())
	r, err := newTimeMachineRequest(defaultToken, defaultLat, defaultLng, t, opts)

	if err != nil {
		return err
	}

	return validateURL(r, fmt.Sprintf(defaultTimeMachineURL, ts), opts)
}

func testOptionError(t *testing.T, expectedError error, opts ...Option) {
	_, err := newForecastRequest(defaultToken, defaultLat, defaultLng, opts)

	if err.Error() != expectedError.Error() {
		t.Errorf("Should have error : %s", expectedError)
	}

	_, err = newTimeMachineRequest(defaultToken, defaultLat, defaultLng, time.Now(), opts)

	if err.Error() != expectedError.Error() {
		t.Errorf("Should have error : %s", expectedError)
	}
}

func validateURL(r *http.Request, expected string, opts []Option) error {
	q := make(url.Values)

	for _, opt := range opts {
		opt(&q)
	}

	if queryString := q.Encode(); queryString != "" {
		expected += "?" + q.Encode()
	}

	if r.URL.String() != expected {
		return fmt.Errorf("Request URL should be %s, %s given", expected, r.URL.String())
	}

	return nil
}
