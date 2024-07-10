package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run imply.go <file_name> [interface_name] [struct_name] [package_name]")
		os.Exit(1)
	}

	fileName := os.Args[1]
	interfaceName := ""
	structName := ""
	packageName := ""

	if len(os.Args) > 2 {
		interfaceName = os.Args[2]
	}
	if len(os.Args) > 3 {
		structName = os.Args[3]
	}
	if len(os.Args) > 4 {
		packageName = os.Args[4]
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// If package name is not provided, use the original package name
	if packageName == "" {
		packageName = node.Name.Name
	}

	// Find the interface and collect imports
	var interfaceType *ast.InterfaceType
	imports := make(map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.ImportSpec:
			if t.Name != nil {
				imports[t.Name.Name] = strings.Trim(t.Path.Value, "\"")
			} else {
				parts := strings.Split(strings.Trim(t.Path.Value, "\""), "/")
				imports[parts[len(parts)-1]] = strings.Trim(t.Path.Value, "\"")
			}
		case *ast.TypeSpec:
			if i, ok := t.Type.(*ast.InterfaceType); ok {
				if interfaceName == "" || t.Name.Name == interfaceName {
					interfaceType = i
					if structName == "" {
						structName = "Impl" + t.Name.Name
					}
					return false
				}
			}
		}
		return true
	})
	if interfaceType == nil {
		fmt.Println("Interface not found")
		os.Exit(1)
	}

	// Generate the package and imports
	fmt.Printf("package %s\n\n", packageName)
	if len(imports) > 0 {
		fmt.Println("import (")
		for alias, path := range imports {
			if alias == path {
				fmt.Printf("\t\"%s\"\n", path)
			} else {
				fmt.Printf("\t%s \"%s\"\n", alias, path)
			}
		}
		fmt.Println(")")
		fmt.Println()
	}

	// Generate the struct and methods
	fmt.Printf("type %s struct{}\n\n", structName)

	for _, method := range interfaceType.Methods.List {
		funcType := method.Type.(*ast.FuncType)
		funcName := method.Names[0].Name

		params := generateParams(funcType.Params)
		returns := generateReturns(funcType.Results)

		fmt.Printf("func (i *%s) %s(%s) %s {\n", structName, funcName, params, returns)
		fmt.Print(generateReturnStatement(funcType.Results))
		fmt.Print("}\n\n")
	}

}

func generateParams(fields *ast.FieldList) string {
	var params []string
	for _, field := range fields.List {
		paramType := ""
		if len(field.Names) > 0 {
			paramType = field.Names[0].Name + " "
		}
		paramType += getTypeString(field.Type)
		params = append(params, paramType)
	}
	return strings.Join(params, ", ")
}

func generateReturns(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	var returns []string
	for _, field := range fields.List {
		returns = append(returns, getTypeString(field.Type))
	}
	if len(returns) == 1 {
		return returns[0]
	}
	return "(" + strings.Join(returns, ", ") + ")"
}

func generateReturnStatement(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	var returns []string
	for _, field := range fields.List {
		returns = append(returns, getZeroValue(field.Type))
	}
	return "\treturn " + strings.Join(returns, ", ") + "\n"
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", getTypeString(t.Len), getTypeString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeString(t.Key), getTypeString(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + getTypeString(t.Elt)
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func getZeroValue(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "\"\""
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "byte", "rune":
			return "0"
		case "float32", "float64":
			return "0.0"
		case "bool":
			return "false"
		case "error":
			return "nil"
		default:
			return t.Name + "{}"
		}
	case *ast.StarExpr:
		return "nil"
	case *ast.ArrayType:
		return "nil"
	case *ast.MapType:
		return "nil"
	case *ast.InterfaceType:
		return "nil"
	case *ast.SelectorExpr:
		if t.Sel.Name == "Error" {
			return "nil"
		}
		return getTypeString(t) + "{}"
	default:
		return "nil"
	}
}
