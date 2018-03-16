// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package check

import (
	stdlog "log"
	"time"

	"github.com/circonus-labs/circonus-agent/internal/config"
	"github.com/circonus-labs/circonus-gometrics/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// New returns a new check instance
func New(apiClient API) (*Check, error) {
	c := Check{
		manage:     false,
		bundle:     nil,
		metrics:    nil,
		refreshTTL: time.Duration(0),
		logger:     log.With().Str("pkg", "check").Logger(),
	}

	isCreate := viper.GetBool(config.KeyCheckCreate)
	isManaged := viper.GetBool(config.KeyCheckEnableNewMetrics)
	isReverse := viper.GetBool(config.KeyReverse)
	cid := viper.GetString(config.KeyCheckBundleID)
	needCheck := false

	if isReverse || isManaged || (isCreate && cid == "") {
		needCheck = true
	}

	if !needCheck {
		c.logger.Info().Msg("check management disabled")
		return &c, nil // if we don't need a check, return a NOP object
	}

	if apiClient != nil {
		c.client = apiClient
	} else {
		// create an API client
		cfg := &api.Config{
			TokenKey: viper.GetString(config.KeyAPITokenKey),
			TokenApp: viper.GetString(config.KeyAPITokenApp),
			URL:      viper.GetString(config.KeyAPIURL),
			Log:      stdlog.New(c.logger.With().Str("pkg", "check.api").Logger(), "", 0),
			Debug:    viper.GetBool(config.KeyDebugCGM),
		}

		client, err := api.New(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "creating circonus api client")
		}

		c.client = client
	}

	if err := c.setCheck(); err != nil {
		return nil, errors.Wrap(err, "unable to configure check")
	}

	// ensure a) the global check bundle id is set and b) it is set correctly to the
	// check bundle actually being used - need to do this even if the check was
	// created initially since user 'nobody' cannot create or update the configuration
	viper.Set(config.KeyCheckBundleID, c.bundle.CID)

	if isManaged {
		// refresh ttl
		ttl, err := time.ParseDuration(viper.GetString(config.KeyCheckMetricRefreshTTL))
		if err != nil {
			return nil, errors.Wrap(err, "parsing check metric refresh TTL")
		}
		c.refreshTTL = ttl
		c.manage = isManaged
		c.refreshMetrics()
	}

	return &c, nil
}

// RefreshCheckConfig re-loads the check bundle using the API and reconfigures reverse (if needed)
func (c *Check) RefreshCheckConfig() error {
	return c.setCheck()
}

// GetReverseConfig returns the reverse configuration to use for the broker
func (c *Check) GetReverseConfig() (ReverseConfig, error) {
	c.Lock()
	defer c.Unlock()
	if c.revConfig == nil {
		return ReverseConfig{}, errors.New("invalid reverse configuration")
	}
	return *c.revConfig, nil
}

// EnableNewMetrics updates the check bundle enabling any new metrics
func (c *Check) EnableNewMetrics(m *map[string]interface{}) error {
	c.Lock()
	defer c.Unlock()

	if !c.manage {
		return nil
	}

	c.refreshMetrics()

	// compare metric states
	// add any new metrics to check bundle
	// update check bundle via api if needed

	return nil
}
