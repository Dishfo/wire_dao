package gen_dao

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

type DaoModel struct {
	Package string
	DMType  *ast.TypeSpec
	Methods *DaoMethod
	//parse info
	Name string
}

type DaoMethod struct {
	Function *ast.FuncDecl
	//parse info
	Name string

	Argument []*DaoMethodArgument
	Ret      []*DaoMethodRet
}

type DaoMethodRet struct {
	Name string
	Type *ast.TypeSpec
}

type DaoMethodArgument struct {
	Name string
	Type *ast.TypeSpec
}

type UsedType int

const (
	ReceiveTypeDef    UsedType = 0x1
	ReceiveTypePtr    UsedType = 0x2
	ReceiveTypeBase   UsedType = 0x3
	ReceiveTypeArray  UsedType = 0x4
	ReceiveTypeMap    UsedType = 0x5
	ReceiveTypeIntr   UsedType = 0x6
	ReceiveTypeExpand UsedType = 0x7
	ReceiveTypeFunc   UsedType = 0x8
)

type UsedTypeDesc struct {
	Name            string //this not a real type name
	TypeDescription *TypeDescription
}

type FunctionReceive struct {
	UsedTypeDesc
}

func (m *FunctionReceive) TypeName() string {
	return m.TypeDescription.TypeName
}

type FunctionArgument struct {
	UsedTypeDesc
}

type ImportDescription struct {
	Path string
	Name string
}

type FunctionRet struct {
	UsedTypeDesc
}

type FunctionDescription struct {
	Package  string
	FuncName string
	FuncDecl *ast.FuncDecl
	Receive  *FunctionReceive
	Ret      []*FunctionRet
	Params   []*FunctionArgument
}

func (m *FunctionDescription) ReceiveType() string {
	if m.Receive == nil {
		return ""
	}

	return m.Receive.TypeDescription.TypeName
}

func (m *FunctionDescription) ID() string {

	if m.FuncDecl.Recv == nil {
		return fmt.Sprintf("%s.%s", m.Package, m.FuncDecl.Name.Name)
	}

	return fmt.Sprintf("%s.%s.%s", m.Package, m.Receive.TypeName(), m.FuncDecl.Name.Name)
}

type BuiltinType string

type TypeDescription struct {
	Package   string
	PackageSt *packages.Package
	//define type
	TypeSpec *ast.TypeSpec
	TypeName string

	ItemType *TypeDescription

	BuiltinType BuiltinType

	KeyType *TypeDescription
	ValType *TypeDescription

	Type UsedType

	//description define type
}

type parsedField struct {
	pack *packages.Package

	rawField *ast.Field
	name     string
	typ      UsedType

	//other pack
	selItems []string

	//local pack
	spec *ast.TypeSpec

	//array type
	itemType *parsedField

	//base type
	baseTypeName string

	//map type
	keyType *parsedField
	valType *parsedField
}

func (m *parsedField) Description() *TypeDescription {
	if m == nil {
		return nil
	}
	ret := &TypeDescription{
		PackageSt:   m.pack,
		Type:        m.typ,
		ItemType:    m.itemType.Description(),
		BuiltinType: BuiltinType(m.baseTypeName),
		KeyType:     m.keyType.Description(),
		ValType:     m.valType.Description(),
		TypeSpec:    m.spec,
	}

	if ret.PackageSt != nil {
		ret.Package = ret.PackageSt.ID
	}

	if ret.TypeSpec != nil && ret.TypeSpec.Name != nil {
		ret.TypeName = ret.TypeSpec.Name.Name
	}

	return ret
}

func (s *Scanner) findTypeSpec(name string, pack *packages.Package) *ast.TypeSpec {
	for _, syntax := range pack.Syntax {
		for _, decl := range syntax.Decls {
			GenDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if GenDecl.Tok != token.TYPE {
				continue
			}

			typeSpec, ok := GenDecl.Specs[0].(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name == name {
				return typeSpec
			}
		}
	}

	return nil
}

func (s *Scanner) findSelType(selItems []string) (pack *packages.Package, typeSpec *ast.TypeSpec) {
	imports := s.filesImport[s.curFile]
	for _, description := range imports {
		if description.Name != selItems[0] {
			continue
		}

		pack := s.packagesIndex[description.Path]

		return pack, s.findTypeSpec(selItems[1], pack)

	}

	return nil, nil
}
