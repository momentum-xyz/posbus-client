package main

// ESBuild - https://esbuild.github.io

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/gzuidhof/tygo/tygo"
	"github.com/pkg/errors"
)

var buildOptions = api.BuildOptions{
	EntryPoints: []string{"ts/index.ts", "ts/worker.ts"},
	Bundle:      true,
	Outdir:      "dist",
	Format:      api.FormatESModule,
	OutExtension: map[string]string{
		".js": ".mjs",
	},
	AssetNames: "[name]",
	Loader: map[string]api.Loader{
		".wasm": api.LoaderFile,
	},
	External:          []string{"pbc.wasm"},
	MinifyWhitespace:  true,
	MinifyIdentifiers: true,
	MinifySyntax:      true,
	Target:            api.ES2022,
	Engines: []api.Engine{
		{Name: api.EngineChrome, Version: "100"},
		{Name: api.EngineFirefox, Version: "100"},
		{Name: api.EngineSafari, Version: "16"},
	},
	Write:     true,
	Metafile:  true,
	Sourcemap: api.SourceMapLinked,
}

var serveOptions = api.ServeOptions{
	Host:     "localhost",
	Servedir: "dist",
}

// TODO: go:embed this?
// A number of types are outside posbus/types.go.
// Avoids scanning whole project and using which list of irrelevant types that should not be used.
var extraTypes = `
export type byte = number; // TODO: single use, as bitmask
export interface Vec3 {
  x: number;
  y: number;
  z: number;
}
export interface ObjectTransform {
  position: Vec3;
  rotation: Vec3;
  scale: Vec3;
}
export interface UserTransform {
  location: Vec3;
  rotation: Vec3;
}
export type SlotType = "" | "texture" | "string" | "number";

export interface SetUserTransform {
    id: string;
    transform: UserTransform
}
export interface SetUsersTransforms {
    value: SetUserTransform[];
}
`

func main() {
	var (
		help  bool = false
		serve bool = false
		port  int  = 0
	)
	flag.BoolVar(&help, "h", false, "Help me!")
	flag.BoolVar(&serve, "s", false, "Serve build with http")
	flag.IntVar(&port, "p", 0, "Port to serve on")
	flag.Parse()
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := generateTypes(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if serve {
		if err := runServer(ctx, port, buildOptions); err != nil {
			log.Fatal(err)
		}
	} else {
		result, err := build(ctx, buildOptions)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", api.AnalyzeMetafile(result.Metafile, api.AnalyzeMetafileOptions{Verbose: true}))
	}

	//ioutil.WriteFile("dist/meta.json", []byte(result.Metafile), 0644)
}

func build(ctx context.Context, buildOptions api.BuildOptions) (*api.BuildResult, error) {
	buildCtx, err := api.Context(buildOptions)
	if err != nil {
		return nil, errors.Wrap(err, "build context")
	}
	result := buildCtx.Rebuild()
	return &result, nil
}

func runServer(ctx context.Context, port int, buildOptions api.BuildOptions) error {
	buildOptions.Write = true
	buildOptions.Outdir = "dist/js"
	buildCtx, err := api.Context(buildOptions)
	if err != nil {
		return errors.Wrap(err, "build context")
	}
	serveOptions.Port = uint16(port)
	sr, srvErr := buildCtx.Serve(serveOptions)
	if srvErr != nil {
		return errors.Wrap(srvErr, "serve")
	}
	fmt.Printf("Serving: http://%s:%v\n", sr.Host, sr.Port)
	<-ctx.Done()
	return nil
}

func generateTypes(ctx context.Context) error {
	config := &tygo.Config{
		Packages: []*tygo.PackageConfig{
			&tygo.PackageConfig{
				Path:         "github.com/momentum-xyz/ubercontroller/pkg/posbus",
				OutputPath:   "build/posbus.d.ts",
				IncludeFiles: []string{"types.go"},
				Frontmatter:  extraTypes,
				TypeMappings: map[string]string{
					"umid.UMID":             "string",
					"dto.Asset3dType":       "string",
					"cmath.ObjectTransform": "ObjectTransform",
					"cmath.UserTransform":   "UserTransform",
					"entry.UnitySlotType":   "SlotType",
				},
				FallbackType: "any",
			},
		},
	}
	gen := tygo.New(config)
	err := gen.Generate()
	return err
}
