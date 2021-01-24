package gen_dao

import "errors"

//util error
var (
	ErrPackageNotFound   = errors.New("package has not been parsed")
	ErrFuncNotFound      = errors.New("function not exist ")
	ErrTypeNotFound      = errors.New("type not exist")
	ErrInvalidFunctionID = errors.New("invalid function id ")
)
