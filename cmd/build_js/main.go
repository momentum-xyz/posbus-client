package main

// ESBuild - https://esbuild.github.io

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/momentum-xyz/ubercontroller/pkg/posbus"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/gzuidhof/tygo/tygo"
	"github.com/pkg/errors"
)

var buildOptions = api.BuildOptions{
	LogLevel:    api.LogLevelInfo,
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
export interface Transform {
  location: Vec3;
  rotation: Vec3;
}
export interface ObjectTransform extends Transform {
  scale: Vec3;
}

export type SlotType = "" | "texture" | "string" | "number";

// source: object_data.go replaced
export interface ObjectData {
    id: string,
    entries: { SlotType: {[key: string]: any} }
}

// source: message.go replaced
export type MsgType = number;

// source: users_transform_list.go patch
const MsgUUIDTypeSize = 16;


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

	err := generateConstants(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated constants")

	err = generateChannelTypes(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated channel types")

	err = generateTypes(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated types")

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
	if len(result.Errors) > 0 {
		log.Fatal("Build failed.")
	}
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
				Path:       "github.com/momentum-xyz/ubercontroller/pkg/posbus",
				OutputPath: "build/posbus.ts",
				//IncludeFiles: []string{"types.autogen.go"},
				Frontmatter: extraTypes,
				TypeMappings: map[string]string{
					"umid.UMID":             "string",
					"dto.Asset3dType":       "string",
					"cmath.ObjectTransform": "ObjectTransform",
					"cmath.UserTransform":   "Transform",
					"cmath.Float32Bytes":    "4",
					"entry.UnitySlotType":   "SlotType",
				},
				ExcludeFiles: []string{
					"message.go",
					"object_data.go", // TODO: fix ObjectData.Entries type
				},
				FallbackType: "any",
			},
		},
	}
	gen := tygo.New(config)
	err := gen.Generate()
	return err
}

func generateConstants(ctx context.Context) error {

	f, err := os.Create("build/constants.ts")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	_, err = fmt.Fprintf(w, "export enum MsgType {\n")
	check_error(err)

	for _, msgId := range posbus.GetMessageIds() {
		msgName := posbus.MessageNameById(msgId)
		_, err = fmt.Fprintf(w, "  %+v = \"%s\",\n", strings.ToUpper(msgName), msgName)
		check_error(err)
	}
	_, err = fmt.Fprintf(w, "}")
	check_error(err)
	w.Flush()

	return nil
}

func generateChannelTypes(ctx context.Context) error {

	f2, err := os.Create("build/channel_types.ts")
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	w2 := bufio.NewWriter(f2)

	_, err = fmt.Fprintf(
		w2, "import type * as posbus from \"./posbus\";\nimport type { MsgType } from \"./constants\";\n",
	)
	check_error(err)

	var unionType = "export type PosbusMessage ="
	for _, msgId := range posbus.GetMessageIds() {
		msgName := posbus.MessageNameById(msgId)
		typeName := posbus.MessageTypeNameById(msgId)
		t := typeName + "Type"
		_, err = fmt.Fprintf(
			w2, "export type %s = [MsgType.%s, posbus.%s];\n", t, strings.ToUpper(msgName), typeName,
		)
		check_error(err)
		unionType += fmt.Sprintf("\n  | %s", t)
	}

	unionType += ";\n"
	_, err = fmt.Fprintf(w2, "\n\n%s\n", unionType)
	check_error(err)
	w2.Flush()

	return nil
}

func check_error(err error) {
	if err != nil {
		panic(err)
	}
}
