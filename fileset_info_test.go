package toolbox_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestNewFileSetInfoInfo(t *testing.T) {
	if (32 << uintptr(^uintptr(0)>>63)) < 64 {
		t.Skip()
	}
	fileSetInfo, err := toolbox.NewFileSetInfo("./test/fileset_info/")
	if err != nil {
		panic(err)
	}
	assert.True(t, len(fileSetInfo.FilesInfo()) > 0)

	fileInfo := fileSetInfo.FileInfo("user.go")
	assert.NotNil(t, fileInfo)

	addresses := fileSetInfo.Type("Addresses")
	assert.NotNil(t, addresses)

	assert.False(t, fileInfo.HasType("F"))
	assert.True(t, fileInfo.HasType("User"))

	assert.Equal(t, 10, len(fileInfo.Types()))

	address := fileSetInfo.Type("Address")
	assert.NotNil(t, address)

	assert.Equal(t, 2, len(address.Fields()))
	country := address.Field("Country")
	assert.NotNil(t, country)
	assert.True(t, country.IsAnonymous)

	z := fileSetInfo.Type("Z")
	assert.NotNil(t, z)

	address2 := fileSetInfo.Type("Address2")
	assert.Nil(t, address2)

	userInfo := fileInfo.Type("User")
	assert.NotNil(t, userInfo)

	assert.True(t, userInfo.HasField("ID"))
	assert.True(t, userInfo.HasField("Name"))
	assert.False(t, userInfo.HasField("FF"))

	assert.Equal(t, 14, len(userInfo.Fields()))

	idInfo := userInfo.Field("ID")
	assert.True(t, idInfo.IsPointer)
	assert.Equal(t, "int", idInfo.TypeName)
	assert.Equal(t, true, idInfo.IsPointer)

	dobInfo := userInfo.Field("DateOfBirth")

	assert.Equal(t, "time.Time", dobInfo.TypeName)
	assert.Equal(t, "time", dobInfo.TypePackage)

	assert.Equal(t, "`foo=\"bar\"`", dobInfo.Tag)

	addressPointer := userInfo.Field("AddressPointer")
	assert.NotNil(t, addressPointer)
	assert.Equal(t, "Address", addressPointer.TypeName)


	aMapField := userInfo.Field("AMap1")
	assert.NotNil(t, aMapField)
	assert.EqualValues(t, "AMap1", aMapField.TypeName)
	aMapType := fileSetInfo.Type(aMapField.TypeName)
	assert.NotNil(t, aMapType)
	assert.True(t, aMapType.IsMap)
	assert.EqualValues(t, "string", aMapType.KeyTypeName)
	assert.EqualValues(t, "[]int", aMapType.ValueTypeName)


	aMapField2 := userInfo.Field("AMap2")
	assert.NotNil(t, aMapField2)
	assert.EqualValues(t, "AMap2", aMapField2.TypeName)
	aMapType2 := fileSetInfo.Type(aMapField2.TypeName)
	assert.NotNil(t, aMapType2)
	assert.True(t, aMapType2.IsMap)
	assert.EqualValues(t, "string", aMapType2.KeyTypeName)
	assert.EqualValues(t, "[]*Country", aMapType2.ValueTypeName)


	aMapField3 := userInfo.Field("AMap3")
	assert.NotNil(t, aMapField3)
	assert.EqualValues(t, "AMap3", aMapField3.TypeName)
	aMapType3 := fileSetInfo.Type(aMapField3.TypeName)
	assert.NotNil(t, aMapType3)
	assert.True(t, aMapType3.IsMap)
	assert.EqualValues(t, "string", aMapType3.KeyTypeName)
	assert.EqualValues(t, "*Country", aMapType3.ValueTypeName)




	cInfo := userInfo.Field("C")
	assert.True(t, cInfo.IsChannel)

	mInfo := userInfo.Field("M")
	assert.True(t, mInfo.IsMap)
	assert.Equal(t, "string", mInfo.KeyTypeName)
	assert.Equal(t, "[]string", mInfo.ValueTypeName)

	intsInfo := userInfo.Field("Ints")
	assert.True(t, intsInfo.IsSlice)
	assert.Equal(t, "my comments", userInfo.Comment)

	assert.False(t, userInfo.HasReceiver("Abc"))

	assert.True(t, len(userInfo.Receivers()) > 1)
	assert.True(t, userInfo.HasReceiver("Test"))
	assert.True(t, userInfo.HasReceiver("Test2"))

	receiver := userInfo.Receiver("Test")
	assert.NotNil(t, receiver)

	appointments := userInfo.Field("Appointments")
	assert.NotNil(t, appointments)
	assert.Equal(t, "time.Time", appointments.ComponentType)

}
