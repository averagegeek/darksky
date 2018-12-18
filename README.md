# Darksky API Implementation in go

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
- Language through `LanguageOption(lang string)`
- Extend option through `ExtendOption()`
- Exclude option through `ExcludeOption(ex []string)`
- Unit option through `UnitOption(unit string)`

For more information on the API, please visit https://darksky.net/dev/docs
