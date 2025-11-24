package startup

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mycelia/globals"
	"mycelia/logging"
	"mycelia/system"

	"github.com/google/uuid"
	"github.com/signal-weave/rhizome"
	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// Mycelia will check for a Mycelia_Config.yaml file in the same
// directory as the .exe file.
//
// The YAML file acts exactly like the original JSON one. It overwrites
// matching CLI values and allows defining parameters, routes, channels,
// subscribers, and transformers.
// -----------------------------------------------------------------------------

/* -----------------------------------------------------------------------------
Example Config File:

parameters:
  address: '192.168.1.238'
  port: 5000
  verbosity: 3
  log-output: 2
  print-tree: true
  xform-timeout: '45s'
  consolidate: true
  security-tokens:
    - lockheed
    - martin

routes:
  - name: default
    channels:
      - name: inmem
        strategy: pub-sub
        transformers:
          - address: '127.0.0.1:7010'
          - address: '10.0.0.52:8008'
        subscribers:
          - address: '127.0.0.1:1234'
          - address: '16.70.18.1:9999'
------------------------------------------------------------------------------*/

func getConfigData() {
	_, err := os.Stat(system.ConfigFile)
	if err != nil {
		logging.LogSystemAction(
			"No Mycelia_Config.yaml found, skipping pre-init process.",
		)
		return
	}

	data, err := os.ReadFile(system.ConfigFile)
	if err != nil {
		logging.LogSystemError(
			"Could not import PreInit YAML data - Skipping Pre-Init.",
		)
		return
	}

	var bd system.Data
	err = yaml.Unmarshal(data, &bd)
	if err != nil {
		logging.LogSystemError(fmt.Sprintf("Cannot unmarshal config file: %s", err))
		return
	}

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
	if pd.LogOutput != nil {
		globals.LogOutput = *pd.LogOutput
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

	obj := rhizome.NewObject(
		globals.ObjChannel,
		globals.CmdAdd,
		globals.AckPlcyNoreply,
		id,
		routeName,
		channelName,
		strategy,
		"",
		rhizome.EncodingJson,
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

		obj := rhizome.NewObject(
			globals.ObjTransformer,
			globals.CmdAdd,
			globals.AckPlcyNoreply,
			id,
			routeName,
			channelName,
			addr,
			"",
			rhizome.EncodingJson,
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

		obj := rhizome.NewObject(
			globals.ObjSubscriber,
			globals.CmdAdd,
			globals.AckPlcyNoreply,
			id,
			routeName,
			channelName,
			addr,
			"",
			rhizome.EncodingJson,
			[]byte{},
		)

		system.ObjectList = append(system.ObjectList, obj)
	}
}
