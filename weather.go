package weather

import (
  "encoding/json"
  "errors"
  "fmt"
  "net/url"
  "strconv"
  //log "github.com/sirupsen/logrus"
  "net/http"
)

// For weather.com/wunderground api, night follows day.
// This means when retrieving a forecast in the middle of a day,
// there is data for night but not day part
// of the day on which the forecast is retrieved.

type Metadata struct {
  // ex: "en-US"
  Language string `json:"language"`
  // ex: "1532212454337:-56034070"
  TransactionId string `json:"transaction_id"`
  // ex: "1"
  Version string `json:"version"`
  // Latitude, either rounded to 2 decimal places or taken from the
  // weather station supplying the data rather than the request lat/lng
  // ex: 40.75
  Latitude float64 `json:"latitude"`
  // Longitude, either rounded to 2 decimal places or taken from the
  // weather station supplying the data rather than the request lat/lng
  // ex: -74
  Longitude float64 `json:"longitude"`
  // ex: "e"
  Units string `json:"units"`
  // ex: 1532213015
  ExpireTimeGmt int64 `json:"expire_time_gmt"`
  // ex: 200
  StatusCode int `json:"status_code"`
}

type DaypartForecast struct {
  // UTC timestamp for the forecast, e.g. 1531782000
  FcstValid int64 `json:"fcst_valid"`
  // ISO8601 time for the forecast, e.g. "2018-07-16T19:00:00-0400"
  FcstValidLocal string `json:"fcst_valid_local"`

  // "D" for day, "N" for night
  DayInd string `json:"day_ind"`
  // "Tonight", "Tomorrow", "Wednesday"
  DaypartName string `json:"daypart_name"`
  // "Monday night", "Tuesday night", "Tuesday"
  LongDaypartName string `json:"long_daypart_name"`
  // This is sometimes the same as DaypartName and sometimes same as
  // LongDaypartName
  AltDaypartName string `json:"alt_daypart_name"`

  // Number of this forecast in the returned data, starting with 1.
  // Forecasts for day parts (this struct) and days overall have separate numbering.
  // Forecast for today has num=1. If today only has a night day part,
  // that night's forecast would have num=1 as well.
  // Forecast for tomorrow will have num=2, tomorrow's day num=2,
  // tomorrow's night num=3. The day after tomorrow will have num=3 for the
  // entire day, num=4 for the day part, num=5 for the night part.
  Num int `json:"num"`

  // Max temperature for day day part, min temperature for night day part: 70
  Temp int `json:"temp"`
  // Temperature phrase incorporating whether this is a maximum or minimum
  // temperature: "Low near 70F.",
  TempPhrase string `json:"temp_phrase"`

  // Cloud cover in percent? ex: 5, 74
  Clds int `json:"clds"`

  // Precipitation type: "rain"
  PrecipType string `json:"precip_type"`
  // Probability of precipitation, in pecent: 90
  Pop int `json:"pop"`
  // Probability of precipitation phrase; "Chance of rain 100%."
  PopPhrase string `json:"pop_phrase"`
  // Precipitation accumulation, ex: "Rainfall around a half an inch."
  AccumulationPhrase string `json:"accumulation_phrase"`
  // ex: "A stray shower or thunderstorm is possible."
  // ex: "Locally heavy rainfall possible."
  Qualifier *string `json:"qualifier"`

  // Thunder possibility flag: 0-2
  ThunderEnum int `json:"thunder_enum"`
  // Thunder possibility phrase:
  // ThunderEnum=0: "No thunder"
  // ThunderEnum=1: "Thunder possible"
  // ThunderEnum=2: "Thunder expected"
  ThunderEnumPhrase string `json:"thunder_enum_phrase"`

  // Wind speed: 12
  // This should be within the wind speed range given by the narrative
  Wspd int `json:"wspd"`
  // Wind direction in degrees: 211
  Wdir int `json:"wdir"`
  // Wind direction as a string: SSW
  WdirCardinal string `json:"wdir_cardinal"`
  // Wind information phrase: "Winds SSW at 10 to 15 mph."
  WindPhrase string `json:"wind_phrase"`

  // ex: "Sct T-Storms"
  Phrase12char string `json:"phrase_12char"`
  // ex: "Sct Thunderstorms"
  Phrase22char string `json:"phrase_22char"`
  // ex: "Scattered Thunderstorms"
  Phrase32char string `json:"phrase_32char"`
  // ex: "Scattered"
  SubphrasePt1 string `json:"subphrase_pt1"`
  // ex: "T-Storms"
  SubphrasePt2 string `json:"subphrase_pt2"`
  // Always "" in data I've seen
  SubphrasePt3 string `json:"subphrase_pt3"`

  // ex: Scattered thunderstorms"
  Shortcast string `json:"shortcast"`
  // ex: "Variable clouds with scattered thunderstorms. High 81F. Winds S at 5 to 10 mph. Chance of rain 60%."
  Narrative string `json:"narrative"`

  Qpf float64 `json:"qpf"`
  // may be int
  SnowQpf    float64 `json:"snow_qpf"`
  SnowRange  string  `json:"snow_range"`
  SnowPhrase string  `json:"snow_phrase"`
  SnowCode   string  `json:"snow_code"`
  // this was always null even when qualifier is present, don't know type
  QualifierCode *string `json:"qualifier_code"`

  // ex: 7.9
  UvIndexRaw float64 `json:"uv_index_raw"`
  // UvIndexRaw rounded to an integer? ex: 8
  UvIndex   int `json:"uv_index"`
  UvWarning int `json:"uv_warning"`
  // ex: "Very High"
  UvDesc string `json:"uv_desc"`

  // ex: 5
  GolfIndex *int `json:"golf_index"`
  // ex: "Fair"
  // "" when GolfIndex is null
  GolfCategory string `json:"golf_category"`

  // ex: "wx1600"
  Wxman string `json:"wxman"`
  // ex: 82
  Hi int `json:"hi"`
  // ex: 72
  Wc int `json:"wc"`
  // ex: 76
  Rh int `json:"rh"`
  // ex: "D16:DA07:X3700380043:S380043:TL72:W08R04:P9041"
  VocalKey string `json:"vocal_key"`

  // ex: 30
  // Maps to https://icons.wxug.com/i/c/v4/30.svg
  IconCode int `json:"icon_code"`
  // ex: 3809
  IconExtd int `json:"icon_extd"`
}

