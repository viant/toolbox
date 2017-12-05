package bridge_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox/bridge"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func startTargetProxyTestEndpoint(port string, responses map[string]string) error {
	mux := http.NewServeMux()
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	mux.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {

		if strings.ToUpper(request.Method) == "POST" {
			response.WriteHeader(http.StatusOK)
			data, err := ioutil.ReadAll(request.Body)
			if err != nil {
				log.Fatal(err)
			}
			response.Write(data)
			return
		}

		if body, ok := responses[request.URL.Path]; ok {
			response.WriteHeader(http.StatusOK)
			response.Write([]byte(body))
		} else {
			response.WriteHeader(http.StatusNotFound)
		}
	})
	go http.Serve(listener, mux)
	return nil
}

func startTestHttpBridge(port string, factory bridge.HttpBridgeProxyHandlerFactory, routes ...*bridge.HttpBridgeProxyRoute) (map[string]http.Handler, error) {
	config := &bridge.HttpBridgeConfig{
		Endpoint: &bridge.HttpBridgeEndpointConfig{
			Port: port,
		},
		Proxy: &bridge.HttpBridgeProxyConfig{
			BufferPoolSize: 2,
			BufferSize:     8 * 1024,
		},
		Routes: routes,
	}
	httpBridge, err := bridge.NewHttpBridge(config, factory)
	if err != nil {
		return nil, err
	}
	go httpBridge.ListenAndServe()
	return httpBridge.Handlers, nil
}

func TestNewHttpBridge(t *testing.T) {

	for i := 0; i < 2; i++ {
		responses := make(map[string]string)
		responses[fmt.Sprintf("/test%v", i+1)] = fmt.Sprintf("Response1 from %v", 8088+i)
		err := startTargetProxyTestEndpoint(fmt.Sprintf("%v", 8088+i), responses)
		assert.Nil(t, err)
	}

	routes := make([]*bridge.HttpBridgeProxyRoute, 0)
	for i := 0; i < 2; i++ {
		targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%v/test%v", 8088+i, i+1))
		assert.Nil(t, err)
		if err != nil {
			log.Fatal(err)
		}
		route := &bridge.HttpBridgeProxyRoute{
			Pattern:   fmt.Sprintf("/test%v", i+1),
			TargetURL: targetURL,
		}
		routes = append(routes, route)
	}
	startTestHttpBridge("8085", bridge.NewProxyHandler, routes...)

	time.Sleep(1 * time.Second)
	//Test direct responses
	for i := 0; i < 2; i++ {
		response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%v/test%v", 8088+i, i+1))
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("Response1 from %v", 8088+i), string(content))

	}

	//Test proxy responses
	for i := 0; i < 2; i++ {
		response, err := http.Get(fmt.Sprintf("http://127.0.0.1:8085/test%v", i+1))
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("Response1 from %v", 8088+i), string(content))
	}
}

func TestNewHttpBridgeWithListeningHandler(t *testing.T) {
	basePort := 9098
	for i := 0; i < 2; i++ {
		responses := make(map[string]string)
		responses[fmt.Sprintf("/test%v", i+1)] = fmt.Sprintf("Response1 from %v", basePort+i)
		err := startTargetProxyTestEndpoint(fmt.Sprintf("%v", basePort+i), responses)
		assert.Nil(t, err)
	}

	var responses = make([]*http.Response, 0)
	var listener = func(request *http.Request, response *http.Response) {
		responses = append(responses, response)
	}
	routes := make([]*bridge.HttpBridgeProxyRoute, 0)
	for i := 0; i < 2; i++ {
		targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%v/test%v", basePort+i, i+1))
		assert.Nil(t, err)
		if err != nil {
			log.Fatal(err)
		}
		route := &bridge.HttpBridgeProxyRoute{
			Pattern:   fmt.Sprintf("/test%v", i+1),
			TargetURL: targetURL,
			Listener:  listener,
		}
		routes = append(routes, route)
	}
	_, err := startTestHttpBridge("9085", bridge.NewProxyRecordingHandler, routes...)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	//Test direct responses
	for i := 0; i < 2; i++ {
		response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%v/test%v", basePort+i, i+1))
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("Response1 from %v", basePort+i), string(content))

	}

	//Test proxy responses
	for i := 0; i < 2; i++ {
		response, err := http.Get(fmt.Sprintf("http://127.0.0.1:9085/test%v", i+1))
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("Response1 from %v", basePort+i), string(content))
	}

	{
		i := 0
		response, err := http.Post(fmt.Sprintf("http://127.0.0.1:9085/test%v", i+1), "text/json", strings.NewReader("{\"a\":1}"))
		assert.Nil(t, err)
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, "{\"a\":1}", string(content))
	}

	{
		assert.Equal(t, 3, len(responses))
		time.Sleep(1 * time.Second)
		request := responses[0].Request
		assert.Equal(t, "/test1", request.URL.Path)

		response := responses[0]
		content, err := ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, "Response1 from 9098", string(content))

		response = responses[2]
		content, err = ioutil.ReadAll(response.Body)
		assert.Nil(t, err)
		assert.Equal(t, "{\"a\":1}", string(content))
	}

}
