# Kernloom Standard v1alpha1

## Decision

The Kernloom standard is a registry standard, not a contracts standard.

Contracts define the shape of data exchanged between Forge and KLIQ. Registries
define what words are valid inside that data and what those words mean.

The standard must cover the semantics that should be portable across local PEPs,
vendor sub-control-planes, PIPs and risk engines:

- entity taxonomy: subject, device, workload, resource, session, environment,
  network and communication edge;
- canonical context keys and their value sets;
- signal IDs and evidence semantics;
- metric IDs, units, scopes and label policy;
- capability IDs and whether they observe, assess, decide, enforce or export;
- runtime action IDs, severity, monotonicity, TTL and revert requirements;
- risk levels, risk scopes, confidence and completeness semantics;
- support, fidelity, downgrade and delegation vocabulary for mappings.

Adapter manifests and requirement mappings are not themselves the core standard.
They are standard-conforming artifacts that reference these canonical IDs.

## Non-Goals

- Do not store enterprise policy intent here.
- Do not store customer-specific profiles here.
- Do not store vendor-native target config here.
- Do not make KLIQ a source of enterprise semantics.
- Do not let runtime-learned baselines or graph edges become durable policy
  without proposal and approval.

## Required Integration

Forge:

1. Load this repository as a versioned dependency, Git submodule, release
   artifact or OCI artifact.
2. Validate all policy intent keys, required signals, capabilities, actions,
   risk references and metric IDs against these registries.
3. Validate adapter manifests and requirement mappings against canonical IDs.
4. Emit coverage, downgrade and delegation reports using the standard support
   and fidelity vocabulary.
5. Pin the registry revision and digest in RuntimeBundles and reports.
6. Include or reference a minimal RegistrySnapshot for managed KLIQ nodes.

KLIQ:

1. In managed mode, activate only bundles whose registry revision is trusted and
   whose runtime IDs validate against the pinned snapshot.
2. Reject unknown action capabilities and disallowed action severities.
3. Validate adapter manifests against Metrics, Signals, Capabilities and Label
   Policies.
4. Treat unknown labels as forbidden in managed mode.
5. Do not maintain independent standalone defaults. Standalone operation may
   use a linked registry release, but the registry remains the source of truth.

Contracts:

1. Continue to define data shapes only.
2. Add registry metadata fields only where needed for pinning, snapshots and
   compatibility negotiation.

## Versioning

Registry entries are append-only within a minor line.

Breaking semantic changes require a new major registry version or a new ID.
Deprecation must use `deprecated: true` and, when possible, `replacedBy`.

Runtime artifacts must pin:

```yaml
registry:
  name: kernloom-standard
  version: 0.2.0
  revision: <git-sha-or-release>
  digest: sha256:<digest>
```

## Runtime Safety Invariants

- Runtime actions are restrictive only. Actions that grant access, such as
  `enforce.*.allow`, are config/proposal path entries and require approval.
- Runtime actions must be TTL bounded, leased, audited and auto-revertable.
- Entity scopes and visibility scopes are separate concepts and must not be
  collapsed into a single `allowedScopes` field.
- `unknown` risk is insufficient evidence, not low risk.
- Unknown labels are forbidden in managed mode.
- Unknown context keys in RuntimePolicyPack conditions are rejected.

## Release Hygiene

Release artifacts must be built from the registry working tree without `.git/`.
The released module must not contain local `replace` directives. Development
checkouts may use a workspace-level `go.work` or consumer-module replaces for
`kernloom-contracts` and `kernloom-registries`.
