package weather

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

const(
api_key = "6532d6454b8aa370768e63d6ba5a832e"
test_lat = 40.754864
test_lng = -74.007156
)

func TestCurrentImperial(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetCurrentByLocation(test_lat, test_lng, "e")
  assert.Nil(t, err)
  assert.NotNil(t, resp.Observation.Imperial)
}

func TestCurrentMetric(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetCurrentByLocation(test_lat, test_lng, "m")
  assert.Nil(t, err)
  assert.NotNil(t, resp.Observation.Metric)
}

func TestCurrentMetricSi(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetCurrentByLocation(test_lat, test_lng, "s")
  assert.Nil(t, err)
  assert.NotNil(t, resp.Observation.MetricSi)
}

func TestCurrentUkHybrid(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetCurrentByLocation(test_lat, test_lng, "h")
  assert.Nil(t, err)
  assert.NotNil(t, resp.Observation.UkHybrid)
}

func TestCurrentAll(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetCurrentByLocation(test_lat, test_lng, "a")
  assert.Nil(t, err)
  assert.NotNil(t, resp.Observation.Imperial)
  assert.NotNil(t, resp.Observation.Metric)
  assert.NotNil(t, resp.Observation.MetricSi)
  assert.NotNil(t, resp.Observation.UkHybrid)
}

func TestWwirImperial(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetWwirByLocation(test_lat, test_lng, "e")
  assert.Nil(t, err)
  assert.Equal(t, "fod_short_range_wwir", resp.Forecast.Class)
}

func TestForecast10Imperial(t *testing.T) {
  c := NewClient(api_key)
  resp, err := c.GetForecast10ByLocation(test_lat, test_lng, "e")
  assert.Nil(t, err)
  assert.Equal(t, "fod_long_range_daily", resp.Forecasts[0].Class)
}
