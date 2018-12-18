# Darksky API Implementation in go

[![Go Report](https://goreportcard.com/badge/github.com/averagegeek/darksky)](https://goreportcard.com/report/github.com/averagegeek/darksky) [![GoDoc](https://camo.githubusercontent.com/c4a9ad8a86803572eb10c2d541e44c79563d2c0e/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f636f636b726f61636864622f636f636b726f6163683f7374617475732e737667)](https://godoc.org/github.com/averagegeek/darksky)


A simple implementation of a [darksky](https://darksky.net) API client.

## Usage

First instantiate an api with your secret (you can sign up for free and get a secret [here](https://darksky.net))
```
    api, err := darksky.NewAPI("my-secret")
```
If you need to pass a custom HTTP client, you can use an option like this:
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

## Contributing

After forking the repo and checking it out, run the following commands to setup the pre-commit scripts:

```
brew update
brew tap alecthomas/homebrew-tap
brew install pre-commit
brew install golangci/tap/golangci-lint
brew install gometalinter

go get github.com/BurntSushi/toml/cmd/tomlv
go get -u github.com/go-critic/go-critic/...
```

To run the pre-commit without the need of commiting a file simply run:
```
pre-commit run --all-files
```
