package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// StructInfo holds information about a struct that needs a Reset method
type StructInfo struct {
	Name     string
	Doc      string
	Fields   []*FieldInfo
	Package  string
	FilePath string
}

type FieldInfo struct {
	Name string
	Type string
}

// ResetGenerator generates Reset() methods for structs marked with // generate:reset
type ResetGenerator struct {
	fset          *token.FileSet
	processedPkgs map[string]bool
}

func main() {
	// Get root directory (current directory)
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	gen := &ResetGenerator{
		fset:          token.NewFileSet(),
		processedPkgs: make(map[string]bool),
	}

	// Walk through all packages
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that are not Go packages
		if !info.IsDir() {
			return nil
		}

		// Skip vendor, testdata, etc.
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "vendor" || base == "testdata" || base == "cmd" || base == "profiles" {
			return nil
		}

		// Process the package
		gen.processPackage(path)

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	log.Println("Reset method generation complete")
}

func (g *ResetGenerator) processPackage(pkgPath string) {
	// Check if we've already processed this package
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return
	}
	if g.processedPkgs[absPath] {
		return
	}
	g.processedPkgs[absPath] = true

	// Parse all Go files in the package
	files, err := parser.ParseDir(g.fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing package %s: %v", pkgPath, err)
		return
	}

	// Find structs with // generate:reset comment
	var structsToGenerate []StructInfo

	for _, pkg := range files {
		for fileName, file := range pkg.Files {
			// Build a map of comment groups by their position
			commentMap := ast.NewCommentMap(g.fset, file, file.Comments)

			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					// Check if struct has // generate:reset comment
					hasResetComment := g.hasResetComment(genDecl, typeSpec, commentMap)

					if hasResetComment {
						structInfo := g.extractStructInfo(typeSpec, file)
						if structInfo.Name != "" {
							structInfo.Package = pkgPath
							structInfo.FilePath = fileName
							structsToGenerate = append(structsToGenerate, structInfo)
						}
					}
				}
			}
		}
	}

	// Generate Reset methods if we found any structs
	if len(structsToGenerate) > 0 {
		g.generateResetFile(pkgPath, structsToGenerate)
	}
}

func (g *ResetGenerator) hasResetComment(genDecl *ast.GenDecl, typeSpec *ast.TypeSpec, commentMap ast.CommentMap) bool {
	// Check genDecl.Doc
	if genDecl.Doc != nil {
		for _, comment := range genDecl.Doc.List {
			if strings.TrimSpace(comment.Text) == "generate:reset" {
				return true
			}
		}
	}

	// Check typeSpec.Doc (inline comment on the type spec)
	if typeSpec.Doc != nil {
		for _, comment := range typeSpec.Doc.List {
			if strings.TrimSpace(comment.Text) == "generate:reset" {
				return true
			}
		}
	}

	// Check comments in the comment map
	if comments, ok := commentMap[typeSpec]; ok {
		for _, group := range comments {
			for _, comment := range group.List {
				if strings.TrimSpace(comment.Text) == "generate:reset" {
					return true
				}
			}
		}
	}

	return false
}

func (g *ResetGenerator) extractStructInfo(typeSpec *ast.TypeSpec, file *ast.File) StructInfo {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return StructInfo{}
	}

	info := StructInfo{
		Name: typeSpec.Name.Name,
	}

	if typeSpec.Doc != nil {
		info.Doc = typeSpec.Doc.Text()
	}

	// Collect fields
	for _, field := range structType.Fields.List {
		if field.Names == nil {
			continue
		}

		for _, name := range field.Names {
			if name.IsExported() {
				fieldType := g.typeToString(field.Type)
				info.Fields = append(info.Fields, &FieldInfo{
					Name: name.Name,
					Type: fieldType,
				})
			}
		}
	}

	return info
}

func (g *ResetGenerator) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + g.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + g.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + g.typeToString(t.Key) + "]" + g.typeToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.SelectorExpr:
		ident, ok := t.X.(*ast.Ident)
		if ok {
			return ident.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	default:
		return "interface{}"
	}
}

