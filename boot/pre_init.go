package boot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"

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

func getPreInitData() {
	data := importPreInitData()
	if data == nil {
		fmt.Println("Could not import PreInit JSON data - Skipping Pre-Init.")
		return
	}
	parseRuntimeConfigurable(data)

	routesAny, ok := data["routes"].([]any)
	if !ok {
		str.ActionPrint("No PreInit route data found, skipping PreInit Routes.")
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
func parseRuntimeConfigurable(data map[string]any) {
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
		globals.Address = *rd.Address
	}
	if rd.Port != nil {
		globals.Port = *rd.Port
	}
	if rd.Verbosity != nil {
		globals.Verbosity = *rd.Verbosity
		globals.UpdateVerbosityEnvironVar()
	}
	if rd.PrintTree != nil {
		globals.PrintTree = *rd.PrintTree
	}
	if rd.TransformTimeout != nil {
		if d, err := time.ParseDuration(*rd.TransformTimeout); err == nil {
			globals.TransformTimeout = d
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
		fmt.Println("route:", routeName)

		rawChannels, exists := route["channels"].([]any)
		if !exists {
			continue
		}

		for _, ch := range rawChannels {
			channel, ok := ch.(map[string]any)
			if !ok {
				continue
			}
			channelName, _ := channel["name"].(string)

			rawTransformers, _ := channel["transformers"].([]any)
			for _, t := range rawTransformers {
				transformer, ok := t.(map[string]any)
				if !ok {
					continue
				}
				id := uuid.New().String()
				addr := transformer["address"].(string)
				cmd := protocol.NewCommand(
					globals.OBJ_TRANSFORMER,
					globals.CMD_ADD,
					addr,
					id,
					routeName,
					channelName,
					addr,
					"",
					[]byte{},
				)
				CommandList = append(CommandList, cmd)
			}

			rawSubscribers, _ := channel["subscribers"].([]any)
			for _, s := range rawSubscribers {
				subscriber, ok := s.(map[string]any)
				if !ok {
					continue
				}
				id := uuid.New().String()
				addr := subscriber["address"].(string)
				cmd := protocol.NewCommand(
					globals.OBJ_SUBSCRIBER,
					globals.CMD_ADD,
					addr,
					id,
					routeName,
					channelName,
					addr,
					"",
					[]byte{},
				)
				CommandList = append(CommandList, cmd)
			}
		}
	}
}
