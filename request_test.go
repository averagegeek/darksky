package darksky

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func TestTimeMachineRequestWithSingleOption(t *testing.T) {
	opts := []Option{
		LanguageOption("fr"),
		ExcludeOption([]string{"minutely", "hourly"}),
		ExtendOption(),
		UnitOption("ca"),
	}

	for _, opt := range opts {
		err := testTimeMachineRequest(opt)

		if err != nil {
			t.Error(err)
		}
	}
}

func TestTimeMachineRequestWithMultipleOptions(t *testing.T) {
	opts := []Option{
		LanguageOption("fr"),
		ExcludeOption([]string{"minutely", "hourly"}),
		ExtendOption(),
		UnitOption("ca"),
	}

	err := testTimeMachineRequest(opts...)

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
	for _, se := range supportedExclude {
		values := []string{se, strings.ToUpper(se)}
		testOptionError(t, ErrExcludeOptionNotUnique, ExcludeOption(values))
	}
}

func TestUnsupportedUnit(t *testing.T) {
	testOptionError(t, ErrUnitNotSupported, UnitOption("zzz"))
}

func TestLanguageUpperCaseOption(t *testing.T) {
	for _, sl := range supportedLanguages {
		values := make(url.Values)
		option := LanguageOption(strings.ToUpper(sl))
		err := option(&values)

		if err != nil {
			t.Error("Should not return an error, only case do not match.")
		}

		if values.Get("lang") != sl {
			t.Error("Language option should have been converted to lower case.")
		}
	}
}

func TestUnitCaseOption(t *testing.T) {
	for _, su := range supportedUnits {
		values := make(url.Values)
		option := UnitOption(strings.ToUpper(su))
		err := option(&values)

		if err != nil {
			t.Error("Should not return an error, only case do not match.")
		}

		if values.Get("units") != su {
			t.Error("Unit option should have been converted to lower case.")
		}
	}
}

func testForecastRequest(opts ...Option) error {
	r, err := newForecastRequest(defaultSecret, defaultLat, defaultLng, opts)

	if err != nil {
		return err
	}

	return validateURL(r, defaultForecastURL, opts)
}

func testTimeMachineRequest(opts ...Option) error {
	t := time.Now()
	ts := int32(t.Unix())
	r, err := newTimeMachineRequest(defaultSecret, defaultLat, defaultLng, t, opts)

	if err != nil {
		return err
	}

	return validateURL(r, fmt.Sprintf(defaultTimeMachineURL, ts), opts)
}

func testOptionError(t *testing.T, expectedError error, opts ...Option) {
	_, err := newForecastRequest(defaultSecret, defaultLat, defaultLng, opts)

	if err == nil || err.Error() != expectedError.Error() {
		t.Errorf("Should have error : %s", expectedError)
	}

	_, err = newTimeMachineRequest(defaultSecret, defaultLat, defaultLng, time.Now(), opts)

	if err == nil || err.Error() != expectedError.Error() {
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