func (g *ResetGenerator) generateResetFile(pkgPath string, structs []StructInfo) {
	var buf bytes.Buffer

	buf.WriteString("// Code generated by cmd/reset. DO NOT EDIT.\n\n")
	buf.WriteString("package ")
	buf.WriteString(filepath.Base(pkgPath))
	buf.WriteString("\n\n")

	for _, structInfo := range structs {
		g.generateResetMethod(&buf, structInfo)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Error formatting generated code for %s: %v", pkgPath, err)
		// Write unformatted code
		formatted = buf.Bytes()
	}

	// Write to reset.gen.go
	outputPath := filepath.Join(pkgPath, "reset.gen.go")
	err = os.WriteFile(outputPath, formatted, 0644)
	if err != nil {
		log.Printf("Error writing generated file %s: %v", outputPath, err)
		return
	}

	log.Printf("Generated Reset methods for %d struct(s) in %s", len(structs), outputPath)
}

func (g *ResetGenerator) generateResetMethod(buf *bytes.Buffer, structInfo StructInfo) {
	// Write doc comment
	if structInfo.Doc != "" {
		buf.WriteString("// ")
		buf.WriteString(structInfo.Name)
		buf.WriteString(" resets the struct to its zero value.\n")
	}
	buf.WriteString("func (rs *")
	buf.WriteString(structInfo.Name)
	buf.WriteString(") Reset() {\n")
	buf.WriteString("\tif rs == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n\n")

	for _, field := range structInfo.Fields {
		g.generateFieldReset(buf, field)
	}

	buf.WriteString("}\n\n")
}

func (g *ResetGenerator) generateFieldReset(buf *bytes.Buffer, field *FieldInfo) {
	fieldName := field.Name
	fieldType := field.Type

	// Check if it's a primitive type
	switch fieldType {
	case "int", "int8", "int16", "int32", "int64":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = 0\n")
	case "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = 0\n")
	case "float32", "float64":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = 0\n")
	case "string":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = \"\"\n")
	case "bool":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = false\n")
	case "byte", "rune", "complex64", "complex128":
		buf.WriteString("\trs.")
		buf.WriteString(fieldName)
		buf.WriteString(" = 0\n")
	default:
		// Handle complex types
		if strings.HasPrefix(fieldType, "[]") {
			// Slice - truncate to [:0]
			buf.WriteString("\trs.")
			buf.WriteString(fieldName)
			buf.WriteString(" = rs.")
			buf.WriteString(fieldName)
			buf.WriteString("[:0]\n")
		} else if strings.HasPrefix(fieldType, "map[") {
			// Map - use clear()
			buf.WriteString("\tclear(rs.")
			buf.WriteString(fieldName)
			buf.WriteString(")\n")
		} else if strings.HasPrefix(fieldType, "*") {
			// Pointer
			elemType := strings.TrimPrefix(fieldType, "*")
			// Check if it's a pointer to primitive
			if isPrimitiveType(elemType) {
				buf.WriteString("\tif rs.")
				buf.WriteString(fieldName)
				buf.WriteString(" != nil {\n")
				buf.WriteString("\t\t*rs.")
				buf.WriteString(fieldName)
				buf.WriteString(" = ")
				buf.WriteString(getZeroValue(elemType))
				buf.WriteString("\n")
				buf.WriteString("\t}\n")
			} else {
				// Pointer to struct - check if it has Reset method
				buf.WriteString("\tif rs.")
				buf.WriteString(fieldName)
				buf.WriteString(" != nil {\n")
				buf.WriteString("\t\tif resetter, ok := interface{}(rs.")
				buf.WriteString(fieldName)
				buf.WriteString(").(interface{ Reset() }); ok {\n")
				buf.WriteString("\t\t\tresetter.Reset()\n")
				buf.WriteString("\t\t}\n")
				buf.WriteString("\t}\n")
			}
		} else {
			// Regular struct - check if it has Reset method
			buf.WriteString("\tif resetter, ok := interface{}(&rs.")
			buf.WriteString(fieldName)
			buf.WriteString(").(interface{ Reset() }); ok {\n")
			buf.WriteString("\t\tresetter.Reset()\n")
			buf.WriteString("\t}\n")
		}
	}
}

func isPrimitiveType(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"string", "bool",
		"byte", "rune", "complex64", "complex128":
		return true
	default:
		return false
	}
}

func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"byte", "rune", "complex64", "complex128":
		return "0"
	case "string":
		return "\"\""
	case "bool":
		return "false"
	default:
		return "zero value"
	}
}
