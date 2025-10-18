package system

type ReleaseType int

const (
	// ReleaseDev builds are for development only and not to be released.
	ReleaseDev ReleaseType = iota

	// ReleaseCandidate builds are potential releases that can be tagged after
	// testing, verification, and quality assurance.
	ReleaseCandidate

	// ReleaseStable builds are fully released builds that have passed all
	// quality checks and vetting.
	ReleaseStable

	// ReleaseHotfix builds are patch version builds that fix a number of bugs
	// or critical issues.
	ReleaseHotfix
)

var releaseName = map[ReleaseType]string{
	ReleaseDev:       "dev",
	ReleaseCandidate: "candidate",
	ReleaseStable:    "stable",
	ReleaseHotfix:    "hotfix",
}

func (rt ReleaseType) String() string {
	return releaseName[rt]
}
