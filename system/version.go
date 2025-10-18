package system

import (
	"fmt"
)

type BuildType struct {
	MajorVersion int
	MinorVersion int
	PatchVersion int

	ReleaseType ReleaseType

	DevVersion       int
	CandidateVersion int
}

func printVersion(major, minor, patch int) {
	brokerVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	verString := fmt.Sprintf("Running verison: %s", brokerVersion)
	fmt.Println(verString)
}

func (bt BuildType) String() string {
	version := fmt.Sprintf(
		"%d.%d.%d", bt.MajorVersion, bt.MinorVersion, bt.PatchVersion,
	)

	switch bt.ReleaseType {

	case ReleaseDev:
		return fmt.Sprintf("%s-dev.%d", version, bt.DevVersion)

	case ReleaseCandidate:
		return fmt.Sprintf("%s-rc.%d", version, bt.CandidateVersion)

	default:
		return version

	}
}

var BuildMetadata = BuildType{
	MajorVersion: 0,
	MinorVersion: 0,
	PatchVersion: 0,

	ReleaseType: ReleaseDev,

	DevVersion:       0,
	CandidateVersion: 0,
}
