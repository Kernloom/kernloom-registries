// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2026 Kernloom Contributors

// Package registries loads the canonical Kernloom registry set and produces the
// compact runtime view that Forge signs into RuntimeBundles.
package registries

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	contracts "github.com/kernloom/kernloom-contracts"
	"gopkg.in/yaml.v3"
)

const (
	Name    = "kernloom-standard"
	Version = "0.2.0"
)

//go:embed registries
var embedded embed.FS

// EmbeddedSnapshot returns the registry snapshot for the version of
// kernloom-registries linked into the binary.
func EmbeddedSnapshot() (contracts.RegistrySnapshot, error) {
	return LoadSnapshotFS(embedded, "embedded")
}

// LoadSnapshotDir loads a registry checkout from disk.
func LoadSnapshotDir(root string) (contracts.RegistrySnapshot, error) {
	return LoadSnapshotFS(os.DirFS(root), root)
}

// LoadSnapshotFS loads all standard registry files from fsys.
func LoadSnapshotFS(fsys fs.FS, revision string) (contracts.RegistrySnapshot, error) {
	contextReg, err := loadContext(fsys, "registries/context/canonical-keys.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	riskReg, err := loadRisk(fsys, "registries/risk/risk-taxonomy.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	capReg, err := loadCapabilities(fsys, "registries/capabilities/canonical-capabilities.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	actionReg, err := loadActions(fsys, "registries/actions/runtime-actions.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	actionContractReg, err := loadActionContracts(fsys, "registries/actions/runtime-action-contracts.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	signalReg, err := loadSignals(fsys, "registries/signals/canonical-signals.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	metricReg, err := loadMetrics(fsys, "registries/metrics/canonical-metrics.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	labelReg, err := loadLabels(fsys, "registries/metrics/label-policies.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	granularityReg, err := loadGranularities(fsys, "registries/granularity/canonical-granularities.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	scopeReg, err := loadScopes(fsys, "registries/scopes/canonical-scopes.yaml")
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}

	snap := contracts.RegistrySnapshot{
		Ref: contracts.RegistryRef{
			Name:     Name,
			Version:  Version,
			Revision: revision,
		},
		ContextVersion:   contextReg.Metadata.Version,
		ContextKeys:      contextReg.Spec.Keys,
		RiskLevels:       riskReg.Spec.Levels,
		RiskTaxonomy:     riskReg.Spec.Snapshot(),
		Capabilities:     capReg.Spec.Capabilities,
		ActionLevels:     actionReg.Spec.Levels,
		ActionContracts:  actionContractReg.Spec.Contracts,
		Signals:          signalReg.Spec.Signals,
		Metrics:          metricReg.Spec.Metrics,
		LabelPolicies:    labelReg.Spec.Labels,
		RetentionClasses: labelReg.Spec.RetentionClasses,
		Granularities:    granularityReg.Spec.Granularities,
		Scopes:           scopeReg.Spec,
	}
	sortSnapshot(&snap)
	digest, err := DigestSnapshot(snap)
	if err != nil {
		return contracts.RegistrySnapshot{}, err
	}
	snap.Ref.Digest = digest
	return snap, nil
}

func DigestSnapshot(snapshot contracts.RegistrySnapshot) (string, error) {
	copy := snapshot
	copy.Ref.Digest = ""
	raw, err := json.Marshal(copy)
	if err != nil {
		return "", fmt.Errorf("registry snapshot digest: %w", err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func readYAML[T any](fsys fs.FS, path string) (T, error) {
	var out T
	raw, err := fs.ReadFile(fsys, filepath.ToSlash(path))
	if err != nil {
		return out, err
	}
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("%s: %w", path, err)
	}
	return out, nil
}

type metadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type contextRegistry struct {
	Metadata metadata `yaml:"metadata"`
	Spec     struct {
		Keys []contracts.ContextKeyEntry `yaml:"keys"`
	} `yaml:"spec"`
}

func loadContext(fsys fs.FS, path string) (contextRegistry, error) {
	return readYAML[contextRegistry](fsys, path)
}

type riskRegistry struct {
	Spec riskSpec `yaml:"spec"`
}

func loadRisk(fsys fs.FS, path string) (riskRegistry, error) {
	return readYAML[riskRegistry](fsys, path)
}

type riskSpec struct {
	Scopes                      []string                                  `yaml:"scopes"`
	Levels                      []contracts.RiskLevelEntry                `yaml:"levels"`
	Quality                     map[string]any                            `yaml:"quality"`
	MinimumQualityByActionLevel map[string]contracts.RiskQualityThreshold `yaml:"minimumQualityByActionLevel"`
	SourceTypes                 []contracts.RiskSourceTypeEntry           `yaml:"sourceTypes"`
}

func (s riskSpec) Snapshot() contracts.RiskTaxonomySnapshot {
	return contracts.RiskTaxonomySnapshot{
		Scopes:                      append([]string(nil), s.Scopes...),
		Quality:                     s.Quality,
		MinimumQualityByActionLevel: s.MinimumQualityByActionLevel,
		SourceTypes:                 append([]contracts.RiskSourceTypeEntry(nil), s.SourceTypes...),
	}
}

type capabilityRegistry struct {
	Spec struct {
		Capabilities []contracts.CapabilityEntry `yaml:"capabilities"`
	} `yaml:"spec"`
}

func loadCapabilities(fsys fs.FS, path string) (capabilityRegistry, error) {
	return readYAML[capabilityRegistry](fsys, path)
}

type actionRegistry struct {
	Spec struct {
		Levels []contracts.ActionLevelEntry `yaml:"levels"`
	} `yaml:"spec"`
}

func loadActions(fsys fs.FS, path string) (actionRegistry, error) {
	return readYAML[actionRegistry](fsys, path)
}

type actionContractRegistry struct {
	Spec struct {
		Contracts []contracts.RuntimeActionContractEntry `yaml:"contracts"`
	} `yaml:"spec"`
}

func loadActionContracts(fsys fs.FS, path string) (actionContractRegistry, error) {
	return readYAML[actionContractRegistry](fsys, path)
}

type signalRegistry struct {
	Spec struct {
		Signals []contracts.SignalEntry `yaml:"signals"`
	} `yaml:"spec"`
}

func loadSignals(fsys fs.FS, path string) (signalRegistry, error) {
	return readYAML[signalRegistry](fsys, path)
}

type metricRegistry struct {
	Spec struct {
		Metrics []contracts.MetricEntry `yaml:"metrics"`
	} `yaml:"spec"`
}

func loadMetrics(fsys fs.FS, path string) (metricRegistry, error) {
	return readYAML[metricRegistry](fsys, path)
}

type labelRegistry struct {
	Spec struct {
		RetentionClasses []contracts.RetentionClassEntry `yaml:"retentionClasses"`
		Labels           []contracts.LabelPolicyEntry    `yaml:"labels"`
	} `yaml:"spec"`
}

func loadLabels(fsys fs.FS, path string) (labelRegistry, error) {
	return readYAML[labelRegistry](fsys, path)
}

type granularityRegistry struct {
	Spec struct {
		Granularities []contracts.GranularityEntry `yaml:"granularities"`
	} `yaml:"spec"`
}

func loadGranularities(fsys fs.FS, path string) (granularityRegistry, error) {
	return readYAML[granularityRegistry](fsys, path)
}

type scopeRegistry struct {
	Spec contracts.ScopeRegistryView `yaml:"spec"`
}

func loadScopes(fsys fs.FS, path string) (scopeRegistry, error) {
	return readYAML[scopeRegistry](fsys, path)
}

func sortSnapshot(s *contracts.RegistrySnapshot) {
	sort.Slice(s.ContextKeys, func(i, j int) bool { return s.ContextKeys[i].ID < s.ContextKeys[j].ID })
	sort.Slice(s.RiskLevels, func(i, j int) bool { return s.RiskLevels[i].ID < s.RiskLevels[j].ID })
	sort.Strings(s.RiskTaxonomy.Scopes)
	sort.Slice(s.RiskTaxonomy.SourceTypes, func(i, j int) bool {
		return s.RiskTaxonomy.SourceTypes[i].ID < s.RiskTaxonomy.SourceTypes[j].ID
	})
	sort.Slice(s.Capabilities, func(i, j int) bool { return s.Capabilities[i].ID < s.Capabilities[j].ID })
	sort.Slice(s.ActionLevels, func(i, j int) bool { return s.ActionLevels[i].ID < s.ActionLevels[j].ID })
	sort.Slice(s.ActionContracts, func(i, j int) bool { return s.ActionContracts[i].ID < s.ActionContracts[j].ID })
	sort.Slice(s.Signals, func(i, j int) bool { return s.Signals[i].ID < s.Signals[j].ID })
	sort.Slice(s.Metrics, func(i, j int) bool { return s.Metrics[i].ID < s.Metrics[j].ID })
	sort.Slice(s.LabelPolicies, func(i, j int) bool { return s.LabelPolicies[i].ID < s.LabelPolicies[j].ID })
	sort.Slice(s.RetentionClasses, func(i, j int) bool { return s.RetentionClasses[i].ID < s.RetentionClasses[j].ID })
	sort.Slice(s.Granularities, func(i, j int) bool { return s.Granularities[i].ID < s.Granularities[j].ID })
	sort.Slice(s.Scopes.EntityScopes, func(i, j int) bool { return s.Scopes.EntityScopes[i].ID < s.Scopes.EntityScopes[j].ID })
	sort.Slice(s.Scopes.VisibilityScopes, func(i, j int) bool {
		return s.Scopes.VisibilityScopes[i].ID < s.Scopes.VisibilityScopes[j].ID
	})
}
