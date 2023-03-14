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
