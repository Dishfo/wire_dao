package gen_dao

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/**
scan all go file and get dao-model
*/

type Scanner struct {
	//options
	cfg           *ScanConfig
	objectsFilter []ObjectsFilter

	dirs           []string
	parsedDaoModel []*DaoModel
	allTypes       map[string][]TypeDescription
	allFunction    map[string][]FunctionDescription
	filesImport    map[string][]*ImportDescription
	packagesIndex  map[string]*packages.Package

	//gen process
	objects map[TypeDescription][]FunctionDescription

	//parse process
	curFile string
	curPack *packages.Package
}

type ScanConfig struct {
	TopDir string
}

func ScanFiles(cfg *ScanConfig) (*Scanner, error) {
	inst := Scanner{
		cfg:           cfg,
		allTypes:      map[string][]TypeDescription{},
		allFunction:   map[string][]FunctionDescription{},
		filesImport:   map[string][]*ImportDescription{},
		packagesIndex: map[string]*packages.Package{},
		objects:       map[TypeDescription][]FunctionDescription{},
		objectsFilter: []ObjectsFilter{
			gormObjFilter,
		},
	}

	err := inst.listAllDir()
	if err != nil {
		return nil, err
	}
	err = inst.loadPackages(inst.dirs)
	if err != nil {
		return nil, err
	}

	return &inst, nil
}

func (s *Scanner) _listAllDir(absPath string) (err error) {

	fptr, err := os.Open(absPath)
	if err != nil {
		return err
	}

	fileInfo, err := fptr.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return nil
	}

	s.dirs = append(s.dirs, filepath.Clean(absPath))

	subFiles, err := ioutil.ReadDir(absPath)
	if err != nil {
		return err
	}

	for _, subFile := range subFiles {
		subName := subFile.Name()
		subPath := filepath.Join(absPath, subName)
		err = s._listAllDir(subPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Scanner) listAllDir() (err error) {
	fptr, err := os.Open(s.cfg.TopDir)
	if err != nil {
		return err
	}

	fileInfo, err := fptr.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New("invalid top dir path")
	}

	return s._listAllDir(s.cfg.TopDir)
}

