package toolbox_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"log"
	"net/http"
	"testing"
	"time"
)

type ReverseService struct{}

func (this ReverseService) Reverse(values []int) []int {
	var result = make([]int, 0)
	for i := len(values) - 1; i >= 0; i-- {
		result = append(result, values[i])
	}

	return result
}

func (this ReverseService) Reverse2(values []int) []int {
	var result = make([]int, 0)
	for i := len(values) - 1; i >= 0; i-- {
		result = append(result, values[i])
	}
	return result
}

var ReverseInvoker = func(serviceRouting *toolbox.ServiceRouting, request *http.Request, response http.ResponseWriter, uriParameters map[string]interface{}) error {
	var function = serviceRouting.Handler.(func(values []int) []int)
	idsParam := uriParameters["ids"]
	ids := idsParam.([]string)
	values := make([]int, 0)
	for _, item := range ids {
		values = append(values, toolbox.AsInt(item))
	}
	var result = function(values)
	err := toolbox.WriteServiceRoutingResponse(response, request, serviceRouting, result)
	if err != nil {
		return err
	}
	return nil
}

func StartServer(port string, t *testing.T) {
	service := ReverseService{}
	router := toolbox.NewServiceRouter(
		toolbox.ServiceRouting{
			HTTPMethod: "GET",
			URI:        "/v1/reverse/{ids}",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "POST",
			URI:        "/v1/reverse/",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod: "DELETE",
			URI:        "/v1/delete/{ids}",
			Handler:    service.Reverse,
			Parameters: []string{"ids"},
		},
		toolbox.ServiceRouting{
			HTTPMethod:     "DELETE",
			URI:            "/v1/delete2/{ids}",
			Handler:        service.Reverse,
			Parameters:     []string{"ids"},
			HandlerInvoker: ReverseInvoker,
		},
	)

	http.HandleFunc("/v1/", func(writer http.ResponseWriter, reader *http.Request) {
		err := router.Route(writer, reader)
		assert.Nil(t, err)
	})

	fmt.Printf("Started test server on port %v\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func TestServiceRouter(t *testing.T) {
	go func() {
		StartServer("8082", t)
	}()

	time.Sleep(2 * time.Second)
	var result = make([]int, 0)
	{

		err := toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("Failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)

	}

	{

		err := toolbox.RouteToService("post", "http://127.0.0.1:8082/v1/reverse/", []int{1, 7, 3}, &result)
		if err != nil {
			t.Errorf("Failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}
	{

		err := toolbox.RouteToService("delete", "http://127.0.0.1:8082/v1/delete/", []int{1, 7, 3}, &result)
		if err != nil {
			t.Errorf("Failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}
	{

		err := toolbox.RouteToService("delete", "http://127.0.0.1:8082/v1/delete/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("Failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}

	{

		err := toolbox.RouteToService("werew", "http://127.0.0.1:8082/v1/delete/1,7,3", nil, &result)
		assert.NotNil(t, err)
	}

	{ //Test custom handler invocation without reflection

		err := toolbox.RouteToService("delete", "http://127.0.0.1:8082/v1/delete2/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("Failed to send delete request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}

}
