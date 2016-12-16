package toolbox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
)

//FieldInfo represents a filed info
type FieldInfo struct {
	Name          string
	TypeName      string
	KeyTypeName   string
	ValueTypeName string
	TypePackage   string
	IsMap         bool
	IsChannel     bool
	IsSlice       bool
	IsStruct      bool
	IsPointer     bool
	Tag           string
	Comment       string
}

//NewFieldInfo creates a new field info.
func NewFieldInfo(field *ast.Field) *FieldInfo {
	result := &FieldInfo{
		Name:     "",
		TypeName: types.ExprString(field.Type),
	}

	if len(field.Names) > 0 {
		result.Name = field.Names[0].Name
	}
	_, result.IsMap = field.Type.(*ast.MapType)
	_, result.IsSlice = field.Type.(*ast.ArrayType)
	_, result.IsPointer = field.Type.(*ast.StarExpr)
	_, result.IsChannel = field.Type.(*ast.ChanType)
	if selector, ok := field.Type.(*ast.SelectorExpr); ok {
		result.TypePackage = types.ExprString(selector.X)
		result.IsStruct = true
	}

	if result.IsPointer {
		if pointerExpr, casted := field.Type.(*ast.StarExpr); casted {
			if identExpr, ok := pointerExpr.X.(*ast.Ident); ok {
				result.TypeName = identExpr.Name
				if reflect.ValueOf(identExpr.Obj).Elem().Kind() == reflect.Struct {
					result.IsStruct = true
				}

			}
		}
	}

	if field.Tag != nil {
		result.Tag = field.Tag.Value
	}
	if ident, ok := field.Type.(*ast.Ident); ok {
		kind := ""
		if ident.Obj != nil {
			result.IsStruct = kind == "type"
		}

	}
	if mapType, ok := field.Type.(*ast.MapType); ok {
		result.KeyTypeName = types.ExprString(mapType.Key)
		result.ValueTypeName = types.ExprString(mapType.Value)
	}
	return result
}

//FunctionInfo represents a function info
type FunctionInfo struct {
	Name             string
	ReceiverTypeName string
	ParameterFields  []*FieldInfo
	ResultsFields    []*FieldInfo
}

//NewFunctionInfo create a new function
func NewFunctionInfo(funcDeclaration *ast.FuncDecl) *FunctionInfo {
	result := &FunctionInfo{
		Name:            "",
		ParameterFields: make([]*FieldInfo, 0),
		ResultsFields:   make([]*FieldInfo, 0),
	}

	if funcDeclaration.Name != nil {
		result.Name = funcDeclaration.Name.Name
	}
	if funcDeclaration.Recv != nil {
		receiverType := funcDeclaration.Recv.List[0].Type
		if ident, ok := receiverType.(*ast.Ident); ok {
			result.ReceiverTypeName = ident.Name
		} else if startExpr, ok := receiverType.(*ast.StarExpr); ok {
			if ident, ok := startExpr.X.(*ast.Ident); ok {
				result.ReceiverTypeName = ident.Name
			}
		}
	}
	return result
}

//StructInfo represents a struct info
type StructInfo struct {
	Name            string
	Package         string
	FileName        string
	Comment         string
	Settings        map[string]string
	fields          []*FieldInfo
	indexedField    map[string]*FieldInfo
	receivers       []*FunctionInfo
	indexedReceiver map[string]*FunctionInfo
}

//AddFields appends fileds to structinfo
func (s *StructInfo) AddFields(fields ...*FieldInfo) {
	s.fields = append(s.fields, fields...)
	for _, field := range fields {
		s.indexedField[field.Name] = field
	}
}

//Field returns filedinfo for supplied file name
func (s *StructInfo) Field(name string) *FieldInfo {
	return s.indexedField[name]
}

//Fields returns all fields
func (s *StructInfo) Fields() []*FieldInfo {
	return s.fields
}

//HasField returns true if struct has passed in field.
func (s *StructInfo) HasField(name string) bool {
	_, found := s.indexedField[name]
	return found
}

//Receivers returns struct functions
func (s *StructInfo) Receivers() []*FunctionInfo {
	return s.receivers
}

//Receiver returns receiver for passed in name
func (s *StructInfo) Receiver(name string) *FunctionInfo {
	return s.indexedReceiver[name]
}

//HasReceiver returns true if receiver is defined for struct
func (s *StructInfo) HasReceiver(name string) bool {
	_, found := s.indexedReceiver[name]
	return found
}

//AddReceivers adds receiver for the struct
func (s *StructInfo) AddReceivers(receivers ...*FunctionInfo) {
	s.receivers = append(s.receivers, receivers...)
	for _, receiver := range receivers {
		s.indexedReceiver[receiver.Name] = receiver
	}
}

//NewStructInfo creates a new struct info
func NewStructInfo(name string) *StructInfo {
	return &StructInfo{Name: name,
		fields:          make([]*FieldInfo, 0),
		receivers:       make([]*FunctionInfo, 0),
		indexedReceiver: make(map[string]*FunctionInfo),
		indexedField:    make(map[string]*FieldInfo),
		Settings:        make(map[string]string)}
}

