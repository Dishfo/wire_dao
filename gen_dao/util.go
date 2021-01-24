package gen_dao

type FunctionID struct {
	Package  string
	Receiver string
	Function string
}

func (s *Scanner) FindFunction(funcID FunctionID) (FunctionDescription, error) {

	var (
		pack     = funcID.Package
		function = funcID.Function
		receive  = funcID.Receiver
	)

	funcs, exist := s.allFunction[pack]
	if !exist {
		return FunctionDescription{}, ErrPackageNotFound
	}

	for _, description := range funcs {
		if description.FuncName != function {
			continue
		}
		//check is same receive type
		if len(receive) > 0 && description.ReceiveType() != receive {
			continue
		}
		//found function
		return description, nil
	}

	return FunctionDescription{}, ErrFuncNotFound
}