type Forecast10 struct {
  // Type of forecast, "fod_long_range_daily" for this data
  Class string `json:"class"`
  // UTC timestamp: 1531769805
  ExpireTimeGmt int64 `json:"expire_time_gmt"`
  // UTC timestamp: 1531911600
  FcstValid int64 `json:"fcst_valid"`
  // ISO8601 local time: "2018-07-18T07:00:00-0400"
  FcstValidLocal string `json:"fcst_valid_local"`
  // Day of week, e.g. "Monday", "Tuesday"
  Dow string `json:"dow"`
  // Number of this forecast in the returned data, starting with 1.
  // Forecasts for days (this struct) and day parts have separate numbering.
  // Forecast for today has num=1. If today only has a night day part,
  // that night's forecast would have num=1 as well.
  // Forecast for tomorrow will have num=2, tomorrow's day num=2,
  // tomorrow's night num=3. The day after tomorrow will have num=3 for the
  // entire day, num=4 for the day part, num=5 for the night part.
  Num int `json:"num"`
  // Same as Day.Temp, can be null if there is no day data
  // which will happen when a forecast is retreived late enough in the day
  MaxTemp *int `json:"max_temp"`
  // Same as Night.Temp
  MinTemp int `json:"min_temp"`
  // was always null
  Torcon *string `json:"torcon"`
  // was always null
  Stormcon *string `json:"stormcon"`
  // was always null
  Blurb *string `json:"blurb"`
  // was always null
  BlurbAuthor *string `json:"blurb_author"`
  // ex: 5
  LunarPhaseDay int `json:"lunar_phase_day"`
  // ex: "Waxing Crescent"
  LunarPhase string `json:"lunar_phase"`
  // ex: "WXC"
  LunarPhaseCode string `json:"lunar_phase_code"`
  // ISO8601 local time: "2018-07-17T05:22:44-0400"
  Sunrise string `json:"sunrise"`
  // ISO8601 local time: "2018-07-17T20:17:41-0400"
  Sunset string `json:"sunset"`
  // ISO8601 local time: "2018-07-17T10:41:01-0400"
  Moonrise string `json:"moonrise"`
  // ISO8601 local time: "2018-07-17T23:32:29-0400"
  Moonset string `json:"moonset"`
  // assuming *string
  QualifierCode *string `json:"qualifier_code"`
  // assuming *string
  // This was null even when a day part forecast had a qualifier
  Qualifier *string `json:"qualifier"`
  // Narrative for the entire day (both day parts), in particular
  // it includes both high and low temperatures.
  // ex: "Times of sun and clouds. Highs in the upper 70s and lows in the mid 60s."
  Narrative string  `json:"narrative"`
  Qpf       float64 `json:"qpf"`
  // may be int
  SnowQpf    float64          `json:"snow_qpf"`
  SnowRange  string           `json:"snow_range"`
  SnowPhrase string           `json:"snow_phrase"`
  SnowCode   string           `json:"snow_code"`
  Night      DaypartForecast  `json:"night"`
  Day        *DaypartForecast `json:"day"`
}

