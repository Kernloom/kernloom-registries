// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2026 Kernloom Contributors

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	idPattern      = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
	versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
)

type registryDoc struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"metadata"`
	Spec map[string]any `yaml:"spec"`
}

type linter struct {
	errors []string

	contextKeys      map[string]bool
	metrics          map[string]bool
	labels           map[string]bool
	labelsAllowed    map[string]bool
	retentionClasses map[string]bool
	signals          map[string]bool
	capabilities     map[string]capabilityInfo
	actionLevels     map[string]bool
	actionContracts  map[string]actionContractInfo
	entityScopes     map[string]bool
	visibilityScopes map[string]bool
	granularities    map[string]bool
	subjectTypes     map[string]bool
	resourceTypes    map[string]bool
	policyKinds      map[string]bool
	conditionTypes   map[string]bool
	policyOperators  map[string]bool
}

type capabilityInfo struct {
	RuntimeAction  bool
	ActionContract string
	Effect         string
	Severity       int
}

type actionContractInfo struct {
	RuntimeAllowed   bool
	RequiresTTL      bool
	RequiresLease    bool
	RequiresAudit    bool
	AutoRevert       string
	CanGrantAccess   bool
	ConfigPathOnly   bool
	RequiresApproval bool
}

func main() {
	root := "registries"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	l := &linter{
		contextKeys:      map[string]bool{},
		metrics:          map[string]bool{},
		labels:           map[string]bool{},
		labelsAllowed:    map[string]bool{},
		retentionClasses: map[string]bool{},
		signals:          map[string]bool{},
		capabilities:     map[string]capabilityInfo{},
		actionLevels:     map[string]bool{},
		actionContracts:  map[string]actionContractInfo{},
		entityScopes:     map[string]bool{},
		visibilityScopes: map[string]bool{},
		granularities:    map[string]bool{},
		subjectTypes:     map[string]bool{},
		resourceTypes:    map[string]bool{},
		policyKinds:      map[string]bool{},
		conditionTypes:   map[string]bool{},
		policyOperators:  map[string]bool{},
	}
	if err := l.run(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(l.errors) > 0 {
		for _, err := range l.errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	fmt.Printf("kernloom-registry-lint: OK (%s)\n", root)
}

func (l *linter) run(root string) error {
	var docs []loadedDoc
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var doc registryDoc
		if err := yaml.Unmarshal(raw, &doc); err != nil {
			l.err(path, "invalid YAML: %v", err)
			return nil
		}
		var node yaml.Node
		if err := yaml.Unmarshal(raw, &node); err != nil {
			l.err(path, "invalid YAML node: %v", err)
			return nil
		}
		docs = append(docs, loadedDoc{path: path, doc: doc, node: &node})
		l.validateEnvelope(path, doc)
		l.validateExpectedKind(path, doc.Kind)
		l.rejectKey(path, &node, "allowedScopes")
		return nil
	}); err != nil {
		return err
	}

	for _, d := range docs {
		l.index(d)
	}
	for _, d := range docs {
		l.validateCrossRefs(d)
	}
	return nil
}

type loadedDoc struct {
	path string
	doc  registryDoc
	node *yaml.Node
}

func (l *linter) validateEnvelope(path string, doc registryDoc) {
	if doc.APIVersion != "kernloom.io/registry/v1alpha1" {
		l.err(path, "unsupported apiVersion %q", doc.APIVersion)
	}
	if doc.Kind == "" {
		l.err(path, "kind is required")
	}
	if doc.Metadata.Name == "" {
		l.err(path, "metadata.name is required")
	}
	if !versionPattern.MatchString(doc.Metadata.Version) {
		l.err(path, "metadata.version %q is not semver-compatible", doc.Metadata.Version)
	}
	if doc.Spec == nil {
		l.err(path, "spec is required")
	}
}

func (l *linter) validateExpectedKind(path, kind string) {
	expectedBySuffix := map[string]string{
		"registries/context/canonical-keys.yaml":              "ContextKeyRegistry",
		"registries/risk/risk-taxonomy.yaml":                  "RiskTaxonomy",
		"registries/capabilities/canonical-capabilities.yaml": "CapabilityRegistry",
		"registries/actions/runtime-actions.yaml":             "RuntimeActionRegistry",
		"registries/actions/runtime-action-contracts.yaml":    "RuntimeActionContractRegistry",
		"registries/signals/canonical-signals.yaml":           "SignalRegistry",
		"registries/metrics/canonical-metrics.yaml":           "MetricRegistry",
		"registries/metrics/label-policies.yaml":              "LabelPolicyRegistry",
		"registries/taxonomy/entities.yaml":                   "TaxonomyRegistry",
		"registries/granularity/canonical-granularities.yaml": "GranularityRegistry",
		"registries/scopes/canonical-scopes.yaml":             "ScopeRegistry",
		"registries/mappings/gap-taxonomy.yaml":               "GapTaxonomy",
		"registries/mappings/support-taxonomy.yaml":           "SupportTaxonomy",
		"registries/adapters/adapter-sdk.yaml":                "AdapterSDKRegistry",
		"registries/compatibility/runtime-contracts.yaml":     "CompatibilityRegistry",
		"registries/security/trust-boundaries.yaml":           "TrustBoundaryRegistry",
		"registries/policy/policy-kinds.yaml":                 "PolicyKindRegistry",
		"registries/policy/operators.yaml":                    "PolicyOperatorRegistry",
		"registries/policy/condition-types.yaml":              "PolicyConditionTypeRegistry",
		"registries/policy/access-policy-schema.yaml":         "PolicySchemaRegistry",
	}
	clean := filepath.ToSlash(path)
	for suffix, expected := range expectedBySuffix {
		if strings.HasSuffix(clean, suffix) && kind != expected {
			l.err(path, "kind %q does not match expected %q", kind, expected)
			return
		}
	}
}

func (l *linter) index(d loadedDoc) {
	switch d.doc.Kind {
	case "ContextKeyRegistry":
		for _, item := range list(d.doc.Spec["keys"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			l.addUnique(d.path, l.contextKeys, id, "context key")
		}
	case "MetricRegistry":
		for _, item := range list(d.doc.Spec["metrics"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			l.addUnique(d.path, l.metrics, id, "metric")
		}
	case "LabelPolicyRegistry":
		for _, item := range list(d.doc.Spec["retentionClasses"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.retentionClasses, id, "retention class")
		}
		for _, item := range list(d.doc.Spec["labels"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			l.addUnique(d.path, l.labels, id, "label policy")
			if boolField(item, "allowed") {
				l.labelsAllowed[id] = true
			}
		}
	case "CapabilityRegistry":
		for _, item := range list(d.doc.Spec["capabilities"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			if _, exists := l.capabilities[id]; exists {
				l.err(d.path, "duplicate capability id %q", id)
			}
			l.capabilities[id] = capabilityInfo{
				RuntimeAction:  boolField(item, "runtimeAction"),
				ActionContract: stringField(item, "actionContract"),
				Effect:         stringField(item, "effect"),
				Severity:       intField(item, "severity"),
			}
		}
	case "RuntimeActionRegistry":
		for _, item := range list(d.doc.Spec["levels"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.actionLevels, id, "action level")
		}
	case "RuntimeActionContractRegistry":
		for _, item := range list(d.doc.Spec["contracts"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			if _, exists := l.actionContracts[id]; exists {
				l.err(d.path, "duplicate action contract id %q", id)
			}
			l.actionContracts[id] = actionContractInfo{
				RuntimeAllowed:   boolField(item, "runtimeAllowed"),
				RequiresTTL:      boolField(item, "requiresTTL"),
				RequiresLease:    boolField(item, "requiresLease"),
				RequiresAudit:    boolField(item, "requiresAudit"),
				AutoRevert:       stringField(item, "autoRevert"),
				CanGrantAccess:   boolField(item, "canGrantAccess"),
				ConfigPathOnly:   boolField(item, "configPathOnly"),
				RequiresApproval: boolField(item, "requiresApproval"),
			}
		}
	case "SignalRegistry":
		for _, item := range list(d.doc.Spec["signals"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			l.addUnique(d.path, l.signals, id, "signal")
			if strings.HasPrefix(id, "signals.") {
				l.err(d.path, "signal id %q must not use redundant signals.* prefix", id)
			}
		}
	case "ScopeRegistry":
		for _, item := range list(d.doc.Spec["entityScopes"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.entityScopes, id, "entity scope")
		}
		for _, item := range list(d.doc.Spec["visibilityScopes"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.visibilityScopes, id, "visibility scope")
		}
	case "GranularityRegistry":
		for _, item := range list(d.doc.Spec["granularities"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.granularities, id, "granularity")
		}
	case "TaxonomyRegistry":
		for _, id := range stringList(d.doc.Spec["subjectTypes"]) {
			l.checkID(d.path, id)
			l.addUnique(d.path, l.subjectTypes, id, "subject type")
		}
		for _, id := range stringList(d.doc.Spec["resourceTypes"]) {
			l.checkID(d.path, id)
			l.addUnique(d.path, l.resourceTypes, id, "resource type")
		}
	case "PolicyKindRegistry":
		for _, item := range list(d.doc.Spec["policyKinds"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.policyKinds, id, "policy kind")
			if stringField(item, "wireKind") == "" {
				l.err(d.path, "policy kind %q must declare wireKind", id)
			}
		}
	case "PolicyOperatorRegistry":
		for _, item := range list(d.doc.Spec["operators"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.addUnique(d.path, l.policyOperators, id, "policy operator")
		}
	case "PolicyConditionTypeRegistry":
		for _, item := range list(d.doc.Spec["conditionTypes"]) {
			id := stringField(item, "id")
			l.checkID(d.path, id)
			l.validateDeprecated(d.path, item, id)
			l.addUnique(d.path, l.conditionTypes, id, "condition type")
			for _, alias := range stringList(item["aliases"]) {
				l.checkID(d.path, alias)
			}
		}
	}
}

func (l *linter) validateCrossRefs(d loadedDoc) {
	switch d.doc.Kind {
	case "CapabilityRegistry":
		for _, item := range list(d.doc.Spec["capabilities"]) {
			id := stringField(item, "id")
			info := l.capabilities[id]
			if info.RuntimeAction && info.ActionContract == "" {
				l.err(d.path, "runtimeAction capability %q must reference actionContract", id)
			}
			if info.ActionContract != "" {
				contract, exists := l.actionContracts[info.ActionContract]
				if !exists {
					l.err(d.path, "capability %q references unknown action contract %q", id, info.ActionContract)
				} else if info.RuntimeAction && !contract.RuntimeAllowed {
					l.err(d.path, "runtimeAction capability %q references non-runtime action contract %q", id, info.ActionContract)
				}
			}
			if info.Effect == "grant" && info.RuntimeAction {
				l.err(d.path, "granting capability %q must not be runtimeAction", id)
			}
			if info.Severity < 0 || info.Severity > 3 {
				l.err(d.path, "capability %q severity must be 0..3", id)
			}
			for _, values := range mapField(item, "requiredGranularity") {
				for _, granularity := range stringList(values) {
					if !l.granularities[granularity] {
						l.err(d.path, "capability %q references unknown granularity %q", id, granularity)
					}
				}
			}
		}
	case "RuntimeActionContractRegistry":
		for _, item := range list(d.doc.Spec["contracts"]) {
			id := stringField(item, "id")
			info := l.actionContracts[id]
			level := stringField(item, "level")
			if level != "" && !l.actionLevels[level] {
				l.err(d.path, "action contract %q references unknown level %q", id, level)
			}
			if info.RuntimeAllowed && !l.capabilities[id].RuntimeAction {
				l.err(d.path, "runtimeAllowed action contract %q has no runtimeAction capability", id)
			}
			if info.RuntimeAllowed && !info.CanGrantAccess {
				if !info.RequiresTTL || !info.RequiresLease || !info.RequiresAudit || info.AutoRevert == "" {
					l.err(d.path, "restrictive runtime action %q requires TTL, lease, audit and autoRevert", id)
				}
			}
			if info.CanGrantAccess && info.RuntimeAllowed {
				l.err(d.path, "granting action %q must not be runtimeAllowed", id)
			}
			if info.CanGrantAccess && (!info.ConfigPathOnly || !info.RequiresApproval) {
				l.err(d.path, "granting action %q must be configPathOnly and require approval", id)
			}
		}
	case "MetricRegistry":
		for _, item := range list(d.doc.Spec["metrics"]) {
			id := stringField(item, "id")
			for _, scope := range stringList(item["entityScopes"]) {
				if !l.entityScopes[scope] {
					l.err(d.path, "metric %q references unknown entityScope %q", id, scope)
				}
			}
			for _, scope := range stringList(item["visibilityScopes"]) {
				if !l.visibilityScopes[scope] {
					l.err(d.path, "metric %q references unknown visibilityScope %q", id, scope)
				}
			}
			for _, label := range stringList(item["allowedLabels"]) {
				if !l.labelsAllowed[label] {
					l.err(d.path, "metric %q allows unknown or forbidden label %q", id, label)
				}
			}
			for _, label := range stringList(item["forbiddenLabels"]) {
				if label == "" {
					l.err(d.path, "metric %q has empty forbidden label", id)
					continue
				}
				if !l.labels[label] {
					l.err(d.path, "metric %q forbids unknown label policy %q", id, label)
				}
				if l.labelsAllowed[label] {
					l.err(d.path, "metric %q lists allowed label %q as forbidden", id, label)
				}
			}
			retentionClass := stringField(item, "retentionClass")
			if retentionClass != "" && !l.retentionClasses[retentionClass] {
				l.err(d.path, "metric %q references unknown retentionClass %q", id, retentionClass)
			}
		}
	case "SignalRegistry":
		for _, item := range list(d.doc.Spec["signals"]) {
			id := stringField(item, "id")
			for _, scope := range stringList(item["entityScopes"]) {
				if !l.entityScopes[scope] {
					l.err(d.path, "signal %q references unknown entityScope %q", id, scope)
				}
			}
			for _, scope := range stringList(item["visibilityScopes"]) {
				if !l.visibilityScopes[scope] {
					l.err(d.path, "signal %q references unknown visibilityScope %q", id, scope)
				}
			}
			for _, metric := range stringList(item["derivedFromMetrics"]) {
				if !l.metrics[metric] {
					l.err(d.path, "signal %q references unknown metric %q", id, metric)
				}
			}
			eligible := item["enforcementEligible"]
			if isTruthy(eligible) && len(stringList(item["suggestedResponses"])) == 0 {
				l.err(d.path, "enforcement-eligible signal %q must declare suggestedResponses", id)
			}
			for _, response := range stringList(item["suggestedResponses"]) {
				if !l.capabilities[response].RuntimeAction && response != "observe.signal.emit" {
					l.err(d.path, "signal %q references unsupported response %q", id, response)
				}
			}
			for _, ctx := range stringList(item["requiredContext"]) {
				if !l.contextKeys[ctx] {
					l.err(d.path, "signal %q references unknown context key %q", id, ctx)
				}
			}
		}
	case "PolicyConditionTypeRegistry":
		for _, item := range list(d.doc.Spec["conditionTypes"]) {
			id := stringField(item, "id")
			for _, signal := range stringList(item["allowedSignals"]) {
				if !l.contextKeys[signal] {
					l.err(d.path, "condition type %q references unknown context key %q", id, signal)
				}
			}
			for _, operator := range stringList(item["allowedOperators"]) {
				if !l.policyOperators[operator] {
					l.err(d.path, "condition type %q references unknown operator %q", id, operator)
				}
			}
		}
	case "PolicySchemaRegistry":
		for _, schema := range list(d.doc.Spec["schemas"]) {
			id := stringField(schema, "id")
			l.checkID(d.path, id)
			for _, action := range list(schema["actions"]) {
				l.checkID(d.path, stringField(action, "id"))
			}
			for _, effect := range list(schema["effects"]) {
				l.checkID(d.path, stringField(effect, "id"))
			}
			for _, conditionType := range stringList(schema["conditionTypes"]) {
				if !l.conditionTypes[conditionType] {
					l.err(d.path, "policy schema %q references unknown condition type %q", id, conditionType)
				}
			}
			for _, operator := range stringList(schema["operators"]) {
				if !l.policyOperators[operator] {
					l.err(d.path, "policy schema %q references unknown operator %q", id, operator)
				}
			}
			for _, selector := range list(schema["subjectSelectorTypes"]) {
				selectorID := stringField(selector, "id")
				l.checkID(d.path, selectorID)
				contextKey := stringField(selector, "contextKey")
				if contextKey != "" && !l.contextKeys[contextKey] {
					l.err(d.path, "subject selector %q references unknown context key %q", selectorID, contextKey)
				}
				canonical := stringField(selector, "canonicalSubjectType")
				if canonical != "" && !l.subjectTypes[canonical] {
					l.err(d.path, "subject selector %q references unknown canonicalSubjectType %q", selectorID, canonical)
				}
			}
			for _, selector := range list(schema["resourceSelectorTypes"]) {
				selectorID := stringField(selector, "id")
				l.checkID(d.path, selectorID)
				contextKey := stringField(selector, "contextKey")
				if contextKey != "" && !l.contextKeys[contextKey] {
					l.err(d.path, "resource selector %q references unknown context key %q", selectorID, contextKey)
				}
				canonical := stringField(selector, "canonicalResourceType")
				if canonical != "" && !l.resourceTypes[canonical] {
					l.err(d.path, "resource selector %q references unknown canonicalResourceType %q", selectorID, canonical)
				}
			}
		}
	}
}

func (l *linter) validateDeprecated(path string, item map[string]any, id string) {
	if !boolField(item, "deprecated") {
		return
	}
	if stringField(item, "deprecationReason") == "" {
		l.err(path, "deprecated id %q must declare deprecationReason", id)
	}
	replacement := stringField(item, "replacedBy")
	if replacement != "" {
		l.checkID(path, replacement)
	}
}

func (l *linter) rejectKey(path string, node *yaml.Node, key string) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				l.err(path, "key %q is not allowed", key)
			}
			l.rejectKey(path, node.Content[i+1], key)
		}
		return
	}
	for _, child := range node.Content {
		l.rejectKey(path, child, key)
	}
}

func (l *linter) checkID(path, id string) {
	if id == "" {
		l.err(path, "id is required")
		return
	}
	if !idPattern.MatchString(id) {
		l.err(path, "id %q does not match canonical ID pattern", id)
	}
}

func (l *linter) addUnique(path string, index map[string]bool, id, label string) {
	if index[id] {
		l.err(path, "duplicate %s id %q", label, id)
	}
	index[id] = true
}

func (l *linter) err(path, format string, args ...any) {
	l.errors = append(l.errors, fmt.Sprintf("%s: %s", path, fmt.Sprintf(format, args...)))
}

func list(v any) []map[string]any {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func mapField(m map[string]any, key string) map[string]any {
	if out, ok := m[key].(map[string]any); ok {
		return out
	}
	return nil
}

func stringField(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

func boolField(m map[string]any, key string) bool {
	if b, ok := m[key].(bool); ok {
		return b
	}
	return false
}

func intField(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func stringList(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func isTruthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "true" || x == "conditional"
	default:
		return false
	}
}
