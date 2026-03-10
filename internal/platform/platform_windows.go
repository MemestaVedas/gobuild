//go:build windows

package platform

import "github.com/MemestaVedas/gobuild/internal/platform/windows"

// New returns the Windows implementation of the Platform interface.
func New() Platform {
	return windows.New()
}
