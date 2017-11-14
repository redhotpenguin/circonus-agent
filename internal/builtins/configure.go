// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build !windows,!linux

package builtins

import (
	"runtime"

	appstats "github.com/maier/go-appstats"
)

func (b *Builtins) configure() error {
	appstats.MapAddInt("builtins", "total", 0)
	b.logger.Info().Msg("no builtins available for " + runtime.GOOS)
	return nil
}