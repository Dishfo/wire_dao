package gen_dao

import (
	"log"
)

func (s *Scanner) ListAllType() {
	//for pack, types := range s.allTypes {
	//	fmt.Printf("under package :%s\n", pack)
	//	for _, description := range types {
	//		fmt.Printf("  %s\n",description.TypeName)
	//	}
	//	fmt.Println()
	//}
	s.rebuildObjs()
	var count int
	for typeDesc, funcdescriptions := range s.objects {
		log.Printf("under pack %s ,type : %s \n", typeDesc.Package, typeDesc.TypeName)
		log.Println("	list functions ")
		for _, funcdescription := range funcdescriptions {
			log.Printf("			functions :%s", funcdescription.ID())
		}

		count += 1
	}
	log.Printf("all dao count is %d\n", count)

}

func (s *Scanner) rebuildObjs() {
	//init map keys
	var packsTypeIndex = map[string]map[string]TypeDescription{}

	for packID, descriptions := range s.allTypes {
		for _, description := range descriptions {
			s.objects[description] = nil
			typesMap := packsTypeIndex[packID]
			if typesMap == nil {
				typesMap = map[string]TypeDescription{}
			}
			typesMap[description.TypeName] = description
			packsTypeIndex[packID] = typesMap
		}
	}

	for pack, descriptions := range s.allFunction {
		for _, description := range descriptions {
			if description.Receive == nil {
				continue
			}

			//find target receive type
			types := packsTypeIndex[pack]
			if types == nil {
				log.Printf("can't target pack %s \n", pack)
				continue
			}

			typ, exist := types[description.Receive.TypeDescription.TypeName]
			if !exist {
				log.Printf("can't target type %s in pack \n", description.Receive.TypeDescription.TypeName)
				continue
			}

			functions := s.objects[typ]
			functions = append(functions, description)
			s.objects[typ] = functions
		}
	}

	if s.objectsFilter != nil {
		var toDel []TypeDescription
		for description, descriptions := range s.objects {
			objset := map[TypeDescription][]FunctionDescription{
				description: descriptions,
			}
			for _, filter := range s.objectsFilter {
				if !filter(objset) {
					toDel = append(toDel, description)
				}
			}
		}
		log.Println("to delete ")
		for _, description := range toDel {
			log.Println(description.TypeName)
		}

		for _, description := range toDel {
			delete(s.objects, description)
		}
	}
}

type ObjectsFilter func(objSet map[TypeDescription][]FunctionDescription) bool

var gormObjFilter = func(objSet map[TypeDescription][]FunctionDescription) bool {
	for typeDesc, funcDescriptions := range objSet {
		if typeDesc.TypeName == "Session" {
			return false
		}

		for _, description := range funcDescriptions {
			ay := FunctionAnalysis{}
			ay.Analysis(description)
			if ay.HasUsePackage("github.com/jinzhu/gorm") {
				return true
			}

			if description.FuncDecl.Name.Name == "TableName" {
				return true
			}
		}

	}

	return false
}
