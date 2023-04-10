package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/pkg/errors"

	"github.com/osmosis-labs/osmosis/v15/cmd/modulegen/templates"
)

var (
	moduleTemplate template.Template
	protoTemplate  template.Template
	xTemplate      template.Template
)

func main() {
	// Define and parse the module name flag
	moduleName := flag.String("module_name", "", "The name of the module to be generated")
	flag.Parse()

	if *moduleName == "" {
		fmt.Println("Error: module_name flag is required")
		os.Exit(1)
	}

	// Define templates for proto and x directories
	// err := parseProtoTemplates()
	// if err != nil {
	// 	fmt.Println(errors.Wrap(err, "error in parsing proto templates"))
	// 	return
	// }

	// err = parseXTemplates()
	// if err != nil {
	// 	fmt.Println(errors.Wrap(err, "error in parsing x templates"))
	// 	return
	// }

	// Create directories for the new module
	// protoDir := filepath.Join("proto", "osmosis", *moduleName, "v1beta1")
	// err = os.MkdirAll(protoDir, os.ModePerm)
	// if err != nil {
	// 	panic(err)
	// }

	// xDir := filepath.Join("x", *moduleName)
	// err = os.MkdirAll(xDir, os.ModePerm)
	// if err != nil {
	// 	panic(err)
	// }

	// Create and write the generated files
	// protoFile, err := os.Create(filepath.Join(protoDir, fmt.Sprintf("%s.proto", *moduleName)))
	// if err != nil {
	// 	panic(err)
	// }
	// defer protoFile.Close()

	// xFile, err := os.Create(filepath.Join(xDir, fmt.Sprintf("%s.go", *moduleName)))
	// if err != nil {
	// 	panic(err)
	// }
	// defer xFile.Close()

	err := parseModuleTemplates()
	if err != nil {
		fmt.Println(errors.Wrap(err, "error in template parsing"))
		return
	}

	moduleYmls := "./cmd/modulegen/templates/module/module.yml"
	err = codegenModuleYml(moduleYmls)
	if err != nil {
		fmt.Println(errors.Wrap(err, fmt.Sprintf("error in code generating %s ", moduleYmls)))
	}
}

func parseProtoTemplates() error {
	protoTemplatePtr, err := template.ParseFiles("cmd/modulegen/templates/proto_template.tmpl")
	if err != nil {
		return err
	}
	protoTemplate = *protoTemplatePtr
	return nil
}

func parseXTemplates() error {
	xTemplatePtr, err := template.ParseFiles("cmd/querygen/templates/x_template.tmpl")
	if err != nil {
		return err
	}
	xTemplate = *xTemplatePtr
	return nil
}

func parseModuleTemplates() error {
	moduleTemplatePtr, err := template.ParseFiles("cmd/modulegen/templates/module/module_template.tmpl")
	if err != nil {
		return err
	}
	moduleTemplate = *moduleTemplatePtr
	return nil
}

func codegenModuleYml(filepath string) error {
	moduleYml, err := templates.ReadYmlFile(filepath)
	if err != nil {
		return err
	}

	err = codegenModulePackage(moduleYml)
	if err != nil {
		return err
	}
	return err
}

func codegenModulePackage(moduleYml templates.ModuleYml) error {
	moduleTemplateData := templates.ModuleTemplateFromModuleYml(moduleYml)

	// create directory
	fsModulePath := templates.ParseFilePathFromImportPath(moduleTemplateData.ModulePath)
	if err := os.MkdirAll(fsModulePath+"/module", os.ModePerm); err != nil {
		// ignore directory already exists error
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	// generate file
	f, err := os.Create(fsModulePath + "/module/module.go")
	if err != nil {
		return err
	}
	defer f.Close()

	return moduleTemplate.Execute(f, moduleTemplateData)
}
