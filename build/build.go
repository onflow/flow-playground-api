// Package build contains information about the build that is injected at build-time.
//
// To use this package, simply import it in your program, then add build
// arguments like the following:
//
//   go build -ldflags "-X github.com/dapperlabs/flow-playground-api/build.version=v1.0.0"
package build

import (
	"github.com/Masterminds/semver"
)

// Default value for build-time-injected version strings.
const undefined = "undefined"

// The following variables are injected at build-time using ldflags.
var version string

// Version returns the semantic version of this build.
func Version() *semver.Version {
	if version == undefined {
		return nil
	}

	return semver.MustParse(version)
}

// If any of the build-time-injected variables are empty at initialization,
// mark them as undefined.
func init() {
	if len(version) == 0 {
		version = undefined
	}
}
