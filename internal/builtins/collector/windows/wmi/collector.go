// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build windows

package wmi

import (
	"time"

	"github.com/circonus-labs/circonus-agent/internal/builtins/collector"
	cgm "github.com/circonus-labs/circonus-gometrics/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Define stubs to satisfy the collector.Collector interface.
//
// The individual wmi collector implementations must override Collect and Flush.
//
// ID and Inventory are generic and do not need to be overridden unless the
// collector implementation requires it.

// Collect metrics
func (c *wmicommon) Collect() error {
	c.Lock()
	defer c.Unlock()
	return collector.ErrNotImplemented
}

// Flush returns last metrics collected
func (c *wmicommon) Flush() cgm.Metrics {
	c.Lock()
	defer c.Unlock()
	if c.lastMetrics == nil {
		c.lastMetrics = cgm.Metrics{}
	}
	return c.lastMetrics
}

// ID returns id of collector
func (c *wmicommon) ID() string {
	c.Lock()
	defer c.Unlock()
	return c.id
}

// Inventory returns collector stats for /inventory endpoint
func (c *wmicommon) Inventory() collector.InventoryStats {
	c.Lock()
	defer c.Unlock()
	return collector.InventoryStats{
		ID:              c.id,
		LastRunStart:    c.lastStart.Format(time.RFC3339Nano),
		LastRunEnd:      c.lastEnd.Format(time.RFC3339Nano),
		LastRunDuration: c.lastRunDuration.String(),
		LastError:       c.lastError,
	}
}

// Logger returns collector's instance of logger
func (c *wmicommon) Logger() zerolog.Logger {
	return c.logger
}

// cleanName is used to clean the metric name
func (c *wmicommon) cleanName(name string) string {
	return c.metricNameRegex.ReplaceAllString(name, c.metricNameChar)
}

// addMetric to internal buffer if metric is active
func (c *wmicommon) addMetric(metrics *cgm.Metrics, prefix string, mname, mtype string, mval interface{}) error {
	if metrics == nil {
		return errors.New("invalid metric submission")
	}

	if mname == "" {
		return errors.New("invalid metric, no name")
	}

	if mtype == "" {
		return errors.New("invalid metric, no type")
	}

	// cleanup the raw metric name, if needed
	mname = c.cleanName(mname)
	// check status of cleaned metric name
	active, found := c.metricStatus[mname]

	if (found && active) || (!found && c.metricDefaultActive) {
		metricName := mname
		if prefix != "" {
			metricName = prefix + metricNameSeparator + mname
		}
		metricName = cgm.MetricNameWithStreamTags(metricName, c.baseTags)
		(*metrics)[metricName] = cgm.Metric{Type: mtype, Value: mval}
		return nil
	}

	return errors.Errorf("metric (%s) not active", mname)
}

// setStatus is used in Collect to set the collector status
func (c *wmicommon) setStatus(metrics cgm.Metrics, err error) {
	c.Lock()
	if err == nil {
		c.lastError = ""
		c.lastMetrics = metrics
	} else {
		c.lastError = err.Error()
		// on error, ensure metrics are reset
		// do not keep returning a stale set of metrics
		c.lastMetrics = cgm.Metrics{}
	}
	c.lastEnd = time.Now()
	if !c.lastStart.IsZero() {
		c.lastRunDuration = time.Since(c.lastStart)
	}
	c.running = false
	c.Unlock()
}
