package bridge

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/toolbox"
	"io/ioutil"
	"path"
	"strings"
)

//RecordedHttpTrip represents a recorded http trip
type RecordedHttpTrip struct {
	Request  *HttpRequest
	Response *HttpResponse
}

//ReadRecordedHttpTripsWithPrefix scans provided directory for request, response  pairs
func ReadRecordedHttpTripsWithTemplate(directory string, requestTemplate, respTemplate string) ([]*RecordedHttpTrip, error) {
	if !strings.Contains(requestTemplate, "%") {
		return nil, errors.New("invalid request template: it could contains counter '%' expressions")
	}
	if !strings.Contains(respTemplate, "%") {
		return nil, errors.New("invalid response template: it could contains counter '%' expressions")
	}
	var result = make([]*RecordedHttpTrip, 0)
	var requestTemplatePath = path.Join(directory, requestTemplate)
	var responseTemplatePath = path.Join(directory, respTemplate)

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
		return nil, fmt.Errorf("request and Response count does not match req:%v, resp:%v ", len(requests), len(responses))
	}

	for i := 0; i < len(requests); i++ {
		var ok bool
		var trip = &RecordedHttpTrip{}
		trip.Request, ok = requests[i].(*HttpRequest)
		if !ok {
			return nil, fmt.Errorf("expected HttpRequest but had %T", requests[i])
		}
		if i < len(responses) {
			trip.Response, ok = responses[i].(*HttpResponse)
			if !ok {
				return nil, fmt.Errorf("expected HttpRequest but had %T", responses[i])
			}
		}
		result = append(result, trip)
	}
	return result, nil
}

//ReadRecordedHttpTrips scans provided directory for bridge.HttpRequest-%v.json and  bridge.HttpResponse-%v.json template pairs
func ReadRecordedHttpTrips(directory string) ([]*RecordedHttpTrip, error) {
	return ReadRecordedHttpTripsWithTemplate(directory, "bridge.HttpRequest-%v.json", "bridge.HttpResponse-%v.json")
}

func readAll(pathTemplate string, provider func() interface{}) ([]interface{}, error) {
	var result = make([]interface{}, 0)
	for i := 0; ; i++ {
		filename := fmt.Sprintf(pathTemplate, i)
		if !toolbox.FileExists(filename) {
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

//StartRecordingBridge start recording bridge proxy
func StartRecordingBridge(port string, outputDirectory string, routes ...*HttpBridgeProxyRoute) (*HttpBridge, error) {

	if len(routes) == 0 {
		routes = append(routes, &HttpBridgeProxyRoute{
			Pattern: "/",
		})
	}

	config := &HttpBridgeConfig{
		Endpoint: &HttpBridgeEndpointConfig{
			Port: port,
		},
		Proxy: &HttpBridgeProxyConfig{
			BufferPoolSize: 2,
			BufferSize:     8 * 1024,
		},
		Routes: routes,
	}

	for _, route := range routes {
		route.Listener = HttpFileRecorder(outputDirectory, false)
	}

	httpBridge, err := NewHttpBridge(config, NewProxyRecordingHandler)
	if err != nil {
		return nil, err
	}
	go httpBridge.ListenAndServe()
	return httpBridge, nil
}
