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
	"github.com/viant/toolbox"
)

type IMessage interface {
	Message() string
}

type Message struct {
	message string
}

func (this Message) Message() string {
	return this.message
}



func TestContext(t *testing.T) {
	context := toolbox.NewContext()
	message1 := Message{message: "abc"}

	//operate on pointer test
	assert.False(t, context.Contains((*Message)(nil)), "Should not have message in context")
	err := context.Put((*Message)(nil), &message1)
	assert.Nil(t, err)
	assert.True(t, context.Contains((*Message)(nil)), "Should have meesage in context")
	assert.True(t, context.Contains(&Message{}), "Should have meesage in context")


	value, err := context.GetRequired((*Message)(nil))
	assert.Nil(t, err)
	m1 := value.(*Message)
	assert.Equal(t, "abc", m1.message, "should have the same value field")
	m1.message = "xyz"
	assert.Equal(t, "xyz", message1.message, "should have the same value field")


	err = context.Put((*IMessage)(nil), &message1)
	assert.Nil(t, err)
	m2 := context.GetOptional((*IMessage)(nil)).(*IMessage)
	assert.Equal(t, "xyz", (*m2).Message(), "should have the same value field")


	//operate on struct passing by copy does not enable global changes
	assert.False(t, context.Contains(Message{}), "Should not have message in context")
	err = context.Put(Message{}, message1)
	assert.Nil(t, err)
	m3 := context.GetOptional(Message{}).(Message)
	assert.Equal(t, "xyz", m3.message, "should have the same value field")
	m3.message = "123"
	assert.Equal(t, "123", m3.message, "should have the same value field")
	assert.Equal(t, "xyz", m1.message, "should have the same value field")


	err = context.Put(Message{}, message1)
	assert.NotNil(t, err, "Key is already in context")

	err = context.Replace(1, "abc")
	assert.NotNil(t, err, "Incompatible type")

	removed := context.Remove((*IMessage)(nil))
	assert.NotNil(t, removed)

	_, err = context.GetRequired("abc")
	assert.NotNil(t, err)


}
