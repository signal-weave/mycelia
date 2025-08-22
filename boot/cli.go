package boot

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mycelia/commands"
	"mycelia/errgo"
)

// ------Arg Parsing------------------------------------------------------------

type runtimeConfig struct {
	Address          string
	Port             int
	Verbosity        int // 0=quiet, 1=info, 2=debug, 3=trace...
	PrintTree        bool
	TransformTimeout time.Duration
}

func defaultRuntimeConfig() runtimeConfig {
	return runtimeConfig{
		Address:          "127.0.0.1",
		Port:             5000,
		Verbosity:        0,
		PrintTree:        false,
		TransformTimeout: 5 * time.Second,
	}
}

// -----------------------------------------------------------------------------
// The primary struct for getting cli values. They are stored here rather than
// environment vars so that they can be stored as non-string types like nubmers
// or time durations.
// -----------------------------------------------------------------------------
var RuntimeCfg = defaultRuntimeConfig()

// Parses and stores the runtime flags in public var.
func ParseRuntimeArgs(argv []string) error {
	cfg, err := parseRuntimeArgs(argv)
	if err != nil {
		return err
	}
	RuntimeCfg = cfg

	_, err = os.Stat(preInitFile)
	if err == nil {
		getPreInitData(&cfg)
	}

	return nil
}

// ParseRuntimeArgs parses only runtime flags validates, and returns
// (config, error).
//
// Duration examples: 500ms, 3s, 2m, 1h.
func parseRuntimeArgs(argv []string) (runtimeConfig, error) {
	cfg := defaultRuntimeConfig()

	fs := flag.NewFlagSet("runtime", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	fs.StringVar(&cfg.Address, "address", cfg.Address, "Bind address (IP or hostname)")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "Bind port (1-65535)")
	fs.BoolVar(&cfg.PrintTree, "print-tree", cfg.PrintTree, "Print router tree at startup")
	fs.DurationVar(&cfg.TransformTimeout, "xform-timeout", cfg.TransformTimeout, "Transformer timeout (e.g. 30s, 2m)")
	fs.IntVar(&cfg.Verbosity, "verbosity", cfg.Verbosity,
		`0 - None
    1 - Errors
    2 - Warnings + Errors
    3 - Errors + Warnings + Actions`)

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), `Mycelia runtime options:

  -address string         Bind address (IP or hostname)
  -port int            Bind port (1-65535)
  -v int               0, 1, 2, or 3
  -print-tree          Print router tree at startup
  -xform-timeout dur   Transformer timeout

Examples:
  mycelia -addr 0.0.0.0 -port 8080 -verbosity 2 -print-tree -xform-timeout 45s
  MYC_ADDR=0.0.0.0 MYC_PORT=8080 mycelia -v
`)
	}

	if err := fs.Parse(argv); err != nil {
		return cfg, err
	}

	if err := validateRuntimeConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func validateRuntimeConfig(c runtimeConfig) error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port %d (expected 1-65535)", c.Port)
	}
	// Allow hostnames; validate if it looks like an IP.
	if ip := net.ParseIP(c.Address); ip == nil && looksLikeIP(c.Address) {
		return fmt.Errorf("invalid IP address %q", c.Address)
	}
	if c.TransformTimeout <= 0 {
		return errors.New("xform-timeout must be > 0")
	}
	return nil
}

func looksLikeIP(s string) bool {
	// crude: "n.n.n.n" suggests intended IP; otherwise treat as hostname
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return true
		}
		if _, err := strconv.Atoi(p); err != nil {
			return true
		}
	}
	return true
}

// ------Pre-Init File Handling-------------------------------------------------

func getExecDirectory() string {
	exePath := errgo.ValueOrPanic(os.Executable())
	exeDir := filepath.Dir(exePath)
	return exeDir
}

var exeDir = getExecDirectory()
var preInitFile = fmt.Sprintf("%s/PreInit.json", exeDir)

func getPreInitData(cfg *runtimeConfig) {
	data := importPreInitData()
	if data == nil {
		fmt.Println("Could not import PreInit JSon data - Skipping Pre-Init.")
		return
	}

	cfg.Address = data["address"].(string)
	cfg.Port = data["port"].(int)
	cfg.Verbosity = data["verbosity"].(int)
	cfg.PrintTree = data["print-tree"].(bool)
	xFormerDur, err := time.ParseDuration(data["xform-timeout"].(string))
	if err != nil {
		cfg.TransformTimeout = xFormerDur
	}

	routeData, exists := data["routes"].([]map[string]any)
	if exists {
		parseRouteCmds(routeData)
	}
}

func importPreInitData() map[string]interface{} {
	data := errgo.ValueOrPanic(os.ReadFile(preInitFile))
	var obj map[string]any
	errgo.PanicIfError(json.Unmarshal(data, &obj))

	return obj
}

/* -----------------------------------------------------------------------------
Expected PreInit.json.
the "routes" field, or children of it, could not exist.
--------------------------------------------------------------------------------
{
  "runtime": {
    "address": "0.0.0.0",
    "port": 8080,
    "verbosity": 2,
    "print-tree": true,
    "xform-timeout": "45s"
  },
  "routes": [
    {
      "name": "default",
      "channels": [
        {
          "name": "inmem",
          "transformers": [
            { "address": "127.0.0.1:7010" },
            { "address": "10.0.0.52:8008" }
          ],
          "subscribers": [
         { "address": "127.0.0.1:1234" },
            { "address": "16.70.18.1:9999" }
          ]
        }
      ]
    }
  ]
}
----------------------------------------------------------------------------- */

// Parse command type funcs append their command to this list.
var CommandList = []commands.Command{}

func parseRouteCmds(routeData []map[string]any) {
	for _, route := range routeData {
		routeName := route["name"].(string)
		addRouteCmd := commands.NewAddRoute(routeName)
		CommandList = append(CommandList, addRouteCmd)

		channelData, exists := route["channels"].([]map[string]any)
		if exists {
			parseChannelCmds(routeName, channelData)
		}

	}
}

func parseChannelCmds(route string, channelData []map[string]any) {
	for _, channel := range channelData {
		channelName := channel["name"].(string)
		addChannelCmd := commands.NewAddChannel(route, channelName)
		CommandList = append(CommandList, addChannelCmd)

		transformers, exists := channel["transformers"].([]map[string]string)
		if exists {
			parseXformCmds(route, channelName, transformers)
		}

		subscribers, exists := channel["subscribres"].([]map[string]string)
		if exists {
			parseSubscriberCmds(route, channelName, subscribers)
		}
	}
}

func parseXformCmds(route, channel string, transformData []map[string]string) {
	for _, transformer := range transformData {
		addr := transformer["address"]
		addTransformerCmd := commands.NewAddTransformer(
			route, channel, addr,
		)
		CommandList = append(CommandList, addTransformerCmd)
	}
}

func parseSubscriberCmds(route, channel string, subData []map[string]string) {
	for _, subscriber := range subData {
		addr := subscriber["address"]
		addSubscriberCmd := commands.NewAddSubscriber(
			route, channel, addr,
		)
		CommandList = append(CommandList, addSubscriberCmd)
	}
}
