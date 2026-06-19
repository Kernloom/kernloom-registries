# Kernloom Registries

This repository is the normative vocabulary for Kernloom.

It defines the vendor-neutral names, values and semantics that Forge may compile
and KLIQ may execute. It is not a policy repository, not an adapter repository
and not the wire-contract module.

## Boundary

`kernloom-contracts` defines shared wire schemas: RuntimeBundle,
RuntimePolicyPack, RuntimeDecision, LocalRiskAssessment and receipts.

`kernloom-registries` defines the standard vocabulary those schemas may refer
to: policy kinds, policy condition types, operators, context keys,
capabilities, actions, signals, metrics, labels, entity kinds, environment
names, confidence bands and risk levels.

Forge consumes this repository at compile time and pins a registry version or
digest into generated artifacts. KLIQ consumes the pinned registry snapshot at
activation time and rejects unknown or disallowed runtime semantics in managed
mode.

## Registries

- `registries/context/canonical-keys.yaml`: canonical context facts.
- `registries/policy/policy-kinds.yaml`: supported policy document kinds and
  shared envelope rules.
- `registries/policy/access-policy-schema.yaml`: AccessPolicy selectors,
  actions, effects, constraints and schema link.
- `registries/policy/condition-types.yaml`: canonical AccessPolicy condition
  types and their allowed context keys.
- `registries/policy/operators.yaml`: structured condition operators and their
  CEL meaning.
- `registries/risk/risk-taxonomy.yaml`: risk levels, score ranges and quality
  semantics.
- `registries/capabilities/canonical-capabilities.yaml`: canonical capability
  IDs, runtime eligibility, effects and required granularity.
- `registries/actions/runtime-actions.yaml`: action levels, bounds and runtime
  invariants.
- `registries/actions/runtime-action-contracts.yaml`: action contracts for
  runtime and config-path actions.
- `registries/signals/canonical-signals.yaml`: canonical signal IDs, evidence
  requirements and response guidance.
- `registries/metrics/canonical-metrics.yaml`: canonical metric IDs, entity
  scopes, visibility scopes, labels and retention classes.
- `registries/metrics/label-policies.yaml`: metric label cardinality and PII
  policy.
- `registries/scopes/canonical-scopes.yaml`: entity scopes and visibility
  scopes.
- `registries/granularity/canonical-granularities.yaml`: semantic granularity
  levels and downgrade/gap vocabulary.
- `registries/mappings/gap-taxonomy.yaml`: mapping gap categories.
- `registries/mappings/support-taxonomy.yaml`: support, fidelity, strictness and
  delegation vocabulary.
- `registries/taxonomy/entities.yaml`: entity, subject, resource and environment
  vocabulary.
- `registries/adapters/adapter-sdk.yaml`: required adapter roles, target types
  and manifest conformance rules.
- `registries/compatibility/runtime-contracts.yaml`: runtime compatibility and
  activation requirements.
- `registries/security/trust-boundaries.yaml`: runtime trust-boundary
  requirements for Forge, KLIQ and adapters.

## Validation

Run the registry linter before publishing a release:

```sh
go run ./cmd/kernloom-registry-lint ./registries
```

The linter checks registry envelopes, duplicate IDs, ID naming, deprecated
metadata, policy schema references, runtime action contracts, scope references,
metric labels, signal evidence references and forbidden legacy fields such as
`allowedScopes`.

Release archives must exclude `.git/`. Local development may use `go.work` or
consumer-module `replace` directives; the released `kernloom-registries/go.mod`
must not contain local `replace` directives.

## Integration Rule

Adapters may extend capabilities, signals or vendor assessments under their own
namespace, but enterprise policy intent and managed RuntimePolicyPacks must use
canonical registry IDs unless an explicit, versioned mapping declares a
downgrade or vendor delegation.

Policy authoring tools may offer a more natural policy language, but that
language must compile into the canonical policy kinds and values defined here.
For AccessPolicy, `any` is a policy selector/wildcard, not a canonical
`subject.type` or `resource.type` context value. `endpoint` is a canonical
resource type for host, IP or IP:port resources.

Managed RuntimeBundles must include a signed `registry_snapshot`. KLIQ must
reject bundles without a snapshot and must fail closed on unknown runtime
actions, metrics, signals, labels or context references.

## Documentation

- `docs/registry-reference-for-policy-intents.md`: practical table of all
  registries, what they define and where they are used in Policy Intents,
  TargetIntegrationProfiles, adapter mappings and runtime artifacts.
- `docs/registry-authoring-guide.md`: authoring rules for registry entries.
- `docs/compatibility-policy.md`: versioning and compatibility policy.
- `docs/integration-plan.md`: consumer integration plan for Forge, KLIQ and
  contracts.
