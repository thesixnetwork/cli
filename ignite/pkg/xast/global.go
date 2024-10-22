package xast

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/ignite/cli/v29/ignite/pkg/errors"
)

type (
	// globalOpts represent the options for globals.
	globalOpts struct {
		globals []global
	}

	// GlobalOptions configures code generation.
	GlobalOptions func(*globalOpts)

	global struct {
		name, varType, value string
	}

	// GlobalType represents the global type.
	GlobalType string
)

const (
	GlobalTypeVar   GlobalType = "var"
	GlobalTypeConst GlobalType = "const"
)

// WithGlobal add a new global.
func WithGlobal(name, varType, value string) GlobalOptions {
	return func(c *globalOpts) {
		c.globals = append(c.globals, global{
			name:    name,
			varType: varType,
			value:   value,
		})
	}
}

func newGlobalOptions() globalOpts {
	return globalOpts{
		globals: make([]global, 0),
	}
}

// InsertGlobal inserts global variables or constants into the provided Go source code content after the import section.
// The function parses the provided content, locates the import section, and inserts the global declarations immediately after it.
// The type of globals (variables or constants) is specified by the globalType parameter.
// Each global declaration is defined by calling WithGlobal function with appropriate arguments.
// The function returns the modified content with the inserted global declarations.
func InsertGlobal(fileContent string, globalType GlobalType, globals ...GlobalOptions) (modifiedContent string, err error) {
	// apply global options.
	opts := newGlobalOptions()
	for _, o := range globals {
		o(&opts)
	}

	fileSet := token.NewFileSet()

	// Parse the Go source code content.
	f, err := parser.ParseFile(fileSet, "", fileContent, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Find the index of the import declaration or package declaration if no imports.
	var insertIndex int
	for i, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			insertIndex = i + 1
			break
		} else if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			insertIndex = i
			if funcDecl.Doc == nil {
				insertIndex++
			}
			break
		}
	}

	// Create global variable/constant declarations.
	for _, global := range opts.globals {
		// Create an identifier for the global.
		ident := ast.NewIdent(global.name)

		// Create a value expression if provided.
		var valueExpr ast.Expr
		if global.value != "" {
			valueExpr, err = parser.ParseExprFrom(fileSet, "", []byte(global.value), parser.ParseComments)
			if err != nil {
				return "", err
			}
		}

		// Create a declaration based on the global type.
		var spec ast.Spec
		switch globalType {
		case GlobalTypeVar:
			spec = &ast.ValueSpec{
				Names:  []*ast.Ident{ident},
				Type:   ast.NewIdent(global.varType),
				Values: []ast.Expr{valueExpr},
			}
		case GlobalTypeConst:
			spec = &ast.ValueSpec{
				Names:  []*ast.Ident{ident},
				Type:   ast.NewIdent(global.varType),
				Values: []ast.Expr{valueExpr},
			}
		default:
			return "", errors.Errorf("unsupported global type: %s", string(globalType))
		}

		// Insert the declaration after the import section or package declaration if no imports.
		f.Decls = append(
			f.Decls[:insertIndex],
			append([]ast.Decl{
				&ast.GenDecl{
					TokPos: 1,
					Tok:    token.Lookup(string(globalType)),
					Specs:  []ast.Spec{spec},
				},
			}, f.Decls[insertIndex:]...)...)
		insertIndex++
	}

	// Format the modified AST.
	var buf bytes.Buffer
	if err := format.Node(&buf, fileSet, f); err != nil {
		return "", err
	}

	// Return the modified content.
	return buf.String(), nil
}

// AppendFunction appends a new function to the end of the Go source code content.
func AppendFunction(fileContent string, function string) (modifiedContent string, err error) {
	fileSet := token.NewFileSet()

	// Parse the function body as a separate file.
	funcFile, err := parser.ParseFile(fileSet, "", "package main\n"+function, parser.AllErrors)
	if err != nil {
		return "", err
	}

	// Extract the first declaration, assuming it's a function declaration.
	var funcDecl *ast.FuncDecl
	for _, decl := range funcFile.Decls {
		if fDecl, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fDecl
			break
		}
	}
	if funcDecl == nil {
		return "", errors.Errorf("no function declaration found in the provided function body")
	}

	// Parse the Go source code content.
	f, err := parser.ParseFile(fileSet, "", fileContent, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Append the function declaration to the file's declarations.
	f.Decls = append(f.Decls, funcDecl)

	// Format the modified AST.
	var buf bytes.Buffer
	if err := format.Node(&buf, fileSet, f); err != nil {
		return "", err
	}

	return buf.String(), nil
}