type Forecast10Response struct {
  Metadata  Metadata     `json:"metadata"`
  Forecasts []Forecast10 `json:"forecasts"`
}

type HourlyForecast struct {
  // Type of forecast, "fod_short_range_hourly" or "fod_long_range_hourly"
  // for this data depending on how far out the forecast is from the current time
  Class string `json:"class"`
  // UTC timestamp: 1531769805
  ExpireTimeGmt int64 `json:"expire_time_gmt"`
  // UTC timestamp for the beginning of the interval that this forecast is for,
  // ex. 1531911600
  FcstValid int64 `json:"fcst_valid"`
  // ISO8601 local time for the beginning of the interval that this forecast is
  // for, ex. "2018-07-18T07:00:00-0400"
  FcstValidLocal string `json:"fcst_valid_local"`

  // Day of week, e.g. "Monday", "Tuesday"
  Dow string `json:"dow"`
  // "D" for day, "N" for night
  DayInd string `json:"day_ind"`

  // Number of this forecast in the returned data, starting with 1.
  Num int `json:"num"`

  // Temperature in requested units, e.g. 47
  Temp int `json:"temp"`
  // "Feels like" temperature in requested units, e.g. 40
  FeelsLike int `json:"feels_like"`
  // Dew point in requested units, e.g. 28
  Dewpt int `json:"dewpt"`

  // Cloud cover in percent? ex: 5, 74
  Clds int `json:"clds"`

  // Precipitation type: "rain"
  PrecipType string `json:"precip_type"`
  // Probability of precipitation, in pecent: 90
  Pop int `json:"pop"`

  // Wind speed: 12
  // This should be within the wind speed range given by the narrative
  Wspd int `json:"wspd"`
  // Wind direction in degrees: 211
  Wdir int `json:"wdir"`
  // Wind direction as a string: SSW
  WdirCardinal string `json:"wdir_cardinal"`
  // Wind gust speed: 20, or null
  Gust *int `json:"gust"`

  // Visibility?
  Vis float64 `json:"vis"`
  // Mean sea level pressure?
  Mslp float64 `json:"mslp"`

  UvIndexRaw float64 `json:"uv_index_raw"`
  UvIndex    int     `json:"uv_index"`
  UvWarning  int     `json:"uv_warning"`
  // ex: "Very High"
  UvDesc string `json:"uv_desc"`

  // ex: 5
  GolfIndex *int `json:"golf_index"`
  // ex: "Fair"
  // "" when GolfIndex is null
  GolfCategory string `json:"golf_category"`

  // ex: "Sct T-Storms"
  Phrase12char string `json:"phrase_12char"`
  // ex: "Sct Thunderstorms"
  Phrase22char string `json:"phrase_22char"`
  // ex: "Scattered Thunderstorms"
  Phrase32char string `json:"phrase_32char"`
  // ex: "Scattered"
  SubphrasePt1 string `json:"subphrase_pt1"`
  // ex: "T-Storms"
  SubphrasePt2 string `json:"subphrase_pt2"`
  // Always "" in data I've seen
  SubphrasePt3 string `json:"subphrase_pt3"`

  Qpf float64 `json:"qpf"`
  // may be int
  SnowQpf float64 `json:"snow_qpf"`

  // ex: "wx1600"
  Wxman string `json:"wxman"`
  // Same as Temp, apparently
  Hi int `json:"hi"`
  // Same as FeelsLike, apparently
  Wc int `json:"wc"`
  // ex: 76
  Rh       int `json:"rh"`
  Severity int `json:"severity"`

  // ex: 30
  // Maps to https://icons.wxug.com/i/c/v4/30.svg
  IconCode int `json:"icon_code"`
  // ex: 3809
  IconExtd int `json:"icon_extd"`
}

