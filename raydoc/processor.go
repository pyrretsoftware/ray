package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"maps"
	"os"
	"strings"
)

func getType(ex ast.Expr) string {
	switch t := ex.(type) {
		case *ast.Ident:
			return t.Name
		case *ast.StarExpr:
			return getType(t.X)

		case *ast.ArrayType:
			return "[]" + getType(t.Elt)

		case *ast.SelectorExpr:
			return "unknown"

		case *ast.MapType:
			return "map[" + getType(t.Key) + "]" + getType(t.Value) 
	}
	return ""
}

type Field struct {
	Type string
	Comment string
	Doc string
}

type RawStruct struct {
	Doc string
	StructType *ast.StructType
}

type ProcessedStruct struct {
	Doc string
	Fields map[string]Field
}

func TypeToJsonExample(tp string) string {
	if strings.HasPrefix(tp, "[]") {
		after, _ := strings.CutPrefix(tp, "[]")
		return "[" + TypeToJsonExample(after) + "]"
	} else if strings.HasPrefix(tp, "map") {
		after, _ := strings.CutPrefix(tp, "map[string]")
		return `{"": ` + TypeToJsonExample(after) + `}`
	} else if tp == "string" {
		return "\"\""
	} else if tp == "float64" || tp == "float32" {
		return "0.00"
	} else if tp == "int" || tp == "int32" || tp == "int64" {
		return "0"
	} else if tp == "bool" {
		return "false"
	} else {
		return tp
	}
}

func CutTypePrefix(tp string) string {
	if strings.HasPrefix(tp, "[]") {
		after, _ := strings.CutPrefix(tp, "[]")
		return CutTypePrefix(after)
	} else if strings.HasPrefix(tp, "map") {
		after, _ := strings.CutPrefix(tp, "map[string]")
		return CutTypePrefix(after)
	}
	return tp
}

func FormatJson(input string) string {
	var buf bytes.Buffer
	json.Indent(&buf, []byte(input), "  ", "  ")
	
	return buf.String()
}

func UltraProcessStructs(target ProcessedStruct, all map[string]ProcessedStruct) map[string]ProcessedStruct {
	result := map[string]ProcessedStruct{}
	for _, val := range target.Fields {
		if fStruct, ok := all[CutTypePrefix(val.Type)]; ok {
			maps.Copy(result, UltraProcessStructs(fStruct, all))
			result[CutTypePrefix(val.Type)] = fStruct
		}
	}

	return result
}


func ProcessJob(outName string, path string, Base string) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	structs := map[string]RawStruct{}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			tSpec := spec.(*ast.TypeSpec)
			structType, ok := tSpec.Type.(*ast.StructType)
			if ok {
				structs[tSpec.Name.Name] = RawStruct{
					Doc:        genDecl.Doc.Text(),
					StructType: structType,
				}
			}
		}
	}

	processedStructs := map[string]ProcessedStruct{}
	baseStruct := ProcessedStruct{}
	for stName, st := range structs {
		processedStructs[stName] = ProcessedStruct{
			Doc:    st.Doc,
			Fields: map[string]Field{},
		}

		for _, field := range st.StructType.Fields.List {
			doc := ""
			if field.Doc != nil {
				doc = field.Doc.Text()
			}
			comment := ""
			if field.Comment != nil {
				comment = field.Comment.Text()
			}

			resolved := getType(field.Type)
			processedStructs[stName].Fields[field.Names[0].Name] = Field{
				Type:    resolved,
				Doc:     doc,
				Comment: comment,
			}
		}

		if stName == Base {
			baseStruct = processedStructs[stName]
		}
	}

	ultraProcessedStructs := UltraProcessStructs(baseStruct, processedStructs)
	ultraProcessedStructs[Base] = baseStruct

	output := strings.ReplaceAll(DocsIntroduction, "%%BASE%%", Base)
	for name, processedStruct := range ultraProcessedStructs {
		if strings.Contains(processedStruct.Doc, ";") {
			name = strings.Split(processedStruct.Doc, ";")[0] + " (" + name + ")"
			processedStruct.Doc = strings.Split(processedStruct.Doc, ";")[1]
		}

		output += IndentHeadings + "# " + name + "\n"
		output += processedStruct.Doc + "\n"
		output += "::: details View JSON documentation\n"
		output += "```json\n"
		output += "// Note that this is not a proper example, but just shows docs in JSON form\n"
		output += "{\n"

		i := 0
		for fieldName, field := range processedStruct.Fields {
			i++
			output += "  \"" + fieldName + "\": " + FormatJson(TypeToJsonExample(field.Type))
			if i < len(processedStruct.Fields) {
				output += ","
			}
			output += " //" + strings.ReplaceAll(field.Doc, "\n", "") + "\n"
		}
		output += "}\n"
		output += "```\n"
		output += ":::\n"

		for fieldName, field := range processedStruct.Fields {
			typeLink := " ``" + field.Type + "``"
			foundType, typeFound := ultraProcessedStructs[CutTypePrefix(field.Type)]
			if typeFound {
				if strings.Contains(foundType.Doc, ";") {
					typeLink = strings.ToLower(strings.ReplaceAll(strings.Split(foundType.Doc, ";")[0], " ", "-")) + "-" + strings.ToLower(CutTypePrefix(field.Type))
				} else {
					typeLink = CutTypePrefix(field.Type)
				}
				typeLink = " [``" + field.Type + "``](#" + typeLink + ")"
			}
			output += IndentHeadings + "### " + fieldName + typeLink + "\n"
			if strings.Contains(field.Comment, "!DEP") {
				output += "::: warning\nThis property is deprecated, it's highly recommended to avoid it.\n:::\n"
			}

			output += field.Doc
			output += "\n"
		}
	}

	err = os.WriteFile("./out/" + outName + ".md", []byte(output), 0600)
	if err != nil {
		log.Fatal(err)
	}
}