//FileInfo represent hold definition about all defined structs and its receivers in a file
type FileInfo struct {
	basePath            string
	filename            string
	structs             map[string]*StructInfo
	functions           map[string][]*FunctionInfo
	packageName         string
	currentStructInfo   *StructInfo
	fileSet             *token.FileSet
	currentFunctionInfo *FunctionInfo
}

//Struct returns a struct info for passed in name
func (f *FileInfo) Struct(name string) *StructInfo {
	return f.structs[name]
}

//Struct returns a struct info for passed in name
func (f *FileInfo) addFunction(funcion *FunctionInfo) {
	functions, found := f.functions[funcion.ReceiverTypeName]
	if !found {
		functions = make([]*FunctionInfo, 0)
		f.functions[funcion.ReceiverTypeName] = functions
	}
	f.functions[funcion.ReceiverTypeName] = append(f.functions[funcion.ReceiverTypeName], funcion)
}

//Struct returns all struct info
func (f *FileInfo) Structs() []*StructInfo {
	var result = make([]*StructInfo, 0)
	for _, v := range f.structs {
		result = append(result, v)
	}
	return result
}

//HasStructInfo returns truc if struct info is defined in a file
func (f *FileInfo) HasStructInfo(name string) bool {
	_, found := f.structs[name]
	return found
}

//readComment reads comment from the position
func (v *FileInfo) readComment(pos token.Pos) string {
	position := v.fileSet.Position(pos)
	fileName := path.Join(v.basePath, v.filename)
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic("Unable to open file " + fileName)
	}
	line := strings.Split(string(content), "\n")[position.Line-1]
	commentPosition := strings.LastIndex(line, "//")
	if commentPosition != -1 {
		return line[commentPosition+2:]
	}
	return ""
}

//toFieldInfoSlice convers filedList to FiledInfo slice.
func toFieldInfoSlice(source *ast.FieldList) []*FieldInfo {
	var result = make([]*FieldInfo, 0)
	if source == nil || len(source.List) == 0 {
		return result
	}
	for _, fields := range source.List {
		result = append(result, NewFieldInfo(fields))
	}
	return result
}

//Visit visits ast node to extract struct details from the passed file
func (v *FileInfo) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		//	fmt.Printf("visit %v %T\n", node, node)
		switch value := node.(type) {
		case *ast.TypeSpec:
			typeName := value.Name.Name
			structInfo := NewStructInfo(typeName)
			structInfo.Package = v.packageName
			structInfo.FileName = v.filename
			v.currentStructInfo = structInfo
			v.structs[typeName] = structInfo
		case *ast.StructType:
			v.currentStructInfo.Comment = v.readComment(value.Pos())
			v.currentStructInfo.AddFields(toFieldInfoSlice(value.Fields)...)

		case *ast.FuncDecl:
			functionInfo := NewFunctionInfo(value)
			v.currentFunctionInfo = functionInfo
			if len(functionInfo.ReceiverTypeName) > 0 {
				v.addFunction(functionInfo)
			}
		case *ast.FuncType:

			if v.currentFunctionInfo != nil {
				if value.Params != nil {
					v.currentFunctionInfo.ParameterFields = toFieldInfoSlice(value.Params)
				}
				if value.Results != nil {
					v.currentFunctionInfo.ResultsFields = toFieldInfoSlice(value.Results)
				}
				v.currentFunctionInfo = nil
			}
		}
	}
	return v
}

//Visit creates a new file info.
func NewFileInfo(basePath, packageName, filename string, fileSet *token.FileSet) *FileInfo {
	result := &FileInfo{
		basePath:    basePath,
		filename:    filename,
		packageName: packageName,
		structs:     make(map[string]*StructInfo),
		functions:   make(map[string][]*FunctionInfo),
		fileSet:     fileSet}
	return result
}

//FileSetInfo represents a fileset info storing information about go file with their struct definition
type FileSetInfo struct {
	files map[string]*FileInfo
}

//FileInfo returns fileinfo for supplied file name
func (f *FileSetInfo) FileInfo(name string) *FileInfo {
	return f.files[name]
}

//FilesInfo returns all files info.
func (f *FileSetInfo) FilesInfo() map[string]*FileInfo {
	return f.files
}

//StructInfo returns struct info for passed in struct name.
func (f *FileSetInfo) Struct(name string) *StructInfo {
	for _, v := range f.files {
		if v.HasStructInfo(name) {
			return v.Struct(name)
		}
	}
	return nil
}

//NewFileSetInfo creates a new fileset info
func NewFileSetInfo(baseDir string) (*FileSetInfo, error) {
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, baseDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse path %v: %v", baseDir, err)
	}

	var result = &FileSetInfo{
		files: make(map[string]*FileInfo),
	}
	for packageName, pkg := range pkgs {
		for filename, file := range pkg.Files {
			filename := path.Base(filename)
			fileInfo := NewFileInfo(baseDir, packageName, filename, fileSet)
			ast.Walk(fileInfo, file)
			result.files[filename] = fileInfo
		}
	}

	for _, fileInfo := range result.files {

		for k, functionsInfo := range fileInfo.functions {
			structInfo := result.Struct(k)
			if structInfo != nil {
				structInfo.AddReceivers(functionsInfo...)
			}
		}

	}
	return result, nil
}