func (s *Scanner) loadPackages(dirs []string) (err error) {
	for _, dir := range dirs {
		err = s.parsePackage(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Scanner) parsePackage(absPath string) (err error) {
	ctx := context.TODO()
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.LoadAllSyntax,
		Dir:     absPath,
		Env:     os.Environ(),
	}

	loadPakcgaes, err := packages.Load(cfg)
	if err != nil {
		return err
	}

	for _, loadPackage := range loadPakcgaes {
		s.setCurPackage(loadPackage)
		err = s.extractDao(loadPackage)
		if err != nil {
			return err
		}
	}

	s.setCurPackage(nil)
	return nil
}

func (s *Scanner) setCurFile(fileName string) {
	s.curFile = fileName
}

func (s *Scanner) currentFile() string {
	return s.curFile
}

func (s *Scanner) setCurPackage(pack *packages.Package) {
	s.curPack = pack
}

func (s *Scanner) currentPackage() (pack *packages.Package) {
	return s.curPack
}

func (s *Scanner) extractDao(pack *packages.Package) (err error) {
	for _, p := range pack.Imports {
		s.packagesIndex[p.ID] = p
	}

	for i, sf := range pack.Syntax {
		fileName := pack.CompiledGoFiles[i]
		s.setCurFile(fileName)
		for _, impt := range sf.Imports {
			importDesc := &ImportDescription{
				Path: strings.Trim(impt.Path.Value, `"`),
			}

			if impt.Name != nil {
				importDesc.Name = impt.Name.Name
			} else {
				importDesc.Name = filepath.Base(importDesc.Path)
			}

			imports := s.filesImport[fileName]
			imports = append(imports, importDesc)
			s.filesImport[fileName] = imports
		}

		for _, decl := range sf.Decls {
			err = s.rememberDecl(pack, decl)
			if err != nil {
				return err
			}
		}
	}

	s.setCurFile("")

	return nil
}

func (s *Scanner) rememberDecl(pack *packages.Package, decl ast.Decl) (err error) {
	switch realDecl := decl.(type) {
	case *ast.GenDecl:
		if realDecl.Tok != token.TYPE {
			return nil
		}

		typeSpec := realDecl.Specs[0].(*ast.TypeSpec)
		types := s.allTypes[pack.ID]

		types = append(types, TypeDescription{
			TypeSpec:  typeSpec,
			Package:   pack.ID,
			PackageSt: pack,
			TypeName:  typeSpec.Name.Name,
		})
		s.allTypes[pack.ID] = types
	case *ast.FuncDecl:
		funcs := s.allFunction[pack.ID]
		var receive *FunctionReceive

		if realDecl.Recv != nil {
			field := realDecl.Recv.List[0]
			receive = s.parseReceiveField(field)
		}

		var comments []string
		if realDecl.Doc != nil {
			for _, comment := range realDecl.Doc.List {
				comments = append(comments, comment.Text)
			}
		}

		funcs = append(funcs, FunctionDescription{
			Package:  pack.ID,
			FuncName: realDecl.Name.Name,
			FuncDecl: realDecl,
			Receive:  receive,
			Comments: comments,
			Params:   s.parseFuncParams(pack, realDecl.Type.Params),
			Ret:      s.parseFuncResults(pack, realDecl.Type.Results),
		})

		s.allFunction[pack.ID] = funcs
	}
	return nil
}

func (s *Scanner) parseFuncParams(pack *packages.Package, fields *ast.FieldList) []*FunctionArgument {
	var ret []*FunctionArgument
	if fields == nil {
		return nil
	}

	var parseFields []*parsedField
	for _, field := range fields.List {
		parsedF := s.parseField(field)

		parseFields = append(parseFields, parsedF...)
	}

	for _, field := range parseFields {

		val := FunctionArgument{
			UsedTypeDesc: UsedTypeDesc{
				TypeDescription: field.Description(),
				Name:            field.name,
			},
		}

		ret = append(ret, &val)
	}

	return ret
}

func (s *Scanner) parseFuncResults(pack *packages.Package, fields *ast.FieldList) []*FunctionRet {
	var ret []*FunctionRet
	if fields == nil {
		return nil
	}

	var parseFields []*parsedField

	for _, field := range fields.List {
		parsedF := s.parseField(field)

		parseFields = append(parseFields, parsedF...)
	}

	//fmt.Println("aaaa ", len(parseFields))
	for _, field := range parseFields {

		val := FunctionRet{
			UsedTypeDesc: UsedTypeDesc{
				TypeDescription: field.Description(),
				Name:            field.name,
			},
		}

		ret = append(ret, &val)
	}

	return ret
}

func (s *Scanner) parseReceiveField(f *ast.Field) *FunctionReceive {
	var ret *FunctionReceive
	parsedF := s.parseField(f)[0]
	ret = &FunctionReceive{
		UsedTypeDesc: UsedTypeDesc{
			Name:            parsedF.name,
			TypeDescription: parsedF.Description(),
		},
	}
	return ret
}

func (s *Scanner) parseUsedExpr(x ast.Expr) *parsedField {
	//ast.SelectorExpr{}
	var ret *parsedField
	switch realExpr := x.(type) {
	case *ast.SelectorExpr:
		//outer type
		packName := realExpr.X.(*ast.Ident).Name
		typeName := realExpr.Sel.Name

		ret = &parsedField{
			typ:      ReceiveTypeDef,
			selItems: []string{packName, typeName},
		}
		pack, typeSpec := s.findSelType(ret.selItems)
		ret.spec = typeSpec
		ret.pack = pack

	case *ast.Ident:
		//local type
		if realExpr.Obj != nil {
			ret = &parsedField{
				pack: s.currentPackage(),
				typ:  ReceiveTypeDef,
				spec: realExpr.Obj.Decl.(*ast.TypeSpec),
			}
		} else {
			ret = &parsedField{
				typ:          ReceiveTypeBase,
				baseTypeName: realExpr.Name,
			}
		}
	case *ast.InterfaceType:
		ret = &parsedField{
			typ: ReceiveTypeIntr,
		}
	case *ast.ArrayType:
		pf := s.parseUsedExpr(realExpr.Elt)
		coped := &parsedField{
			typ:      ReceiveTypeArray,
			itemType: pf,
		}
		ret = coped
	case *ast.StarExpr:
		pf := s.parseUsedExpr(realExpr.X)
		coped := &parsedField{}
		*coped = *pf
		coped.typ = ReceiveTypePtr
		ret = coped
	case *ast.MapType:
		kt := s.parseUsedExpr(realExpr.Key)
		vt := s.parseUsedExpr(realExpr.Value)
		coped := &parsedField{
			typ:     ReceiveTypeMap,
			keyType: kt,
			valType: vt,
		}
		ret = coped
	case *ast.FuncType:
		coped := &parsedField{
			typ: ReceiveTypeFunc,
		}
		ret = coped
	default:
		log.Printf("can't handle this %v", x)
		fmt.Println("parsed failed :unhit ")
		//todo panic mapType
	}

	return ret
}

//todo think how to handle base type
func (s *Scanner) parseField(f *ast.Field) []*parsedField {
	var ret []*parsedField
	var targetCount = 1
	if len(f.Names) > 1 {
		targetCount = len(f.Names)
	}

	ret = make([]*parsedField, targetCount)
	names := fieldName(f)
	switch readField := f.Type.(type) {
	case *ast.StarExpr:
		//point type
		realField := s.parseUsedExpr(readField.X)
		realField.rawField = f
		for i, _ := range ret {
			coped := &parsedField{}
			*coped = *realField
			coped.typ = ReceiveTypePtr
			coped.name = names[i]
			coped.rawField = f
			ret[i] = coped
		}

	case *ast.Ident:
		//base type
		//fmt.Println("check this ", readField, f.Names)
		for i, _ := range ret {
			var typeSpec *ast.TypeSpec
			coped := &parsedField{
				rawField: f,
				name:     names[i],
			}

			if readField.Obj != nil {
				typeSpec = readField.Obj.Decl.(*ast.TypeSpec)
				coped.spec = typeSpec
				coped.pack = s.currentPackage()
				coped.typ = ReceiveTypeDef
			} else {
				coped.baseTypeName = readField.Name
				coped.typ = ReceiveTypeBase
			}
			ret[i] = coped
		}
	case *ast.ArrayType:
		for i, _ := range ret {
			pf := s.parseUsedExpr(readField.Elt)
			coped := &parsedField{
				rawField: f,
				name:     names[i],
				typ:      ReceiveTypeArray,
				itemType: pf,
			}
			ret[i] = coped
		}
	case *ast.Ellipsis:
		for i, _ := range ret {

			pf := s.parseUsedExpr(readField.Elt)
			coped := &parsedField{
				rawField: f,
				name:     names[i],
				typ:      ReceiveTypeExpand,
				itemType: pf,
			}
			ret[i] = coped
		}
	case *ast.InterfaceType:
		for i, _ := range ret {
			coped := &parsedField{
				rawField: f,
				name:     names[i],
				typ:      ReceiveTypeIntr,
			}
			ret[i] = coped
		}
	case *ast.SelectorExpr:
		realField := s.parseUsedExpr(readField.X)
		for i, _ := range ret {

			coped := &parsedField{}

			*coped = *realField
			coped.name = names[i]
			ret[i] = coped
		}
	case *ast.MapType:
		kt := s.parseUsedExpr(readField.Key)
		vt := s.parseUsedExpr(readField.Value)
		for i, _ := range ret {
			coped := &parsedField{
				rawField: f,
				name:     names[i],
				typ:      ReceiveTypeMap,
				keyType:  kt,
				valType:  vt,
			}
			ret[i] = coped
		}
	case *ast.FuncType:
		for i, _ := range ret {

			coped := &parsedField{}
			coped.name = names[i]
			coped.typ = ReceiveTypeFunc
			coped.rawField = f
			ret[i] = coped
		}
	default:
		fmt.Println("not hit", readField)
		//todo may panic
	}

	//fmt.Println("xxxx ", len(ret))
	return ret
}

func fieldName(f *ast.Field) []string {
	var names []string
	if len(f.Names) > 0 {
		for _, name := range f.Names {
			names = append(names, name.Name)
		}
	} else {
		return []string{""}
	}

	return names
}
