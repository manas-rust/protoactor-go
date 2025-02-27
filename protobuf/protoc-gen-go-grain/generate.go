package main

import (
	"fmt"
	"strings"

	"github.com/asynkron/protoactor-go/protobuf/protoc-gen-go-grain/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const deprecationComment = "// Deprecated: Do not use."

const (
	timePackage    = protogen.GoImportPath("time")
	errorsPackage  = protogen.GoImportPath("errors")
	fmtPackage     = protogen.GoImportPath("fmt")
	slogPackage    = protogen.GoImportPath("log/slog")
	protoPackage   = protogen.GoImportPath("google.golang.org/protobuf/proto")
	actorPackage   = protogen.GoImportPath("github.com/asynkron/protoactor-go/actor")
	clusterPackage = protogen.GoImportPath("github.com/asynkron/protoactor-go/cluster")
)

var (
	noLowerCaser = cases.Title(language.AmericanEnglish, cases.NoLower)
	caser        = cases.Title(language.AmericanEnglish)
)

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	filename := file.GeneratedFilenamePrefix + "_grain.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	generateHeader(gen, g, file)
	generateContent(gen, g, file)
}

func generateHeader(gen *protogen.Plugin, g *protogen.GeneratedFile, file *protogen.File) {
	g.P("// Code generated by protoc-gen-grain. DO NOT EDIT.")
	g.P("// versions:")
	g.P("//  protoc-gen-grain ", version)
	protocVersion := "(unknown)"
	if v := gen.Request.GetCompilerVersion(); v != nil {
		protocVersion = fmt.Sprintf("v%v.%v.%v", v.GetMajor(), v.GetMinor(), v.GetPatch())
		if s := v.GetSuffix(); s != "" {
			protocVersion += "-" + s
		}
	}
	g.P("//  protoc           ", protocVersion)
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
}

func generateContent(gen *protogen.Plugin, g *protogen.GeneratedFile, file *protogen.File) {
	g.P("package ", file.GoPackageName)

	if len(file.Services) == 0 {
		return
	}

	g.QualifiedGoIdent(actorPackage.Ident(""))
	g.QualifiedGoIdent(clusterPackage.Ident(""))
	g.QualifiedGoIdent(protoPackage.Ident(""))
	g.QualifiedGoIdent(fmtPackage.Ident(""))
	g.QualifiedGoIdent(timePackage.Ident(""))
	g.QualifiedGoIdent(slogPackage.Ident(""))

	for _, enum := range file.Enums {
		if enum.Desc.Name() == "ErrorReason" {
			generateErrorReasons(g, enum)
		}
	}

	for _, service := range file.Services {
		generateService(service, file, g)
		g.P()
	}

	generateRespond(g)
}

func generateService(service *protogen.Service, file *protogen.File, g *protogen.GeneratedFile) {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}

	sd := &serviceDesc{
		Name: service.GoName,
	}

	for i, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}

		methodOptions, ok := proto.GetExtension(method.Desc.Options(), options.E_MethodOptions).(*options.MethodOptions)
		if !ok {
			continue
		}

		if methodOptions == nil {
			methodOptions = &options.MethodOptions{}
		}

		md := &methodDesc{
			Name:    method.GoName,
			Input:   g.QualifiedGoIdent(method.Input.GoIdent),
			Output:  g.QualifiedGoIdent(method.Output.GoIdent),
			Index:   i,
			Options: methodOptions,
		}

		sd.Methods = append(sd.Methods, md)
	}

	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func generateRespond(g *protogen.GeneratedFile) {
	g.P("func respond[T proto.Message](ctx cluster.GrainContext) func (T) {")
	g.P("return func (resp T) {")
	g.P("ctx.Respond(resp)")
	g.P("}")
	g.P("}")
}

func generateErrorReasons(g *protogen.GeneratedFile, enum *protogen.Enum) {
	var es errorsWrapper
	for _, v := range enum.Values {
		comment := v.Comments.Leading.String()
		if comment == "" {
			comment = v.Comments.Trailing.String()
		}

		err := &errorDesc{
			Name:       string(enum.Desc.Name()),
			Value:      string(v.Desc.Name()),
			CamelValue: toCamel(string(v.Desc.Name())),
			Comment:    comment,
			HasComment: len(comment) > 0,
		}
		es.Errors = append(es.Errors, err)
	}
	if len(es.Errors) != 0 {
		g.P(es.execute())
	}
}

func toCamel(s string) string {
	if !strings.Contains(s, "_") {
		if s == strings.ToUpper(s) {
			s = strings.ToLower(s)
		}
		return noLowerCaser.String(s)
	}

	slice := strings.Split(s, "_")
	for i := 0; i < len(slice); i++ {
		slice[i] = caser.String(slice[i])
	}
	return strings.Join(slice, "")
}
