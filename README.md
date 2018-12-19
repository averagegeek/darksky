# Darksky API Client Implementation in go

A Simple implementation of a [darksky](https://darksky.net) API client.

## Usage

First instantiate an api with your secret (you can sign up for free and get a secret [here](https://darksky.net))
```
    api, err := darksky.NewAPI("my-secret")
```
If you need to pass a custom http client, you can use an option like this:
```
    api, err := darksky.NewAPI(
        "my-secret",
        darksky.HTTPClientOption(&MyCustomClient{}),
    )
```
It's also possible to pass a logger, for errors that are less relevant to the API user, but still relevant to report. The default logger will be logging to standard error.
```
    logger := log.New(os.Stderr, "Darksky API Client - ", log.LstdFlags)

    api, err := darksky.NewAPI(
        "my-secret",
        darksky.LoggerOption(logger),
    )
```

Then, you can query the API for forecast or time machine request like this:

```
    data, err := api.Forecast(42.3601, -71.0589)
    data, err := api.TimeMachine(42.3601, -71.0589, time.Now())
```

You can pass options to the query like this:

```
    data, err := api.Forecast(
        42.3601, -71.0589,
        darksky.LanguageOption(darksky.LangFR),
        ...
        )
```

Those options are :
- Language through `ex. LanguageOption(LangFR)`
- Extend option through `ex. ExtendOption()`
- Exclude option through `ex. ExcludeOption(ExMinutely, ExHourly)`
- Unit option through `ex. UnitOption(UnitCA)`

For more information on the API, please visit https://darksky.net/dev/docs
