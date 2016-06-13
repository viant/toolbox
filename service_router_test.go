/*
 *
 *
 * Copyright 2012-2016 Viant.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  use this file except in compliance with the License. You may obtain a copy of
 *  the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  License for the specific language governing permissions and limitations under
 *  the License.
 *
 */
package toolbox_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"fmt"
	"time"
	"github.com/viant/toolbox"
)


type ReverseService struct {}

func (this ReverseService) Reverse(values []int)  []int {
	var result = make([]int, 0)
	for i := len(values) - 1;i>= 0;i-- {
		result = append(result, values[i])
	}
	return result
}

func (this ReverseService) Reverse2(values []int)  []int {
	var result = make([]int, 0)
	for i := len(values) - 1;i>= 0;i-- {
		result = append(result, values[i])
	}
	return result
}



func StartServer(port string, t * testing.T) {
	service := ReverseService{}
	router := toolbox.NewServiceRouter(
			toolbox.ServiceRouting{
				HTTPMethod:"GET",
				URI:"/v1/reverse/{ids}",
				Handler:service.Reverse,
				Parameters:[]string{"ids"},
			},
			toolbox.ServiceRouting{
				HTTPMethod:"POST",
				URI:"/v1/reverse/",
				Handler:service.Reverse,
				Parameters:[]string{"ids"},
			},
	)

	http.HandleFunc("/v1/reverse/", func (writer http.ResponseWriter, reader *http.Request) {
		err := router.Route(writer, reader)
		assert.Nil(t, err)
	})

	fmt.Printf("Started test server on port %v\n", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}









func TestServiceRouter(t *testing.T) {
	go func() {
		StartServer("8082", t)
	}()

	time.Sleep(2 * time.Second)
	var result = make([]int, 0)
	{

		err:= toolbox.RouteToService("get", "http://127.0.0.1:8082/v1/reverse/1,7,3", nil, &result)
		if err != nil {
			t.Errorf("Failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)

	}

	{

		err:= toolbox.RouteToService("post", "http://127.0.0.1:8082/v1/reverse/", []int{1, 7, 3}, &result)
		if err != nil {
			t.Errorf("Failed to send get request  %v", err)
		}
		assert.EqualValues(t, []int{3, 7, 1}, result)
	}
}