type HourlyForecastResponse struct {
  Metadata  Metadata         `json:"metadata"`
  Forecasts []HourlyForecast `json:"forecasts"`
}

type Wwir struct {
  // Type of forecast, "fod_short_range_wwir" for this data
  Class string `json:"class"`
  // UTC timestamp: 1531769805
  ExpireTimeGmt int64 `json:"expire_time_gmt"`
  // UTC timestamp: 1531911600
  // FcstValid seems to precede ExpireTimeGmt in this data
  FcstValid int64 `json:"fcst_valid"`
  // ISO8601 local time: "2018-07-18T07:00:00-0400"
  FcstValidLocal string `json:"fcst_valid_local"`
  // ex: 1
  OverallType int `json:"overall_type"`
  // ex: "Expect occasional rain to continue for the next several hours."
  Phrase string `json:"phrase"`
  // ex: "Rain will continue."
  TersePhrase string `json:"terse_phrase"`
  // Template from which Phrase is constructed. Sometimes there are
  // no variables to substitute and Phrase is the same as PhraseTemplate
  // ex: "Expect occasional rain to continue for the next several hours."
  PhraseTemplate string `json:"phrase_template"`
  // Template from which TersePhrase is constructed. Sometimes there are
  // no variables to substitute and TersePhrase is the same as TersePhraseTemplate
  // ex: "Rain will continue."
  TersePhraseTemplate string `json:"terse_phrase_template"`
  // Always null in data I have seen
  PrecipDay *string `json:"precip_day"`
  // Always null in data I have seen
  PrecipTime24hr *string `json:"precip_time_24hr"`
  // Always null in data I have seen
  PrecipTime12hr *string `json:"precip_time_12hr"`
  // Always null in data I have seen
  PrecipTimeIso *string `json:"precip_time_iso"`
  // ex: "EDT"
  TimeZoneAbbrv *string `json:"time_zone_abbrv"`
}

type WwirResponse struct {
  Metadata Metadata `json:"metadata"`
  Forecast Wwir     `json:"forecast"`
}

