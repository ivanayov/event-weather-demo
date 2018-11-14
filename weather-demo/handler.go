// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type CalendarEvent struct {
	Content  string
	Location string
	Month    string
}

// HomepageTokens contains tokens for Golang HTML template
type HomepageTokens struct {
	Events []Event
}

type Event struct {
	City       string
	Name       string
	Weather    string
	WeatherPic string
}

type Forecast struct {
	Weather     Weather     `json:"weather"`
	Temperature Temperature `json:"main"`
	Wind        Wind        `json:"wind"`
}

type Weather struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type Temperature struct {
	TempMin float64 `json:"temp_min"`
	TempMax float64 `json:"temp_max"`
	Summary float64 `json:"temp"`
}

type Wind struct {
	Speed float64 `json:"speed"`
}

var WeatherPicMap = map[string]string{
	"Rain":    "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/160/google/146/cloud-with-rain_1f327.png",
	"Clear":   "http://www.clker.com/cliparts/I/P/P/k/H/G/sun-logo-hi.png",
	"Clouds":  "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/160/apple/118/cloud_2601.png",
	"Snow":    "http://png.clipart-library.com/images4/10/snowing-clipart-10.png",
	"Extreme": "https://i.pinimg.com/originals/a8/65/9d/a8659dbe09e5785523a85746044a0722.png",
	"Other":   "http://icongal.com/gallery/image/152928/status_scattered_showers_weather_day.png",
}

// Handle a serverless request
func Handle(req []byte) string {

	var err error
	tmpl, err := template.ParseFiles("./html/index.html")
	if err != nil {
		return fmt.Sprintf("Internal server error with homepage: %s", err.Error())
	}
	var tpl bytes.Buffer

	realEventBytes, readFileErr := ioutil.ReadFile("./html/events.json")
	if readFileErr != nil {
		return readFileErr.Error()
	}

	realEvents := []CalendarEvent{}
	unmarshalErr := json.Unmarshal(realEventBytes, &realEvents)
	if unmarshalErr != nil {
		return unmarshalErr.Error()
	}

	// os.Stderr.Write(realEventBytes)

	month := ""
	path := os.Getenv("Http_Path")
	pathArr := strings.Split(path, "/")
	if len(pathArr) > 1 {
		month = pathArr[1]
	}

	log.Printf("Path is %s\n", path)
	log.Printf("Month is %s\n", month)

	events := []Event{}
	for _, realEvent := range realEvents {
		if len(month) != 0 {
			if realEvent.Month != month {
				continue
			}
		}
		weatherForecast, err := getWeatherForecast(realEvent.Location)
		if err != nil {
			return err.Error()
		}
		temperature := fmt.Sprintf("%.1f ", weatherForecast.Temperature.Summary-272.15)

		var weatherPicSelector string
		if pictureURL, ok := WeatherPicMap[weatherForecast.Weather.Main]; ok {
			weatherPicSelector = pictureURL
		} else {
			weatherPicSelector = WeatherPicMap["Other"]
		}
		events = append(events, Event{City: realEvent.Location, Name: realEvent.Content, Weather: temperature, WeatherPic: weatherPicSelector})
	}

	err = tmpl.Execute(&tpl, HomepageTokens{
		Events: events,
	})
	if err != nil {
		return fmt.Sprintf("Internal server error with homepage template: %s", err.Error())
	}

	return string(tpl.Bytes())
}

func getWeatherForecast(city string) (*Forecast, error) {

	client := http.Client{}

	weatherAppGatewayURL := os.Getenv("weather_app_gateway_url")

	bodyReader := bytes.NewBuffer([]byte(city))
	req, _ := http.NewRequest(http.MethodGet, weatherAppGatewayURL+"/function/weather-app/"+city, bodyReader)
	res, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to get request from weather-app %s", err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code for function: weather-app` - %d\n", res.StatusCode)
		resBody, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("Error in getting weather forecast: %s", resBody)
	}
	resBody, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("Error while reading response body from weather-ap: %s", readErr)
	}

	forecast := &Forecast{}
	err = json.Unmarshal(resBody, forecast)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal weather forecast: %s", err)
	}
	return forecast, nil
}
