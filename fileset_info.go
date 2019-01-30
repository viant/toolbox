package toolbox

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path"
	"strings"
)

//FieldInfo represents a filed info
type FieldInfo struct {
	Name               string
	TypeName           string
	ComponentType      string
	IsPointerComponent bool
	KeyTypeName        string
	ValueTypeName      string
	TypePackage        string
	IsAnonymous        bool
	IsMap              bool
	IsChannel          bool
	IsSlice            bool
	IsPointer          bool
	Tag                string
	Comment            string
}

//NewFieldInfo creates a new field info.
func NewFieldInfo(field *ast.Field) *FieldInfo {
	result := &FieldInfo{
		Name:     "",
		TypeName: types.ExprString(field.Type),
	}

	if len(field.Names) > 0 {
		result.Name = field.Names[0].Name
	} else {
		result.Name = strings.Replace(strings.Replace(result.TypeName, "[]", "", len(result.TypeName)), "*", "", len(result.TypeName))
		result.IsAnonymous = true
	}
	_, result.IsMap = field.Type.(*ast.MapType)
	var arrayType *ast.ArrayType
	if arrayType, result.IsSlice = field.Type.(*ast.ArrayType); result.IsSlice {
		switch x := arrayType.Elt.(type) {
		case *ast.Ident:
			result.ComponentType = x.Name
		case *ast.StarExpr:
			switch y := x.X.(type) {
			case *ast.Ident:
				result.ComponentType = y.Name
			case *ast.SelectorExpr:
				result.ComponentType = y.X.(*ast.Ident).Name + "." + y.Sel.Name
			}
			result.IsPointerComponent = true
		case *ast.SelectorExpr:
			result.ComponentType = x.X.(*ast.Ident).Name + "." + x.Sel.Name
		}
	}
	_, result.IsPointer = field.Type.(*ast.StarExpr)
	_, result.IsChannel = field.Type.(*ast.ChanType)
	if selector, ok := field.Type.(*ast.SelectorExpr); ok {
		result.TypePackage = types.ExprString(selector.X)
	}
	if result.IsPointer {
		if pointerExpr, casted := field.Type.(*ast.StarExpr); casted {
			if identExpr, ok := pointerExpr.X.(*ast.Ident); ok {
				result.TypeName = identExpr.Name
			}
		}
	} else if identExpr, ok := field.Type.(*ast.Ident); ok {
		result.TypeName = identExpr.Name
	}

	if field.Tag != nil {
		result.Tag = field.Tag.Value
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

//TypeInfo represents a struct info
type TypeInfo struct {
	Name                   string
	Package                string
	FileName               string
	Comment                string
	IsSlice                bool
	IsStruct               bool
	IsDerived              bool
	ComponentType          string
	IsPointerComponentType bool
	Derived                string
	Settings               map[string]string
	fields                 []*FieldInfo
	indexedField           map[string]*FieldInfo
	receivers              []*FunctionInfo
	indexedReceiver        map[string]*FunctionInfo
}

//AddFields appends fileds to structinfo
func (s *TypeInfo) AddFields(fields ...*FieldInfo) {
	s.fields = append(s.fields, fields...)
	for _, field := range fields {
		s.indexedField[field.Name] = field
	}
}

//Field returns filedinfo for supplied file name
func (s *TypeInfo) Field(name string) *FieldInfo {
	return s.indexedField[name]
}

//Fields returns all fields
func (s *TypeInfo) Fields() []*FieldInfo {
	return s.fields
}

//HasField returns true if struct has passed in field.
func (s *TypeInfo) HasField(name string) bool {
	_, found := s.indexedField[name]
	return found
}

//Receivers returns struct functions
func (s *TypeInfo) Receivers() []*FunctionInfo {
	return s.receivers
}

//Receiver returns receiver for passed in name
func (s *TypeInfo) Receiver(name string) *FunctionInfo {
	return s.indexedReceiver[name]
}

//HasReceiver returns true if receiver is defined for struct
func (s *TypeInfo) HasReceiver(name string) bool {
	_, found := s.indexedReceiver[name]
	return found
}

//AddReceivers adds receiver for the struct
func (s *TypeInfo) AddReceivers(receivers ...*FunctionInfo) {
	s.receivers = append(s.receivers, receivers...)
	for _, receiver := range receivers {
		s.indexedReceiver[receiver.Name] = receiver
	}
}

//NewTypeInfo creates a new struct info
func NewTypeInfo(name string) *TypeInfo {
	return &TypeInfo{Name: name,
		fields:          make([]*FieldInfo, 0),
		receivers:       make([]*FunctionInfo, 0),
		indexedReceiver: make(map[string]*FunctionInfo),
		indexedField:    make(map[string]*FieldInfo),
		Settings:        make(map[string]string)}
}

//FileInfo represent hold definition about all defined types and its receivers in a file
type FileInfo struct {
	basePath            string
	filename            string
	types               map[string]*TypeInfo
	functions           map[string][]*FunctionInfo
	packageName         string
	currentTypInfo      *TypeInfo
	fileSet             *token.FileSet
	currentFunctionInfo *FunctionInfo
}

//Type returns a type info for passed in name
func (f *FileInfo) Type(name string) *TypeInfo {
	return f.types[name]
}

//Type returns a struct info for passed in name
func (f *FileInfo) addFunction(funcion *FunctionInfo) {
	functions, found := f.functions[funcion.ReceiverTypeName]
	if !found {
		functions = make([]*FunctionInfo, 0)
		f.functions[funcion.ReceiverTypeName] = functions
	}
	f.functions[funcion.ReceiverTypeName] = append(f.functions[funcion.ReceiverTypeName], funcion)
}

//Type returns all struct info
func (f *FileInfo) Types() []*TypeInfo {
	var result = make([]*TypeInfo, 0)
	for _, v := range f.types {
		result = append(result, v)
	}
	return result
}

//HasType returns truc if struct info is defined in a file
func (f *FileInfo) HasType(name string) bool {
	_, found := f.types[name]
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
	for _, field := range source.List {

		result = append(result, NewFieldInfo(field))
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
			typeInfo := NewTypeInfo(typeName)
			typeInfo.Package = v.packageName
			typeInfo.FileName = v.filename

			switch typeValue := value.Type.(type) {
			case *ast.ArrayType:
				typeInfo.IsSlice = true
				if ident, ok := typeValue.Elt.(*ast.Ident); ok {
					typeInfo.ComponentType = ident.Name
				} else if startExpr, ok := typeValue.Elt.(*ast.StarExpr); ok {
					if ident, ok := startExpr.X.(*ast.Ident); ok {
						typeInfo.ComponentType = ident.Name
					}
					typeInfo.IsPointerComponentType = true
				}
			case *ast.StructType:
				typeInfo.IsStruct = true
			case *ast.Ident:
				typeInfo.Derived = typeValue.Name
				typeInfo.IsDerived = true
			}
			v.currentTypInfo = typeInfo
			v.types[typeName] = typeInfo
		case *ast.StructType:
			if v.currentTypInfo != nil {//TODO fixme - understand why current type would be nil
				v.currentTypInfo.Comment = v.readComment(value.Pos())
				v.currentTypInfo.AddFields(toFieldInfoSlice(value.Fields)...)
			}
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
		types:       make(map[string]*TypeInfo),
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

//TypeInfo returns type info for passed in type  name.
func (f *FileSetInfo) Type(name string) *TypeInfo {
	if pointerIndex := strings.LastIndex(name, "*"); pointerIndex != -1 {
		name = name[pointerIndex+1:]
	}
	for _, v := range f.files {
		if v.HasType(name) {
			return v.Type(name)
		}
	}
	return nil
}

//NewFileSetInfo creates a new fileset info
func NewFileSetInfo(baseDir string) (*FileSetInfo, error) {
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, baseDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %v: %v", baseDir, err)
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
			typeInfo := result.Type(k)
			if typeInfo != nil && typeInfo.IsStruct {
				typeInfo.AddReceivers(functionsInfo...)
			}
		}

	}
	return result, nil
}
