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
	assert.Equal(t, 1, len(fileSetInfo.FilesInfo()))

	fileInfo := fileSetInfo.FileInfo("user_test.go")
	assert.NotNil(t, fileInfo)
	assert.False(t, fileInfo.HasStructInfo("F"))
	assert.True(t, fileInfo.HasStructInfo("User"))

	address := fileSetInfo.Struct("Address")
	assert.NotNil(t, address)

	address2 := fileSetInfo.Struct("Address2")
	assert.Nil(t, address2)

	userInfo := fileInfo.Struct("User")
	assert.NotNil(t, userInfo)

	assert.True(t, userInfo.HasField("ID"))
	assert.True(t, userInfo.HasField("Name"))
	assert.False(t, userInfo.HasField("FF"))

	assert.Equal(t, 7, len(userInfo.Fields()))

	idInfo := userInfo.Field("ID")
	assert.True(t, idInfo.IsPointer)
	assert.Equal(t, "*int", idInfo.TypeName)

	dobInfo := userInfo.Field("DateOfBirth")
	assert.True(t, dobInfo.IsStruct)
	assert.Equal(t, "time.Time", dobInfo.TypeName)
	assert.Equal(t, "time", dobInfo.TypePackage)

	assert.Equal(t, "`foo=\"bar\"`", dobInfo.Tag)

	cInfo := userInfo.Field("C")
	assert.True(t, cInfo.IsChannel)

	mInfo := userInfo.Field("M")
	assert.True(t, mInfo.IsMap)

	intsInfo := userInfo.Field("Ints")
	assert.True(t, intsInfo.IsSlice)
	assert.Equal(t, "my comments", userInfo.Comment)

	assert.False(t, userInfo.HasReceiver("Abc"))

	assert.Equal(t, 2, len(userInfo.Receivers()))
	fmt.Printf("!%v!\n", userInfo.Receivers()[0].Name)

	assert.True(t, userInfo.HasReceiver("Test"))

	receiver := userInfo.Receiver("Test")
	assert.NotNil(t, receiver)

}
