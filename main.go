package main

import (
	"flag"
	myplugin "github.com/kirinse/protoc-gen-gorm/plugin"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	// flagSet initialization
	var flags flag.FlagSet
	// flag definitions
	quiet := flags.Bool("quiet", false, "Suppresses warnings if true.")
	stringEnums := flags.Bool("enums", false, "Use string representation of protobuf enums instead of integer value if true.")
	gateway := flags.Bool("gateway", false, "Generates gateway if true.")
	defaultHandlers := flags.Bool("defaultHandlers", false, "Generates defaultHandlers if true.")
	// protogen options, passing flagset callback.
	opts := &protogen.Options{
		ParamFunc: flags.Set,
	}
	// pass plugin callback to options run func. initialize
	// internal plugin & generate.
	opts.Run(func(p *protogen.Plugin) error {
		plugin := &myplugin.OrmPlugin{
			SuppressWarnings: *quiet,
			StringEnums:      *stringEnums,
			Gateway:          *gateway,
			DefaultHandlers:  *defaultHandlers,
		}
		plugin.Init(p)
		plugin.Generate()
		return nil
	})
}
