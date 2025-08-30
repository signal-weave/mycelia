package routing

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/str"
)

// -----------------------------------------------------------------------------
// Herein are functions for the broker to update the runtime values of globals.
// It contains security handling and command parsing.
// -----------------------------------------------------------------------------

// The payload struct sent from a globals object by a client.
// SecurityToken is measured against the globals token list. If it can't be
// found in globals then the user message is rejected.
type runtimeUpdater struct {
	SecurityToken    *string `json:"security-token"`
	Address          *string `json:"address"`
	Port             *int    `json:"port"`
	Verbosity        *int    `json:"verbosity"`
	PrintTree        *bool   `json:"print-tree"`
	TransformTimeout *string `json:"xform-timeout"`
	AutoConsolidate  *bool   `json:"consolidate"`
}

func updateGlobals(cmd *protocol.Command) {
	var rv runtimeUpdater
	err := json.Unmarshal(cmd.Payload, &rv)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not parse payload for globals update from %s",
			cmd.Sender,
		)
		errgo.NewError(wMsg, globals.VERB_WRN)
		return
	}

	// Is user authorized
	if rv.SecurityToken == nil {
		str.ErrorPrint(
			fmt.Sprintf("Message lacks security token from %s", cmd.Sender),
		)
	} else {
		if !slices.Contains(globals.SecurityTokens, *rv.SecurityToken) {
			str.ErrorPrint(
				fmt.Sprintf(
					"Unauthorized user attempting globals update from %s",
					cmd.Sender,
				),
			)
			return
		}
	}

	unPackGlobals(rv, cmd.Sender)
	globals.PrintDynamicValues()
}

func unPackGlobals(rv runtimeUpdater, sender string) {
	if rv.Address != nil {
		globals.Address = *rv.Address
	}
	if rv.Port != nil {
		globals.Port = *rv.Port
	}
	if rv.Verbosity != nil {
		globals.Verbosity = *rv.Verbosity
		globals.UpdateVerbosityEnvironVar()
	}
	if rv.PrintTree != nil {
		globals.PrintTree = *rv.PrintTree
	}
	if rv.TransformTimeout != nil {
		newTimeout, err := time.ParseDuration(*rv.TransformTimeout)
		if err != nil {
			wMsg := fmt.Sprintf(
				"Unable to parse transform timeout expr from %s", sender,
			)
			str.WarningPrint(wMsg)
		} else {
			globals.TransformTimeout = newTimeout
		}
	}
	if rv.AutoConsolidate != nil {
		globals.AutoConsolidate = *rv.AutoConsolidate
	}
}
