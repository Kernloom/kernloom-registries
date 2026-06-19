# Registry Authoring Guide

Registry entries are normative Kernloom semantics. Treat IDs as public API.

## IDs

- Use lowercase dot-separated IDs: `domain.name` or `category.domain.name`.
- Do not reuse an ID with changed meaning.
- Prefer a new ID for breaking semantic changes.
- Mark old IDs with `deprecated: true`, `deprecationReason` and `replacedBy`
  when a replacement exists.

## Runtime Actions

- Runtime actions must be restrictive, TTL bounded, leased, audited and
  reversible.
- Actions that may grant access must set `runtimeAction: false` and use a
  config/proposal path with approval.
- Every runtime capability must reference a runtime action contract with the
  same ID.

## Scopes

- Use `entityScopes` for the thing the metric or signal is about.
- Use `visibilityScopes` for where the observation is valid.
- Do not add `allowedScopes`; the linter rejects it.

## Metrics And Labels

- Metric labels must reference `label-policies.yaml`.
- Unknown labels are forbidden.
- Raw paths, URIs, credentials, cookies, usernames and request IDs must not be
  default labels.
- Use `requiresNormalization: true` when a value needs templating, bucketing or
  other cardinality control before storage.

## Signals

- Signals must declare required evidence.
- Enforcement-eligible signals must list suggested runtime responses.
- Signals may reference only known metrics, context keys and runtime action
  capabilities.

## Risk

- `unknown` means insufficient evidence and must never be treated as low risk.
- Runtime enforcement must meet the minimum quality threshold for the selected
  action level.

## Validation

Run:

```sh
go run ./cmd/kernloom-registry-lint ./registries
```

The linter is the first gate. Consumer modules may add stricter domain-specific
validation, but they must not weaken this registry standard.
