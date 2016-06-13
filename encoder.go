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
package toolbox

import (
	"encoding/json"
	"io"
)

//Encoder writes an instance to output stream
type Encoder interface {

	//Encode encodes  an instance to output stream
	Encode(object interface{}) error
}

//EncoderFactory create an encoder for an output stream
type EncoderFactory interface {
	//Create creates an encoder for an output stream
	Create(writer io.Writer) Encoder
}

type jsonEncoderFactory struct{}

func (e jsonEncoderFactory) Create(writer io.Writer) Encoder {
	return json.NewEncoder(writer)
}

//NewJSONEncoderFactory creates new NewJSONEncoderFactory
func NewJSONEncoderFactory() EncoderFactory {
	return &jsonEncoderFactory{}
}
