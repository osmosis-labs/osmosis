package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/osmosis-labs/osmosis/v27/cmd/querygen/templates"
)

const V2 = "v2"

var grpcTemplate template.Template

func main() {
	err := parseTemplates()
	if err != nil {
		fmt.Println(errors.Wrap(err, "error in template parsing"))
		return
	}

	queryYMLs := crawlForQueryYMLs()
	for _, path := range queryYMLs {
		err := codegenQueryYml(path)
		if err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("error in code generating %s ", path)))
		}
	}
}

func parseTemplates() error {
	// Create a function to upper case the version suffix if it exists.
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	grpcTemplatePtr, err := template.New("grpc_template.tmpl").Funcs(funcMap).ParseFiles("cmd/querygen/templates/grpc_template.tmpl")
	if err != nil {
		return err
	}
	grpcTemplate = *grpcTemplatePtr
	return nil
}

func crawlForQueryYMLs() []string {
	queryYmls := []string{}
	err := filepath.Walk("./proto",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// if path (case insensitive) ends with query.yml, append path
			if strings.HasSuffix(strings.ToLower(path), "query.yml") {
				queryYmls = append(queryYmls, path)
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
	return queryYmls
}

func codegenQueryYml(filepath string) error {
	queryYml, err := templates.ReadYmlFile(filepath)
	if err != nil {
		return err
	}

	err = codegenGrpcPackage(queryYml)
	if err != nil {
		return err
	}
	return err
}

func codegenGrpcPackage(queryYml templates.QueryYml) error {
	grpcTemplateData := templates.GrpcTemplateFromQueryYml(queryYml)

	// If proto path contains v2 then add folder and template
	// suffix to properly package the files.
	grpcTemplateData.VersionSuffix = ""
	if strings.Contains(grpcTemplateData.ProtoPath, V2) {
		grpcTemplateData.VersionSuffix = V2
	}

	// create directory
	fsClientPath := templates.ParseFilePathFromImportPath(grpcTemplateData.ClientPath)
	if err := os.MkdirAll(fsClientPath+"/grpc"+grpcTemplateData.VersionSuffix, os.ModePerm); err != nil {
		// ignore directory already exists error
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}
	// generate file
	f, err := os.Create(fsClientPath + "/grpc" + grpcTemplateData.VersionSuffix + "/grpc_query.go")
	if err != nil {
		return err
	}
	defer f.Close()

	return grpcTemplate.Execute(f, grpcTemplateData)
}
