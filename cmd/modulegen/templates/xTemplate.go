package templates

type ModuleTemplate struct {
	FilePath     string
	ModulePath   string
	ModuleName   string
	SimtypesPath string
}

type KeeperTemplate struct {
	FilePath   string
	ModulePath string
	ModuleName string
}

func ModuleTemplateFromXYml(xYml XYml) ModuleTemplate {
	return ModuleTemplate{
		FilePath:     xYml.filePath,
		ModulePath:   xYml.ModulePath,
		ModuleName:   xYml.ModuleName,
		SimtypesPath: xYml.SimtypesPath,
	}
}

func KeeperTemplateFromXYml(xYml XYml) KeeperTemplate {
	return KeeperTemplate{
		FilePath:   xYml.filePath,
		ModulePath: xYml.ModulePath,
		ModuleName: xYml.ModuleName,
	}
}
