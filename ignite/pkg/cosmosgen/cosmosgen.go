package cosmosgen

import (
	"context"
	"path/filepath"

	gomodmodule "golang.org/x/mod/module"

	"github.com/ignite/cli/ignite/pkg/cache"
	"github.com/ignite/cli/ignite/pkg/cosmosanalysis/module"
)

// generateOptions used to configure code generation.
type generateOptions struct {
	includeDirs []string
	gomodPath   string

	jsOut               func(module.Module) string
	jsIncludeThirdParty bool
	tsClientRootPath    string

	vuexOut      func(module.Module) string
	vuexRootPath string

	specOut string

	dartOut               func(module.Module) string
	dartIncludeThirdParty bool
	dartRootPath          string
}

// TODO add WithInstall.

// ModulePathFunc defines a function type that returns a path based on a Cosmos SDK module.
type ModulePathFunc func(module.Module) string

// Option configures code generation.
type Option func(*generateOptions)

// WithTSClientGeneration adds Typescript Client code generation.
// The tsClientRootPath is used to determine the root path of generated Typescript classes.
func WithTSClientGeneration(out ModulePathFunc, tsClientRootPath string) Option {
	return func(o *generateOptions) {
		o.jsOut = out
		o.tsClientRootPath = tsClientRootPath
	}
}

func WithVuexGeneration(includeThirdPartyModules bool, out ModulePathFunc, vuexRootPath string) Option {
	return func(o *generateOptions) {
		o.vuexOut = out
		o.jsIncludeThirdParty = includeThirdPartyModules
		o.vuexRootPath = vuexRootPath
	}
}

func WithDartGeneration(includeThirdPartyModules bool, out ModulePathFunc, rootPath string) Option {
	return func(o *generateOptions) {
		o.dartOut = out
		o.dartIncludeThirdParty = includeThirdPartyModules
		o.dartRootPath = rootPath
	}
}

// WithGoGeneration adds Go code generation.
func WithGoGeneration(gomodPath string) Option {
	return func(o *generateOptions) {
		o.gomodPath = gomodPath
	}
}

// WithOpenAPIGeneration adds OpenAPI spec generation.
func WithOpenAPIGeneration(out string) Option {
	return func(o *generateOptions) {
		o.specOut = out
	}
}

// IncludeDirs configures the third party proto dirs that used by app's proto.
// relative to the projectPath.
func IncludeDirs(dirs []string) Option {
	return func(o *generateOptions) {
		o.includeDirs = dirs
	}
}

// generator generates code for sdk and sdk apps.
type generator struct {
	ctx          context.Context
	cacheStorage cache.Storage
	appPath      string
	protoDir     string
	o            *generateOptions
	sdkImport    string
	deps         []gomodmodule.Version
	appModules   []module.Module
	thirdModules map[string][]module.Module // app dependency-modules pair.
}

// Generate generates code from protoDir of an SDK app residing at appPath with given options.
// protoDir must be relative to the projectPath.
func Generate(ctx context.Context, cacheStorage cache.Storage, appPath, protoDir string, options ...Option) error {
	g := &generator{
		ctx:          ctx,
		appPath:      appPath,
		protoDir:     protoDir,
		o:            &generateOptions{},
		thirdModules: make(map[string][]module.Module),
		cacheStorage: cacheStorage,
	}

	for _, apply := range options {
		apply(g.o)
	}

	if err := g.setup(); err != nil {
		return err
	}

	// Go generation must run first so the types are created before other
	// generated code that requires sdk.Msg implementations to be defined
	if g.o.gomodPath != "" {
		if err := g.generateGo(); err != nil {
			return err
		}
	}

	if g.o.jsOut != nil {
		if err := g.generateTS(); err != nil {
			return err
		}
	}

	if g.o.vuexOut != nil {
		if err := g.generateVuex(); err != nil {
			return err
		}

		// Update Vue app dependeciens when Vuex stores are generated.
		// This update is required to link the "ts-client" folder so the
		// package is available during development before publishing it.
		if err := g.updateVueDependencies(); err != nil {
			return err
		}
	}

	if g.o.dartOut != nil {
		if err := g.generateDart(); err != nil {
			return err
		}
	}

	if g.o.specOut != "" {
		if err := generateOpenAPISpec(g); err != nil {
			return err
		}
	}

	return nil
}

// TypescriptModulePath generates module paths for Cosmos SDK modules.
// The root path is used as prefix for the generated paths.
func TypescriptModulePath(rootPath string) ModulePathFunc {
	return func(m module.Module) string {
		return filepath.Join(rootPath, m.Pkg.Name)
	}
}
