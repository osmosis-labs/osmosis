package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/osmosis-labs/osmosis/v15/cmd/modulegen/templates"
)

var (
	protoTemplate template.Template
	xTemplate     template.Template
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

	// err := parseModuleTemplates()
	// if err != nil {
	// 	fmt.Println(errors.Wrap(err, "error in template parsing"))
	// 	return
	// }

	xYmls := crawlForXYMLs()
	for _, path := range xYmls {
		tmpDir := strings.Replace(path, ".yml", "_template.tmpl", 1)
		xTemplatePtr, err := template.ParseFiles(tmpDir)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error in template parsing"))
			return
		}
		xTemplate = *xTemplatePtr
		err = codegenXYml(path)
		if err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("error in code generating %s ", path)))
			return
		}
		fmt.Println("template file ", tmpDir, " successfully created")
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

func crawlForXYMLs() []string {
	xYmls := []string{}
	err := filepath.Walk("cmd/modulegen/templates/x/",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// if path (case insensitive) ends with query.yml, append path
			if strings.HasSuffix(strings.ToLower(path), ".yml") {
				xYmls = append(xYmls, path)
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
	return xYmls
}

func codegenXYml(filepath string) error {
	xYml, err := templates.ReadXYmlFile(filepath)
	if err != nil {
		return err
	}

	err = codegenXPackage(xYml, filepath)
	if err != nil {
		return err
	}
	return err
}

func codegenXPackage(xYml templates.XYml, filePath string) error {
	// create directory
	fsModulePath := templates.ParseFilePathFromImportPath(xYml.ModulePath)
	fsFolderPath, fsGoFilePath := templates.ParseXFilePath(filePath)
	if err := os.MkdirAll(fsModulePath+"/"+fsFolderPath, os.ModePerm); err != nil {
		// ignore directory already exists error
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	// generate file
	f, err := os.Create(fsModulePath + "/" + fsGoFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return xTemplate.Execute(f, xYml)
}
