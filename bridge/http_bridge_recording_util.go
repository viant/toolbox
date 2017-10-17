package bridge

import (
	"path"
	"fmt"
	"github.com/viant/toolbox"
	"io/ioutil"
	"bytes"
)

//RecordedHttpTrip represents a recorded http trip
type RecordedHttpTrip struct {
	Request  *HttpRequest
	Response *HttpResponse
}


//ReadRecordedHttpTrips scans provided directory for bridge.HttpRequest-%v.json and  bridge.HttpResponse-%v.json pairs
func ReadRecordedHttpTrips(directory string) ([]*RecordedHttpTrip, error) {
	var result = make([]*RecordedHttpTrip, 0)
	var requestTemplatePath = path.Join(directory, "bridge.HttpRequest-%v.json")
	var responseTemplatePath = path.Join(directory, "bridge.HttpResponse-%v.json")

	requests, err := readAll(requestTemplatePath, func() interface{} {
		return &HttpRequest{}
	})
	if err != nil {
		return nil, err
	}
	responses, err := readAll(responseTemplatePath, func() interface{} {
		return &HttpResponse{}
	})
	if err != nil {
		return nil, err
	}

	if len(requests) != len(responses) {
		return nil, fmt.Errorf("Request and Response count does not match req:%v, resp:%v ", len(requests), len(responses))
	}

	for i:=0;i<len(requests);i++ {
		var ok bool
		var trip = &RecordedHttpTrip{}
		trip.Request, ok = requests[i].(*HttpRequest)
		if ! ok {
			return nil, fmt.Errorf("EXpected HttpRequest but had %T",  requests[i])
		}
		if i < len(responses) {
			trip.Response, ok = responses[i].(*HttpResponse)
			if ! ok {
				return nil, fmt.Errorf("EXpected HttpRequest but had %T", responses[i])
			}
		}
		result = append(result, trip)
	}

	return result, nil
}

func readAll(pathTemplate string, provider func() interface{}) ([]interface{}, error) {
	var result = make([]interface{}, 0)
	for i := 0; ; i++ {
		filename := fmt.Sprintf(pathTemplate, i)
		if ! toolbox.FileExists(filename) {
			break
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		aStruct := provider()
		err = toolbox.NewJSONDecoderFactory().Create(bytes.NewReader(data)).Decode(aStruct)
		if err != nil {
			return nil, err
		}
		result = append(result, aStruct)
	}
	return result, nil
}

