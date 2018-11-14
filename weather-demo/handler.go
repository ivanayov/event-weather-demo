package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
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
	tmpl, err := template.ParseFiles("./template/index.html")
	if err != nil {
		return fmt.Sprintf("Internal server error with homepage: %s", err.Error())
	}
	var tpl bytes.Buffer

	events := []Event{
		{
			City: "Sofia",
			Name: "OpenFest",
		},
	}

	realEventBytes, readFileErr := ioutil.ReadFile("./template/events.json")
	if readFileErr != nil {
		return readFileErr.Error()
	}

	realEvents := []CalendarEvent{}
	unmarshalErr := json.Unmarshal(realEventBytes, &realEvents)
	if unmarshalErr != nil {
		return unmarshalErr.Error()
	}

	os.Stderr.Write(realEventBytes)

	err = tmpl.Execute(&tpl, HomepageTokens{
		Events: events,
	})
	if err != nil {
		return fmt.Sprintf("Internal server error with homepage template: %s", err.Error())
	}

	return string(tpl.Bytes())
}
