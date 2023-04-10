package templates

type ModuleTemplate struct {
	FilePath     string
	ModulePath   string
	ModuleName   string
	SimtypesPath string
}

func ModuleTemplateFromModuleYml(moduleYml ModuleYml) ModuleTemplate {
	return ModuleTemplate{
		FilePath:     moduleYml.filePath,
		ModulePath:   moduleYml.ModulePath,
		ModuleName:   moduleYml.ModuleName,
		SimtypesPath: moduleYml.SimtypesPath,
	}
}
