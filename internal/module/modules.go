package module

import "github.com/yannismate/gowlbot/internal/module/logging"

func GetRegisteredModules() []interface{} {
	var modules []interface{}

	modules = append(modules, logging.ProvideLoggingModule)

	return modules
}

func StartModules(logging logging.Module) {
	logging.Start()
}
