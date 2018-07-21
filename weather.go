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

type ForecastResponseMetadata struct {
  Language      string  `json:"language"`
  TransactionId string  `json:"transaction_id"`
  Version       string  `json:"version"`
  Latitude      float64 `json:"latitude"`
  Longitude     float64 `json:"longitude"`
  Units         string  `json:"units"`
  ExpireTimeGmt int64   `json:"expire_time_gmt"`
  StatusCode    int     `json:"status_code"`
}

type ForecastResponseDaypart struct {
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
  Phrase12Char string `json:"phrase_12char"`
  // ex: "Sct Thunderstorms"
  Phrase22Char string `json:"phrase_22char"`
  // ex: "Scattered Thunderstorms"
  Phrase32Char string `json:"phrase_32char"`
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

  // ex: 3809
  IconExtd int `json:"icon_extd"`
  // ex: 47
  IconCode int `json:"icon_code"`
}

type ForecastResponseForecast struct {
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
  SnowQpf    float64                  `json:"snow_qpf"`
  SnowRange  string                   `json:"snow_range"`
  SnowPhrase string                   `json:"snow_phrase"`
  SnowCode   string                   `json:"snow_code"`
  Night      ForecastResponseDaypart  `json:"night"`
  Day        *ForecastResponseDaypart `json:"day"`
}

type Forecast10Response struct {
  Metadata  ForecastResponseMetadata   `json:"metadata"`
  Forecasts []ForecastResponseForecast `json:"forecasts"`
}

type Client struct {
  api_key     string
  http_client http.Client
}

func NewClient(api_key string) (*Client, error) {
  client := Client{
    api_key,
    http.Client{},
  }
  return &client, nil
}

func (c *Client) doGetForecast10(url string) (*Forecast10Response, error) {
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return nil, errors.New("Could not send request:" + err.Error())
    return nil, err
  }

  res, err := c.http_client.Do(req)
  if err != nil {
    return nil, errors.New("Could not read response:" + err.Error())
  }

  defer res.Body.Close()

  var payload Forecast10Response
  dec := json.NewDecoder(res.Body)
  err = dec.Decode(&payload)
  if err != nil {
    return nil, errors.New("Could not decode forecast:" + err.Error())
  }

  return &payload, nil
}

func (c *Client) GetForecast10ByLocation(lat float64, lng float64, units string) (*Forecast10Response, error) {
if units == "" { units ="e" }
  url := fmt.Sprintf("https://api.weather.com/v1/geocode/%f/%f/forecast/daily/10day.json?apiKey=%s&units=%s",
    url.PathEscape(format_float(lat)), url.PathEscape(format_float(lng)),
    url.PathEscape(c.api_key), url.PathEscape(units))
  //log.Debug(url)
  return c.doGetForecast10(url)
}

func format_float(f float64) string {
  return strconv.FormatFloat(f, 'f', -1, 32)
}
