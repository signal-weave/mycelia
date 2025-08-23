package boot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"mycelia/commands"
	"mycelia/errgo"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Mycelia will check for a PreInit.json file in the same directory as the .exe
// file.

// The PreInit.json file acts as an alternative or addition to providing values
// on startup.

// It will overwrite any CLI values that were placed in it if they were also
// provided by CLI.

// Any CLI value can be placed under the "runtime" field in the json as well as
// any pre-defined routes/channels/transformers/subscribers (provided they
// specify their parent object) that the router should start up with.
// -----------------------------------------------------------------------------

// ------Pre-Init File Handling-------------------------------------------------

func printJson(data map[string]any) {
	b, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Could not pretty print json data")
		return
	}
	fmt.Println(string(b))
}

func getPreInitData(cfg *runtimeConfig) {
    data := importPreInitData()
    if data == nil {
        fmt.Println("Could not import PreInit JSON data - Skipping Pre-Init.")
        return
    }
    parseRuntimeConfigurable(cfg, data)

    routesAny, ok := data["routes"].([]any)
    if !ok {
        return
    }
    routes := make([]map[string]any, 0, len(routesAny))
    for _, r := range routesAny {
        if m, ok := r.(map[string]any); ok {
            routes = append(routes, m)
        }
    }
    parseRouteCmds(routes)
}

func importPreInitData() map[string]any {
	data := errgo.ValueOrPanic(
		os.ReadFile(preInitFile),
	)

	var obj map[string]any

	errgo.PanicIfError(
		json.Unmarshal(data, &obj),
	)

	return obj
}

// Proxy struct for unmarshalling the PreInit.json runtime data into cleanly.
// This handles type conversion - Go marshals json integers to float64 by
// default for whatever fucking reason.
type runtimeData struct {
	Address          *string `json:"address"`
	Port             *int    `json:"port"`
	Verbosity        *int    `json:"verbosity"`
	PrintTree        *bool   `json:"print-tree"`
	TransformTimeout *string `json:"xform-timeout"`
}

// Pipes the non-shape data into the RuntimeCfg
func parseRuntimeConfigurable(cfg *runtimeConfig, data map[string]any) {
	rawRuntimeData, exists := data["runtime"].(map[string]any)
	if !exists {
		return
	}

	fmt.Println(
		"PreInit runtime values found - these will overwrite any CLI values...",
	)

	b, err := json.Marshal(rawRuntimeData)
	if err != nil {
		fmt.Println("Error marshaling runtime data in PreInit.json")
		return
	}

	var rd runtimeData
	err = json.Unmarshal(b, &rd)
	if err != nil {
		fmt.Println("Error unmarshaling runtime data in PreInit.json")
		return
	}

	if rd.Address != nil {
		cfg.Address = *rd.Address
	}
	if rd.Port != nil {
		cfg.Port = *rd.Port
	}
	if rd.Verbosity != nil {
		cfg.Verbosity = *rd.Verbosity
	}
	if rd.PrintTree != nil {
		cfg.PrintTree = *rd.PrintTree
	}
	if rd.TransformTimeout != nil {
		if d, err := time.ParseDuration(*rd.TransformTimeout); err == nil {
			cfg.TransformTimeout = d
		}
	}
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

func parseRouteCmds(routeData []map[string]any) {
    for _, route := range routeData {
        routeName, _ := route["name"].(string)
        id := uuid.New().String()
        addRouteCmd := commands.NewAddRoute(id, routeName)
        CommandList = append(CommandList, addRouteCmd)

        chansAny, ok := route["channels"].([]any)
        if !ok {
            continue
        }
        chans := make([]map[string]any, 0, len(chansAny))
        for _, c := range chansAny {
            if m, ok := c.(map[string]any); ok {
                chans = append(chans, m)
            }
        }
        parseChannelCmds(routeName, chans)
    }
}

func parseChannelCmds(route string, channelData []map[string]any) {
    for _, channel := range channelData {
        channelName, _ := channel["name"].(string)

        id := uuid.New().String()
        addChannelCmd := commands.NewAddChannel(id, route, channelName)
        CommandList = append(CommandList, addChannelCmd)

        // transformers: []any → []map[string]any
        if xAny, ok := channel["transformers"].([]any); ok {
            x := make([]map[string]any, 0, len(xAny))
            for _, t := range xAny {
                if m, ok := t.(map[string]any); ok {
                    x = append(x, m)
                }
            }
            parseXformCmds(route, channelName, x)
        }

        // subscribers: []any → []map[string]any
        if sAny, ok := channel["subscribers"].([]any); ok {
            s := make([]map[string]any, 0, len(sAny))
            for _, v := range sAny {
                if m, ok := v.(map[string]any); ok {
                    s = append(s, m)
                }
            }
            parseSubscriberCmds(route, channelName, s)
        }
    }
}

func parseXformCmds(route, channel string, transformData []map[string]any) {
    for _, t := range transformData {
        addr, _ := t["address"].(string)
        id := uuid.New().String()
        addTransformerCmd := commands.NewAddTransformer(id, route, channel, addr)
        CommandList = append(CommandList, addTransformerCmd)
    }
}

func parseSubscriberCmds(route, channel string, subData []map[string]any) {
    for _, s := range subData {
        addr, _ := s["address"].(string)
        id := uuid.New().String()
        addSubscriberCmd := commands.NewAddSubscriber(id, route, channel, addr)
        CommandList = append(CommandList, addSubscriberCmd)
    }
}
