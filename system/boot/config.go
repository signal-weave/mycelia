package boot

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
	"mycelia/system"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Mycelia will check for a Mycelia_Config.json file in the same directory as
// the .exe file.

// The Mycelia_Config.json file acts as an alternative or addition to providing
// values on startup.

// It will overwrite any CLI values that were placed in it if they were also
// provided by CLI.

// Any CLI value can be placed under the "parameters" field in the json as well
// as any pre-defined routes/channels/transformers/subscribers (provided they
// specify their parent object) that the router should start up with.
// -----------------------------------------------------------------------------

func getConfigData() {
	_, err := os.Stat(system.ConfigFile)
	if err != nil {
		str.ActionPrint("No Mycelia_Config.json found, skipping pre-init process.")
	}
	data, err := os.ReadFile(system.ConfigFile)
	if err != nil {
		str.ErrorPrint("Could not import PreInit JSON data - Skipping Pre-Init.")
		return
	}

	var bd system.SystemData
	err = json.Unmarshal(data, &bd)

	if bd.Parameters != nil {
		parseRuntimeConfigurable(*bd.Parameters)
	}
	if bd.Routes != nil {
		parseRouteObjects(*bd.Routes)
	}
}

// Update globals from non-routing data.
func parseRuntimeConfigurable(pd system.ParamData) {
	fmt.Println(
		"PreInit runtime values found - these will overwrite any CLI values...",
	)

	if pd.Address != nil {
		globals.Address = *pd.Address
	}
	if pd.Port != nil {
		globals.Port = *pd.Port
	}
	if pd.Verbosity != nil {
		globals.Verbosity = *pd.Verbosity
		globals.UpdateVerbosityEnvironVar()
	}
	if pd.PrintTree != nil {
		globals.PrintTree = *pd.PrintTree
	}
	if pd.TransformTimeout != nil {
		if d, err := time.ParseDuration(*pd.TransformTimeout); err == nil {
			globals.TransformTimeout = d
		}
	}
	if pd.AutoConsolidate != nil {
		globals.AutoConsolidate = *pd.AutoConsolidate
	}
	if pd.SecurityToken != nil {
		globals.SecurityTokens = *pd.SecurityToken
	}
}

/* -----------------------------------------------------------------------------
Expected Mycelia_Config.json.
the "routes" field, or children of it, could not exist.
--------------------------------------------------------------------------------
{
  "parameters": {
    "address": "0.0.0.0",
    "port": 8080,
    "verbosity": 2,
    "print-tree": true,
    "xform-timeout": "45s",
	"consolidate": true,
    "security-tokens": [
      "lockheed",
      "martin"
    ]
  },
  "routes": [
    {
      "name": "default",
      "channels": [
        {
          "name": "inmem",
		  "strategy": "pub-sub",
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

func parseRouteObjects(routeData []map[string]any) {
	for _, route := range routeData {
		routeName, _ := route["name"].(string)

		rawChannels, exists := route["channels"].([]any)
		if !exists {
			continue
		}

		for _, ch := range rawChannels {
			channel, ok := ch.(map[string]any)
			if !ok {
				continue
			}

			parseChannels(channel, routeName)
			parseTransformers(channel, routeName)
			parseSubscribers(channel, routeName)
		}
	}
}

func parseChannels(channelData map[string]any, routeName string) {
	channelName, _ := channelData["name"].(string)
	strategyName, _ := channelData["strategy"].(string)
	strategy := strconv.Itoa(int(globals.StrategyValue[strategyName]))
	id := uuid.New().String()

	obj := protocol.NewObject(
		globals.OBJ_CHANNEL,
		globals.CMD_ADD,
		globals.ACK_PLCY_NOREPLY,
		id,
		routeName,
		channelName,
		strategy,
		"",
		[]byte{},
	)

	system.ObjectList = append(system.ObjectList, obj)
}

func parseTransformers(channelData map[string]any, routeName string) {
	channelName, _ := channelData["name"].(string)
	rawTransformers, _ := channelData["transformers"].([]any)
	for _, t := range rawTransformers {
		transformer, ok := t.(map[string]any)
		if !ok {
			continue
		}
		id := uuid.New().String()
		addr := transformer["address"].(string)
		obj := protocol.NewObject(
			globals.OBJ_TRANSFORMER,
			globals.CMD_ADD,
			globals.ACK_PLCY_NOREPLY,
			id,
			routeName,
			channelName,
			addr,
			"",
			[]byte{},
		)
		system.ObjectList = append(system.ObjectList, obj)
	}
}

func parseSubscribers(channelData map[string]any, routeName string) {
	channelName, _ := channelData["name"].(string)
	rawSubscribers, _ := channelData["subscribers"].([]any)
	for _, s := range rawSubscribers {
		subscriber, ok := s.(map[string]any)
		if !ok {
			continue
		}
		id := uuid.New().String()
		addr := subscriber["address"].(string)
		obj := protocol.NewObject(
			globals.OBJ_SUBSCRIBER,
			globals.CMD_ADD,
			globals.ACK_PLCY_NOREPLY,
			id,
			routeName,
			channelName,
			addr,
			"",
			[]byte{},
		)
		system.ObjectList = append(system.ObjectList, obj)
	}
}
