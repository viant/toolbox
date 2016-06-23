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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestDecoderFactory(t *testing.T) {
	reader := strings.NewReader("[1, 2, 3]")
	decoder := toolbox.NewJSONDecoderFactory().Create(reader)
	aSlice := make([]int, 0)
	err := decoder.Decode(&aSlice)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(aSlice))
}
