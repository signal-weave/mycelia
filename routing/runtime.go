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

// Verify that the values and sender are valid and then update the globals.
func updateGlobals(cmd *protocol.Command) {
	var rv runtimeUpdater
	err := json.Unmarshal(cmd.Payload, &rv)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not parse payload for globals update from %s",
			cmd.ReturnAdress,
		)
		errgo.NewError(wMsg, globals.VERB_WRN)
		return
	}

	// Is user authorized
	if rv.SecurityToken == nil {
		str.ErrorPrint(
			fmt.Sprintf("Message lacks security token from %s", cmd.ReturnAdress),
		)
	} else {
		if !slices.Contains(globals.SecurityTokens, *rv.SecurityToken) {
			str.ErrorPrint(
				fmt.Sprintf(
					"Unauthorized user attempting globals update from %s",
					cmd.ReturnAdress,
				),
			)
			return
		}
	}

	unPackGlobals(rv, cmd.ReturnAdress)
	globals.PrintDynamicValues()
}

// Unpack the unrtimeUpdater values into the globals.
func unPackGlobals(ru runtimeUpdater, sender string) {
	if ru.Address != nil {
		globals.Address = *ru.Address
	}
	if ru.Port != nil {
		globals.Port = *ru.Port
	}
	if ru.Verbosity != nil {
		globals.Verbosity = *ru.Verbosity
		globals.UpdateVerbosityEnvironVar()
	}
	if ru.PrintTree != nil {
		globals.PrintTree = *ru.PrintTree
	}
	if ru.TransformTimeout != nil {
		newTimeout, err := time.ParseDuration(*ru.TransformTimeout)
		if err != nil {
			wMsg := fmt.Sprintf(
				"Unable to parse transform timeout expr from %s", sender,
			)
			str.WarningPrint(wMsg)
		} else {
			globals.TransformTimeout = newTimeout
		}
	}
	if ru.AutoConsolidate != nil {
		globals.AutoConsolidate = *ru.AutoConsolidate
	}
}
