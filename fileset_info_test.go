package toolbox_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
	"testing"
)

func TestNewFileSetInfoInfo(t *testing.T) {

	fileSetInfo, err := toolbox.NewFileSetInfo("./fileset_info_test/")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 2, len(fileSetInfo.FilesInfo()))

	fileInfo := fileSetInfo.FileInfo("user_test.go")
	assert.NotNil(t, fileInfo)
	assert.False(t, fileInfo.HasStructInfo("F"))
	assert.True(t, fileInfo.HasStructInfo("User"))

	assert.Equal(t, 2, len(fileInfo.Structs()))

	address := fileSetInfo.Struct("Address")
	assert.NotNil(t, address)

	address2 := fileSetInfo.Struct("Address2")
	assert.Nil(t, address2)

	userInfo := fileInfo.Struct("User")
	assert.NotNil(t, userInfo)

	assert.True(t, userInfo.HasField("ID"))
	assert.True(t, userInfo.HasField("Name"))
	assert.False(t, userInfo.HasField("FF"))

	assert.Equal(t, 8, len(userInfo.Fields()))

	idInfo := userInfo.Field("ID")
	assert.True(t, idInfo.IsPointer)
	assert.Equal(t, "int", idInfo.TypeName)
	assert.Equal(t, true, idInfo.IsPointer)

	dobInfo := userInfo.Field("DateOfBirth")
	assert.True(t, dobInfo.IsStruct)
	assert.Equal(t, "time.Time", dobInfo.TypeName)
	assert.Equal(t, "time", dobInfo.TypePackage)

	assert.Equal(t, "`foo=\"bar\"`", dobInfo.Tag)

	addressPointer := userInfo.Field("AddressPointer")
	assert.NotNil(t, addressPointer)
	assert.Equal(t, true, addressPointer.IsStruct)
	assert.Equal(t, "Address", addressPointer.TypeName)

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

	assert.Equal(t, 3, len(userInfo.Receivers()))
	fmt.Printf("!%v!\n", userInfo.Receivers()[0].Name)

	assert.True(t, userInfo.HasReceiver("Test"))
	assert.True(t, userInfo.HasReceiver("Test2"))

	receiver := userInfo.Receiver("Test")
	assert.NotNil(t, receiver)

}
