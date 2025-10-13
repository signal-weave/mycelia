package routing

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"mycelia/errgo"
	"mycelia/globals"
	"mycelia/logging"
	"mycelia/protocol"

	"github.com/signal-weave/siglog"
)

// -----------------------------------------------------------------------------
// Herein are functions for the broker to update the runtime values of globals.
// It contains security handling and object parsing.
// -----------------------------------------------------------------------------

// The payload struct sent from a globals object by a client.
// SecurityToken is measured against the globals token list. If it can't be
// found in globals then the user message is rejected.
type runtimeUpdater struct {
	SecurityToken    *string          `json:"security-token"`
	Address          *string          `json:"address"`
	Port             *int             `json:"port"`
	Verbosity        *siglog.LogLevel `json:"verbosity"`
	PrintTree        *bool            `json:"print-tree"`
	TransformTimeout *string          `json:"xform-timeout"`
	AutoConsolidate  *bool            `json:"consolidate"`
}

// Verify that the values and sender are valid and then update the globals, if
// they are.
// Returns if the user was verified or not.
func updateGlobals(obj *protocol.Object) bool {
	var rv runtimeUpdater
	err := json.Unmarshal(obj.Payload, &rv)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not parse payload for globals update from %s",
			obj.Responder.C.RemoteAddr().String(),
		)
		errgo.NewError(wMsg, globals.VERB_WRN)
		return false
	}

	// Is user authorized
	if rv.SecurityToken == nil {
		logging.LogObjectError(
			fmt.Sprintf("Message lacks security token from %s",
				obj.Responder.RemoteAddr(),
			), obj.UID,
		)
	} else {
		if !slices.Contains(globals.SecurityTokens, *rv.SecurityToken) {
			logging.LogObjectError(
				fmt.Sprintf(
					"Unauthorized user attempting globals update from %s",
					obj.Responder.RemoteAddr(),
				), obj.UID,
			)
			return false
		}
	}

	unpackGlobals(rv, obj.Responder.C.RemoteAddr().String())
	globals.PrintDynamicValues()
	return true
}

// Unpack the unrtimeUpdater values into the globals.
func unpackGlobals(ru runtimeUpdater, sender string) {
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
			logging.LogSystemWarning(wMsg)
		} else {
			globals.TransformTimeout = newTimeout
		}
	}
	if ru.AutoConsolidate != nil {
		globals.AutoConsolidate = *ru.AutoConsolidate
	}
}
