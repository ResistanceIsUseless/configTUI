package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fileName := "nuclei.go"
	if err := parseGoFile2(fileName); err != nil {
		fmt.Println("Error parsing file:", err)
	}
}
func printFlagDetails(callExpr *ast.CallExpr) {
	if len(callExpr.Args) > 1 {
		flagName := exprToString(callExpr.Args[1]) // Adjust index based on where the flag name is
		fmt.Println("Nested Flag Name:", flagName, "Value:", exprToString(callExpr.Args[3]))
		// Add more details as needed
	}
}
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.Ident:
		return e.Name
	// Add more cases here as needed for different types of expressions
	default:
		return fmt.Sprintf("%v", expr)
	}
}
func parseGoFile2(fileName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Check if it is a function declaration
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "readConfig" {
			for _, stmt := range fn.Body.List {
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}
				callExpr, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}
				// Check for CreateGroup method call
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "flagSet" && selExpr.Sel.Name == "CreateGroup" {
						for _, arg := range callExpr.Args[2:] { // Skipping first two arguments (group name and description)
							if nestedCallExpr, ok := arg.(*ast.CallExpr); ok {
								printFlagDetails(nestedCallExpr)
							}
						}
					}
				}
			}
		}
		return true
	})

	return nil
}
func parseGoFile(fileName string) error {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file given in arguments
	node, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Inspect the AST and find our function
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok && fn.Name.Name == "readFlags" {
			for _, stmt := range fn.Body.List {
				// Check for ExprStmt to find expression statements
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}

				// Check for CallExpr to find function calls
				callExpr, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}

				// Check if the call is to flagSet
				if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "flagSet" {
						if selExpr.Sel.Name == "CreateGroup" {
							fmt.Println("FlagSet Group:", exprToString(callExpr.Args[0])) // Group name
							for _, arg := range callExpr.Args[3:] {                       // Assuming the first three arguments are not flags
								if nestedCallExpr, ok := arg.(*ast.CallExpr); ok {
									printFlagDetails(nestedCallExpr)
								}
							}
						}
					}
				}

			}
		}
		return true
	})

	return nil
}
