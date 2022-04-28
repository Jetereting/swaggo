package parser

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"
	"swaggo/swagger"
)

var (
	vendor  = ""
	goPaths = []string{}
	goRoot  = ""
	devMode bool
)

func init() {
	goPaths = filepath.SplitList(os.Getenv("GOPATH"))
	if len(goPaths) == 0 {
		panic("GOPATH environment variable is not set or empty")
	}
	goRoot = runtime.GOROOT()
	if goRoot == "" {
		panic("GOROOT environment variable is not set or empty")
	}
}

// Parse the project by args
func Parse(projectPath, swaggerGo, output, t string, dev bool) (err error) {
	absPPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	vendor = filepath.Join(absPPath, "vendor")
	devMode = dev

	sw := swagger.NewV2()
	if err = doc2Swagger(projectPath, swaggerGo, dev, sw); err != nil {
		return
	}

	// 分页特殊处理 （因为interface 问题，后台无法优雅的生成对应模型）
	var paginationDefs []string
	// 1.修改定义指向
	for k, path := range sw.Paths {
		if path.Get == nil || path.Get.Responses == nil || path.Get.Responses["200"] == nil ||
			path.Get.Responses["200"].Schema == nil || path.Get.Responses["200"].Schema.Type != "array" {
			continue
		}
		if strings.Contains(path.Get.Responses["200"].Description, "pagination") {
			sc := path.Get.Responses["200"].Schema.Items.Title
			paginationDef := "pagination_" + sc
			paginationDefs = append(paginationDefs, paginationDef)
			sw.Paths[k].Get.Responses["200"].Schema.Title = paginationDef
			sw.Paths[k].Get.Responses["200"].Schema.Ref = "#/definitions/" + sc
			sw.Paths[k].Get.Responses["200"].Schema.Type = "object"
		}
	}
	// 2.添加定义
	for _, v := range paginationDefs {
		if _, ok := sw.Definitions[v]; !ok {
			sw.Definitions[v] = &swagger.Schema{
				Title: v,
				Type:  "object",
				Properties: map[string]*swagger.Propertie{
					"pagination": {
						Ref:  "#/definitions/HTTPPagination",
						Type: "object",
					},
					"list": {
						Type: "array",
						Items: &swagger.Propertie{
							Ref:  "#/definitions/" + strings.Replace(v, "pagination_", "", 1),
							Type: "object",
						},
					},
				},
			}
		}
	}

	var (
		data     []byte
		filename string
	)

	switch t {
	case "json":
		filename = jsonFile
		data, err = json.Marshal(sw)
	case "yaml":
		filename = yamlFile
		data, err = yaml.Marshal(sw)
	default:
		err = fmt.Errorf("missing swagger file type(%s), only support in (json, yaml)", t)
	}
	if err != nil {
		return
	}
	return ioutil.WriteFile(filepath.Join(output, filename), data, 0644)
}

func doc2Swagger(projectPath, swaggerGo string, dev bool, sw *swagger.Swagger) error {
	f, err := parser.ParseFile(token.NewFileSet(), swaggerGo, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	// Analyse API comments
	if f.Comments != nil {
		for _, c := range f.Comments {
			for _, s := range strings.Split(c.Text(), "\n") {
				switch {
				case tagTrimPrefixAndSpace(&s, appVersion):
					sw.Infos.Version = s
				case tagTrimPrefixAndSpace(&s, appTitle):
					sw.Infos.Title = s
				case tagTrimPrefixAndSpace(&s, appDesc):
					if sw.Infos.Description != "" {
						sw.Infos.Description += "<br>" + s
					} else {
						sw.Infos.Description = s
					}
				case tagTrimPrefixAndSpace(&s, appTermsOfServiceURL):
					sw.Infos.TermsOfService = s
				case tagTrimPrefixAndSpace(&s, appContact):
					sw.Infos.Contact.EMail = s
				case tagTrimPrefixAndSpace(&s, appName):
					sw.Infos.Contact.Name = s
				case tagTrimPrefixAndSpace(&s, appURL):
					sw.Infos.Contact.URL = s
				case tagTrimPrefixAndSpace(&s, appLicenseURL):
					sw.Infos.License.URL = s
				case tagTrimPrefixAndSpace(&s, appLicense):
					sw.Infos.License.Name = s
				case tagTrimPrefixAndSpace(&s, appSchemes):
					sw.Schemes = strings.Split(s, ",")
				case tagTrimPrefixAndSpace(&s, appHost):
					sw.Host = s
				case tagTrimPrefixAndSpace(&s, appBasePath):
					sw.BasePath = s
				case tagTrimPrefixAndSpace(&s, appConsumes):
					sw.Consumes = contentTypeByDoc(s)
				case tagTrimPrefixAndSpace(&s, appProduces):
					sw.Produces = contentTypeByDoc(s)
				}
			}
		}
	}

	// Analyse controller package
	// like:
	// swagger.go
	// import (
	//     _ "path/to/ctrl1"
	//     _ "path/to/ctrl2"
	//     _ "path/to/ctrl3"
	// )
	// // @APIVersion xxx
	// // @....
	for _, im := range f.Imports {
		importPath := strings.Trim(im.Path.Value, "\"")
		p, err := newResoucre(importPath, true)
		if err != nil {
			return err
		}
		if err = p.run(sw); err != nil {
			return err
		}
	}
	return nil
}
