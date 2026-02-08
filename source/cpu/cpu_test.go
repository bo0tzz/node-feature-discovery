/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cpu

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/node-feature-discovery/pkg/utils/hostpath"
)

func TestCpuSource(t *testing.T) {
	assert.Equal(t, src.Name(), Name)

	// Check that GetLabels works with empty features
	src.features = nil
	l, err := src.GetLabels()

	assert.Nil(t, err, err)
	assert.Empty(t, l)

}

func TestDiscoverFrequency(t *testing.T) {
	// Save original value and restore after test
	origSysfsDir := hostpath.SysfsDir
	defer func() { hostpath.SysfsDir = origSysfsDir }()

	t.Run("valid cpufreq data", func(t *testing.T) {
		dir := t.TempDir()
		hostpath.SysfsDir = hostpath.HostDir(dir)

		policyDir := filepath.Join(dir, "devices/system/cpu/cpufreq/policy0")
		err := os.MkdirAll(policyDir, 0755)
		assert.Nil(t, err)

		// Values in kHz
		assert.Nil(t, os.WriteFile(filepath.Join(policyDir, "base_frequency"), []byte("2400000\n"), 0644))
		assert.Nil(t, os.WriteFile(filepath.Join(policyDir, "cpuinfo_max_freq"), []byte("4800000\n"), 0644))
		assert.Nil(t, os.WriteFile(filepath.Join(policyDir, "cpuinfo_min_freq"), []byte("800000\n"), 0644))

		features := discoverFrequency()

		assert.Equal(t, "2400", features["base_frequency"])
		assert.Equal(t, "4800", features["cpuinfo_max_freq"])
		assert.Equal(t, "800", features["cpuinfo_min_freq"])
	})

	t.Run("missing base_frequency", func(t *testing.T) {
		dir := t.TempDir()
		hostpath.SysfsDir = hostpath.HostDir(dir)

		policyDir := filepath.Join(dir, "devices/system/cpu/cpufreq/policy0")
		err := os.MkdirAll(policyDir, 0755)
		assert.Nil(t, err)

		assert.Nil(t, os.WriteFile(filepath.Join(policyDir, "cpuinfo_max_freq"), []byte("3600000\n"), 0644))
		assert.Nil(t, os.WriteFile(filepath.Join(policyDir, "cpuinfo_min_freq"), []byte("400000\n"), 0644))

		features := discoverFrequency()

		_, hasBase := features["base_frequency"]
		assert.False(t, hasBase)
		assert.Equal(t, "3600", features["cpuinfo_max_freq"])
		assert.Equal(t, "400", features["cpuinfo_min_freq"])
	})

	t.Run("no cpufreq directory", func(t *testing.T) {
		dir := t.TempDir()
		hostpath.SysfsDir = hostpath.HostDir(dir)

		features := discoverFrequency()
		assert.Empty(t, features)
	})
}
