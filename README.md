# New weather.com / Weather Underground API Client for Go

[![GoDoc](https://godoc.org/github.com/go-weather/weather?status.svg)](https://godoc.org/github.com/go-weather/weather)

weather implements a client for weather.com API
at https://api.weather.com/v1/. Sometimes this API is called
the "Weather Underground API".

## Features

The following API endpoints are supported:

- Current conditions by coordinates
- "Imminent" forecast ("Rain starting in 45 minutes")
- 10 day forecast by coordinates

To retrive weather for a location like a city, it must be geocoded first.
I recommend the [geocoder](https://github.com/jasonwinn/geocoder) package.

## Requirements

An API key is required to use this package.
It can be obtained from HTML source of various forecast pages on
weather.com and wunderground.com, for example
https://www.wunderground.com/weather/us/ny/new-york.

## Documentation

Documentation is available on [godoc.org](https://godoc.org/github.com/go-weather/weather).

## License

Released under the MIT license.
