# Changelog

## 0.2.0 - 2026-06-19

- Added policy registries for policy kinds, AccessPolicy schema metadata,
  condition types and structured operators.
- Standardized `endpoint` as a canonical resource type and `any` as a policy
  selector/wildcard only.
- Added action contracts for runtime and config-path actions.
- Split ambiguous scope semantics into entity scopes and visibility scopes.
- Added canonical granularity, gap and support taxonomies.
- Expanded signals with evidence, confidence and response semantics.
- Expanded metrics with label, retention, aggregation and normalization rules.
- Added context keys to the signed runtime registry snapshot.
- Added linter, JSON schema skeletons and release hygiene guidance.
- Removed Forge-local registry defaults as a source of truth.

## 0.1.0 - 2026-06-19

- Initial registry repository structure.
- Added canonical context, risk, capability, action, signal, metric, label,
  taxonomy, adapter, compatibility and trust-boundary registries.
