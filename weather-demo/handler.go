// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
)

type CalendarEvent struct {
	Content  string
	Location string
}

// HomepageTokens contains tokens for Golang HTML template
type HomepageTokens struct {
	Events []Event
}

type Event struct {
	City    string
	Name    string
	Weather string
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
	events := []Event{}
	for _, realEvent := range realEvents {
		events = append(events, Event{City: realEvent.Location, Name: realEvent.Content})
	}

	err = tmpl.Execute(&tpl, HomepageTokens{
		Events: events,
	})
	if err != nil {
		return fmt.Sprintf("Internal server error with homepage template: %s", err.Error())
	}

	return string(tpl.Bytes())
}