type UnitObservation struct {
  Temp      int  `json:"temp"`
  FeelsLike int  `json:"feels_like"`
  Wspd      int  `json:"wspd"`
  Gust      *int `json:"gust"`
  // Visibility?
  Vis float64 `json:"vis"`
  // Mean sea level pressure?
  Mslp             float64 `json:"mslp"`
  Altimeter        float64 `json:"altimeter"`
  Ceiling          float64 `json:"ceiling"`
  Dewpt            int     `json:"dewpt"`
  Rh               int     `json:"rh"`
  Wc               int     `json:"wc"`
  Hi               int     `json:"hi"`
  TempChange24hour int     `json:"temp_change_24hour"`
  TempMax24hour    int     `json:"temp_max_24hour"`
  TempMin24hour    int     `json:"temp_min_24hour"`
  Pchange          float64 `json:"pchange"`
  // These could be ints but more likely are all floats
  Snow1hour    float64 `json:"snow_1hour"`
  Snow6hour    float64 `json:"snow_6hour"`
  Snow24hour   float64 `json:"snow_24hour"`
  SnowMtd      float64 `json:"snow_mtd"`
  SnowSeason   float64 `json:"snow_season"`
  SnowYtd      float64 `json:"snow_ytd"`
  Snow3day     float64 `json:"snow_3day"`
  Snow7day     float64 `json:"snow_7day"`
  Precip1hour  float64 `json:"precip_1hour"`
  Precip6hour  float64 `json:"precip_6hour"`
  Precip24hour float64 `json:"precip_24hour"`
  PrecipMtd    float64 `json:"precip_mtd"`
  PrecipYtd    float64 `json:"precip_ytd"`
  Precip3day   float64 `json:"precip_3day"`
  Precip7day   float64 `json:"precip_7day"`
  // assuming *string, was always null
  ObsQualifier100char *string `json:"obs_qualifier_100char"`
  // assuming *string, was always null
  ObsQualifier50char *string `json:"obs_qualifier_50char"`
  // assuming *string, was always null
  ObsQualifier32char *string `json:"obs_qualifier_32char"`
}

type Observation struct {
  // "observation"
  Class string `json:"class"`
  // UTC timestamp: 1531769805
  ExpireTimeGmt int64 `json:"expire_time_gmt"`
  // UTC timestamp: 1531911600
  ObsTime int64 `json:"obs_time"`
  // ISO8601 local time: "2018-07-18T07:00:00-0400"
  ObsTimeLocal string `json:"obs_time_local"`
  // Day of week, e.g. "Monday", "Tuesday"
  Dow string `json:"dow"`
  // "D" for day, "N" for night
  DayInd string `json:"day_ind"`

  // Wind direction in degrees: 211
  Wdir int `json:"wdir"`
  // Wind direction as a string: SSW
  WdirCardinal string `json:"wdir_cardinal"`
  // ISO8601 local time: "2018-07-17T05:22:44-0400"
  Sunrise string `json:"sunrise"`
  // ISO8601 local time: "2018-07-17T20:17:41-0400"
  Sunset string `json:"sunset"`
  // Pressure tendency, ex: 2
  PtendCode int `json:"ptend_code"`
  // ex: "Falling"
  PtendDesc string `json:"ptend_desc"`
  // ex: "Cloudy"
  SkyCover string `json:"sky_cover"`
  // ex: "BKN"
  // Note this appears to be different from Clds in a forecast
  Clds string `json:"clds"`

  // ex: "Sct T-Storms"
  Phrase12char string `json:"phrase_12char"`
  // ex: "Sct Thunderstorms"
  Phrase22char string `json:"phrase_22char"`
  // ex: "Scattered Thunderstorms"
  Phrase32char string `json:"phrase_32char"`

  UvIndex   int `json:"uv_index"`
  UvWarning int `json:"uv_warning"`
  // ex: "Very High"
  UvDesc string `json:"uv_desc"`

  // ex: "wx1600"
  Wxman string `json:"wxman"`
  // assuming *string
  ObsQualifierCode *string `json:"obs_qualifier_code"`
  // assuming *string
  ObsQualifierSeverity *string `json:"obs_qualifier_severity"`
  // ex: "OT73:OX2600"
  VocalKey string `json:"vocal_key"`

  // ex: 30
  // Maps to https://icons.wxug.com/i/c/v4/30.svg
  IconCode int `json:"icon_code"`
  // ex: 3809
  IconExtd int `json:"icon_extd"`

  // units=e|a
  Imperial *UnitObservation `json:"imperial"`
  // units=m|a
  Metric *UnitObservation `json:"metric"`
  // units=s|a
  MetricSi *UnitObservation `json:"metric_si"`
  // units=h|a
  UkHybrid *UnitObservation `json:"uk_hybrid"`
}

