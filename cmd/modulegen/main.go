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

	protoYml := templates.ProtoYml{
		ModuleName: *moduleName,
		ModulePath: fmt.Sprintf("github.com/osmosis-labs/osmosis/v15/x/%s", *moduleName),
	}

	xYml := templates.XYml{
		ModuleName:   *moduleName,
		ModulePath:   fmt.Sprintf("github.com/osmosis-labs/osmosis/v15/x/%s", *moduleName),
		SimtypesPath: "github.com/osmosis-labs/osmosis/v15/simulation",
	}

	protoYmls := crawlForProtoTemplates()
	for _, path := range protoYmls {
		xTemplatePtr, err := template.ParseFiles(path)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error in template parsing"))
			return
		}
		xTemplate = *xTemplatePtr
		err = codegenProtoPackage(protoYml, path)
		if err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("error in code generating %s ", path)))
			return
		}
		fmt.Println("template file ", path, " successfully created")
	}

	xYmls := crawlForXTemplates()
	for _, path := range xYmls {
		xTemplatePtr, err := template.ParseFiles(path)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error in template parsing"))
			return
		}
		xTemplate = *xTemplatePtr
		err = codegenXPackage(xYml, path)
		if err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("error in code generating %s ", path)))
			return
		}
		fmt.Println("template file ", path, " successfully created")
	}
}

func crawlForXTemplates() []string {
	xYmls := []string{}
	err := filepath.Walk("cmd/modulegen/templates/x",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// if path (case insensitive) ends with query.yml, append path
			if strings.HasSuffix(strings.ToLower(path), ".tmpl") {
				xYmls = append(xYmls, path)
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
	return xYmls
}

func crawlForProtoTemplates() []string {
	xYmls := []string{}
	err := filepath.Walk("cmd/modulegen/templates/proto",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// if path (case insensitive) ends with query.yml, append path
			if strings.HasSuffix(strings.ToLower(path), ".tmpl") {
				xYmls = append(xYmls, path)
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
	return xYmls
}

func codegenXPackage(xYml templates.XYml, filePath string) error {
	// create directory
	fsModulePath := templates.ParseFilePathFromImportPath(xYml.ModulePath)
	fsFolderPath, fsGoFilePath := templates.ParseXFilePath(filePath)
	fmt.Println("result", fsModulePath, fsFolderPath, fsGoFilePath)
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

func codegenProtoPackage(protoYml templates.ProtoYml, filePath string) error {
	// create directory
	fsModulePath := "proto/osmosis/" + protoYml.ModuleName
	fsFolderPath, fsProtoFilePath := templates.ParseProtoFilePath(filePath)
	if err := os.MkdirAll(fsModulePath+"/"+fsFolderPath, os.ModePerm); err != nil {
		// ignore directory already exists error
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	// generate file
	f, err := os.Create(fsModulePath + "/" + fsProtoFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return xTemplate.Execute(f, protoYml)
}
