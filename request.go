package darksky

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	scheme   = "https"
	host     = "api.darksky.net"
	basePath = "forecast"

	languageOptionKey = "lang"
	excludeOptionKey  = "exclude"
	extendOptionKey   = "extend"
	extendOptionValue = "hourly"
	unitOptionKey     = "units"
)

var (
	ErrLanguageNotSupported   = errors.New("Language provided is not supported.")
	ErrUnitNotSupported       = errors.New("Unit provided is not supported.")
	ErrExcludeOptionNotUnique = errors.New("Exclude options must be unique within the same group.")

	// Supported languages
	supportedLanguages = []string{
		"ar",
		"az",
		"be",
		"bg",
		"bs",
		"ca",
		"cs",
		"da",
		"de",
		"el",
		"en",
		"es",
		"et",
		"fi",
		"fr",
		"he",
		"hr",
		"hu",
		"id",
		"is",
		"it",
		"ja",
		"ka",
		"ko",
		"kw",
		"lv",
		"nb",
		"nl",
		"no",
		"pl",
		"pt",
		"ro",
		"ru",
		"sk",
		"sl",
		"sr",
		"sv",
		"te",
		"tr",
		"uk",
		"x-pig-latin",
		"zh",
		"zh-tw",
	}

	// Supported exclude sections
	supportedExclude = []string{
		"currently",
		"minutely",
		"hourly",
		"daily",
		"alerts",
		"flags",
	}

	// Supported units
	supportedUnits = []string{
		"auto",
		"ca",
		"uk2",
		"us",
		"si",
	}
)

type Option func(*url.Values) error

type excludeOptionError struct {
	value string
}

func (oe excludeOptionError) Error() string {
	return fmt.Sprintf("Unsupported value for exclude option : %s", oe.value)
}

func newOptionError(value string) *excludeOptionError {
	return &excludeOptionError{
		value: value,
	}
}

func LanguageOption(lang string) Option {
	return func(v *url.Values) error {
		var supported bool
		var lowerLang = strings.ToLower(lang)

		for _, sl := range supportedLanguages {
			if sl == lowerLang {
				supported = true
			}
		}

		if !supported {
			return ErrLanguageNotSupported
		}

		v.Set(languageOptionKey, lowerLang)

		return nil
	}
}

func ExcludeOption(ex []string) Option {
	return func(v *url.Values) error {
		lowerExcludes := toLower(ex)

		for _, e := range lowerExcludes {
			var supported bool
			var count int

			for _, supex := range supportedExclude {
				if e == supex {
					supported = true
				}
			}

			if !supported {
				return newOptionError(e)
			}

			for _, excl := range lowerExcludes {
				if excl == e {
					count++
				}
			}

			if count > 1 {
				return ErrExcludeOptionNotUnique
			}
		}

		v.Set(excludeOptionKey, "["+strings.Join(lowerExcludes, ",")+"]")

		return nil
	}
}

func ExtendOption() Option {
	return func(v *url.Values) error {
		v.Set(extendOptionKey, extendOptionValue)

		return nil
	}
}

func UnitOption(u string) Option {
	return func(v *url.Values) error {
		var supported bool
		lowerUnit := strings.ToLower(u)

		for _, su := range supportedUnits {
			if su == lowerUnit {
				supported = true
			}
		}

		if !supported {
			return ErrUnitNotSupported
		}

		v.Set(unitOptionKey, lowerUnit)

		return nil
	}
}

func toLower(strs []string) []string {
	lowStrs := make([]string, len(strs))

	for i := 0; i < len(lowStrs); i++ {
		lowStrs[i] = strings.ToLower(strs[i])
	}

	return lowStrs
}

func newForecastRequest(secret string, lat, lng float64, opts []Option) (*http.Request, error) {
	path := fmt.Sprintf("/%s/%s/%3.4f,%3.4f", basePath, secret, lat, lng)

	return newRequest(path, opts)
}

func newTimeMachineRequest(secret string, lat, lng float64, t time.Time, opts []Option) (*http.Request, error) {
	path := fmt.Sprintf("/%s/%s/%3.4f,%3.4f,%d", basePath, secret, lat, lng, int32(t.Unix()))

	return newRequest(path, opts)
}

func newRequest(path string, opts []Option) (*http.Request, error) {
	url := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	q := url.Query()
	var err error

	for _, opt := range opts {
		err = opt(&q)

		if err != nil {
			return nil, err
		}
	}

	url.RawQuery = q.Encode()

	r, err := http.NewRequest(http.MethodGet, url.String(), nil)

	if err != nil {
		return nil, err
	}

	r.Header.Add("Accept-Encoding", "gzip")
	r.Header.Add("Accept", "application/json")

	return r, nil
}