type CurrentResponse struct {
  Metadata    Metadata    `json:"metadata"`
  Observation Observation `json:"observation"`
}

type Client struct {
  api_key     string
  http_client http.Client
}

func NewClient(api_key string) Client {
  return Client{
    api_key,
    http.Client{},
  }
}

func (c *Client) make_api_request(url string, payload interface{}) error {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return errors.New("Could not send request: " + err.Error())
  }

  res, err := c.http_client.Do(req)
  if err != nil {
    return errors.New("Could not read response: " + err.Error())
  }

  defer res.Body.Close()

  dec := json.NewDecoder(res.Body)
  err = dec.Decode(payload)
  if err != nil {
    return errors.New("Could not decode: " + err.Error())
  }

  return nil
}

func (c *Client) doGetForecast10(url string) (*Forecast10Response, error) {
  var payload Forecast10Response
  err := c.make_api_request(url, &payload)
  if err != nil {
    return nil, err
  }
  return &payload, nil
}

func (c *Client) doGetHourlyForecast(url string) (*HourlyForecastResponse, error) {
  var payload HourlyForecastResponse
  err := c.make_api_request(url, &payload)
  if err != nil {
    return nil, err
  }
  return &payload, nil
}

func (c *Client) doGetCurrent(url string) (*CurrentResponse, error) {
  var payload CurrentResponse
  err := c.make_api_request(url, &payload)
  if err != nil {
    return nil, err
  }
  return &payload, nil
}

func (c *Client) doGetWwir(url string) (*WwirResponse, error) {
  var payload WwirResponse
  err := c.make_api_request(url, &payload)
  if err != nil {
    return nil, err
  }
  return &payload, nil
}

//func (c *Client) GetForecast5ByLocation(lat float64, lng float64, units string) (*Forecast5Response, error) {
//  url := c.make_api_url(lat, lng, "forecast/daily/5day", units)
//  return c.doGetForecast5(url)
//}

func (c *Client) GetForecast10ByLocation(lat float64, lng float64, units string) (*Forecast10Response, error) {
  url := c.make_api_url(lat, lng, "forecast/daily/10day", units)
  return c.doGetForecast10(url)
}

func (c *Client) GetHourlyForecast240ByLocation(lat float64, lng float64, units string) (*HourlyForecastResponse, error) {
  url := c.make_api_url(lat, lng, "forecast/hourly/240hour", units)
  return c.doGetHourlyForecast(url)
}

func (c *Client) GetCurrentByLocation(lat float64, lng float64, units string) (*CurrentResponse, error) {
  url := c.make_api_url(lat, lng, "observations/current", units)
  return c.doGetCurrent(url)
}

func (c *Client) GetWwirByLocation(lat float64, lng float64, units string) (*WwirResponse, error) {
  url := c.make_api_url(lat, lng, "forecast/wwir", units)
  return c.doGetWwir(url)
}

func (c *Client) make_api_url(lat float64, lng float64, path_fragment string, units string) string {
  if units == "" {
    units = "e"
  }
  url := fmt.Sprintf("https://api.weather.com/v1/geocode/%f/%f/%s.json?apiKey=%s&units=%s",
    lat, lng,
    path_fragment,
    url.PathEscape(c.api_key), url.PathEscape(units))
  //log.Debug(url)
  return url
}

func format_float(f float64) string {
  return strconv.FormatFloat(f, 'f', -1, 32)
}
