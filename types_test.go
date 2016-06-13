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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/viant/toolbox"
)

type User2 struct {
	Name        string    `column:"name"`
	DateOfBirth time.Time `column:"date" dateLayout:"2006-01-02 15:04:05.000000"`
	Id          int       `autoincrement:"true"`
	Other       string    `transient:"true"`
}

func AssertEqual(test *testing.T, actual interface{}, expected interface{}, message string) {
	if actual != expected {
		test.Fatalf("Failed to "+message+" expected:%s, got %s:", expected, actual)
	}
}

func TestAssertKind(test *testing.T) {
	toolbox.AssertKind(User2{}, reflect.Struct, "user")
	toolbox.AssertKind((*User2)(nil), reflect.Ptr, "user")

	defer func() {
		if err := recover(); err != nil {
			expected := "Failed to check: User - expected kind: ptr but found struct (toolbox_test.User2)"
			actual := fmt.Sprintf("%v", err)
			AssertEqual(test, actual, expected, "Assert Kind")
		}
	}()
	toolbox.AssertKind(User2{}, reflect.Ptr, "User")
}

func TestAssertPointerKind(test *testing.T) {
	toolbox.AssertPointerKind(&User2{}, reflect.Struct, "user")
	toolbox.AssertPointerKind((*User2)(nil), reflect.Struct, "user")
}
