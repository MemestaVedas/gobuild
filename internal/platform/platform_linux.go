//go:build linux

package platform

import "github.com/MemestaVedas/gobuild/internal/platform/linux"

// New returns the Linux implementation of the Platform interface.
func New() Platform {
	return linux.New()
}
