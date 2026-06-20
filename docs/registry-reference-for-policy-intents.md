# Registry Reference For Policy Intents And Targets

This reference explains what each Kernloom registry defines and where it is
used when you write a policy intent, a target profile, or an adapter mapping.

Short version: an `AccessPolicy` describes business intent. A target
(`TargetIntegrationProfile`) describes how a concrete adapter may carry that
intent. Adapter manifests and mappings describe what the target can really do.
Forge connects these layers and creates reports, `RuntimePolicyPack`s, or
signed `RuntimeBundle`s.

## Layers

| Layer | Artifact | Uses Registries For |
|---|---|---|
| Policy intent | `kind: AccessPolicy` | Policy kind, subject/resource selectors, actions, effects, condition types, operators, context keys, risk/posture values, constraints |
| Target | `kind: TargetIntegrationProfile` | Allowed runtime actions, runtime mode, ownership, safety constraints |
| Adapter | Capability, action, and mapping manifests | Metrics/signals, capabilities, support levels, downgrades |
| Forge output | `EnforcementPlan`, report, `RuntimePolicyPack`, `RuntimeBundle` | Compiled coverage, runtime rules, registry snapshot |
| KLIQ runtime | `RuntimePolicyPack` / `RuntimeBundle` | Signature, snapshot, allowed actions, action levels, risk quality, unknown IDs |

## Registry Table

| Registry | File | Defines | Use In Policy Intent / Target |
|---|---|---|---|
| Policy Kinds | `registries/policy/policy-kinds.yaml` | Supported policy document kinds such as `AccessPolicy`, shared envelope fields, status, and schema links. | `apiVersion: kernloom.io/v1` and `kind: AccessPolicy` come from this layer. Natural policy languages must compile to these canonical kinds. |
| AccessPolicy Schema | `registries/policy/access-policy-schema.yaml` and `schemas/access-policy.schema.json` | AccessPolicy selectors, `action: access`, effects `allow` and `deny`, enforcement constraint fields, and the decision that `any` is a selector/wildcard. | Validates the shape of an `AccessPolicy`. `subject.type: any` means wildcard selector; it is not a canonical `subject.type` context value. |
| Condition Types | `registries/policy/condition-types.yaml` | Condition classes such as `authentication_strength`, `risk_level`, `device_posture`, `network_tuple`, `session_context`, and their allowed context keys. | Directly in `AccessPolicy.spec.conditions[].type`. Example: `type: network_tuple` may use `network.source`, `network.destination`, `network.protocol`, or `network.port`. |
| Operators | `registries/policy/operators.yaml` | Structured operators `eq`, `neq`, `gte`, `lte`, `gt`, `lt`, `in`, `not_in` and their CEL meaning. | Directly in `AccessPolicy.spec.conditions[].operator`. Forge converts structured conditions into equivalent CEL when needed. |
| Context Keys | `registries/context/canonical-keys.yaml` | Canonical facts such as `subject.risk.level`, `device.posture.status`, `session.authentication.strength`, `network.protocol`, `network.port`, including type, values, scope, TTL, and source. | Directly in `AccessPolicy.spec.conditions[].signal`. Example: `signal: subject.risk.level`, `operator: eq`, `value: low`. Forge checks whether an adapter or runtime path can provide or compensate this meaning. |
| Risk Taxonomy | `registries/risk/risk-taxonomy.yaml` | Risk levels `low`, `medium`, `high`, `critical`, `unknown`, score ranges, quality fields, and minimum quality per action level. | Used through context keys such as `subject.risk.level`; KLIQ also uses it for local risk and action gates. `unknown` means not enough evidence and must not be treated as `low`. |
| Canonical Capabilities | `registries/capabilities/canonical-capabilities.yaml` | Neutral capabilities such as `observe.network.flow`, `assess.risk.local`, `enforce.access.deny`, `enforce.traffic.rate_limit`, including runtime eligibility, effect, severity, granularity, and allowed paths. | Usually not written directly in `AccessPolicy`. Used by Forge output, target profiles, adapter mappings, and runtime packs. Grant actions such as `enforce.network.allow` are config/proposal-only, not runtime actions. |
| Runtime Actions | `registries/actions/runtime-actions.yaml` | Runtime levels `observe`, `soft`, `hard`, `block`, bounds such as `rate_limit` and `block`, and rules for TTL, lease, audit, auto-revert, and target scopes. | KLIQ uses these levels for RuntimePDP decisions. Policy or target constraints can limit a node to observe, rate limit, or block. |
| Runtime Action Contracts | `registries/actions/runtime-action-contracts.yaml` | The contract for each action: `runtimeAllowed`, level, default/max TTL, required confidence, allowed decision sources, reversibility, and whether the action can grant access. | Forge may only write valid runtime actions into `RuntimePolicyPack` or `RuntimeBundle`. KLIQ must reject bundles with unknown, non-runtime, or out-of-bounds actions. |
| Signals | `registries/signals/canonical-signals.yaml` | Derived signals such as `source.pps_high`, `source.scan_suspected`, `graph.edge_metric_deviation`, `node.trust_untrusted`, including evidence, confidence gates, and suggested responses. | Adapters and KLIQ produce or normalize these signals. Advanced policies may refer to them, but they are more often used indirectly for reports, risk, RuntimePDP facts, and suggested actions. |
| Metrics | `registries/metrics/canonical-metrics.yaml` | Metrics such as `network.packets_per_second`, `network.rate_limit_drop_rate`, `http.auth_fail_rate`, `trust.attestation_fail_rate`, including scope, labels, aggregation, retention, and baseline support. | Not usually written directly in `AccessPolicy`. Adapters declare `provides.metrics`; KLIQ uses them for baselines, signals, and risk. Forge can use them to see if a target can provide evidence for an intent. |
| Label Policies | `registries/metrics/label-policies.yaml` | Allowed and forbidden metric labels, cardinality, PII risk, and retention classes. | Adapters may only publish allowed labels. Policy intent should not rely on raw paths, URIs, credentials, cookies, usernames, or request IDs as stable dimensions. |
| Scopes | `registries/scopes/canonical-scopes.yaml` | Entity scopes such as `src_ip`, `five_tuple`, `service`, `subject`, `device`, `communication_edge`, and visibility scopes `local`, `site`, `tenant`, `global`. | Targets and adapters declare the level where they can observe, decide, or enforce. Forge uses this to detect downgrades such as user-level intent enforced only at IP level. |
| Granularity | `registries/granularity/canonical-granularities.yaml` | Semantic granularities such as `ip`, `ip_port`, `five_tuple`, `user_identity`, `workload_identity`, `service_identity`, `application_segment`, plus gap types. | Key for downgrade reports. Example: a policy asks for `subject.role`, but KLShield can only see IP/cgroup locally; Forge marks a semantic downgrade or partial support. |
| Support Taxonomy | `registries/mappings/support-taxonomy.yaml` | Support levels `full`, `partial`, `conditional`, `full_vendor_native`, `not_supported`, fidelity levels, strictness modes, and delegation vocabulary. | Used in adapter mappings and `EnforcementPlan`s. `AccessPolicy.enforcementConstraints` can decide whether partial support, downgrade, or vendor delegation is allowed. |
| Gap Taxonomy | `registries/mappings/gap-taxonomy.yaml` | Standard gap types such as `granularity_gap`, `semantic_downgrade`, `enforcement_gap`, `context_gap`, `audit_gap`, `revert_gap`, with default severity. | Forge reports and Config PDP use these IDs so operators can see why a target is only partly deployable or not deployable. |
| Entity Taxonomy | `registries/taxonomy/entities.yaml` | Base vocabulary for entity, subject, resource, and environment types; network zones; fact status; support/fidelity; and delegation status. `endpoint` is a canonical resource type. | Used for canonical resource/environment types and concrete subject identities. Forge selectors such as `role`, `group`, or `any` are policy selectors and must map to canonical subjects or wildcard behavior. |
| Adapter SDK | `registries/adapters/adapter-sdk.yaml` | Adapter roles such as `pip_read`, `signal_engine`, `risk_engine`, `pep`, `config_adapter`; target types; and required manifest fields. | Not used directly in `AccessPolicy`. It is the checklist for new adapters: which roles, metrics, signals, actions, labels, and trust boundary they must declare before Forge can use them as a target. |
| Runtime Compatibility | `registries/compatibility/runtime-contracts.yaml` | Runtime contract versions, required `registry_snapshot` fields, and activation rules such as reject unsigned, expired, rollback, unknown action, unknown metric, forbidden label. | Managed mode: Forge embeds the snapshot in `RuntimeBundle`s. KLIQ validates against that snapshot and rejects invalid bundles. |
| Trust Boundaries | `registries/security/trust-boundaries.yaml` | Trust roles `forge`, `kliq`, `pip_adapter`, `pep_adapter`, `config_adapter`, `vendor_control_plane`, and runtime/adapter security requirements. | Policy intent does not describe the trust boundary itself. Targets and adapters must meet it: Forge signs, KLIQ verifies locally, runtime actions need provenance and receipts, and learned state must not override signed policy. |

## Which Registries Do I Use When Writing A Policy?

For a normal `AccessPolicy`, these registries are directly visible:

- Policy Kinds and AccessPolicy Schema for `apiVersion`, `kind`,
  `subject.type`, `resource.type`, `action` and `effect`.
- Entity Taxonomy for canonical subject and resource types. Forge selectors such
  as `role` or `group` must later map to canonical subjects.
- Condition Types for `conditions[].type`.
- Operators for `conditions[].operator`.
- Context Keys for `conditions[].signal`.
- Risk Taxonomy for allowed risk values and the meaning of `unknown`.
- Support, Gap, and Granularity indirectly when `enforcementConstraints` decide
  whether downgrades or delegation are allowed.

Example:

```yaml
apiVersion: kernloom.io/v1
kind: AccessPolicy
metadata:
  name: admin-access
spec:
  subject:
    type: user
    ref: admins
  action: access
  resource:
    type: application_group
    ref: admin-apps
  conditions:
    - id: require-phishing-resistant-mfa
      type: authentication_strength
      signal: session.authentication.strength
      operator: eq
      value: phishing_resistant_mfa
    - id: require-low-risk
      type: risk_level
      signal: subject.risk.level
      operator: eq
      value: low
    - id: require-healthy-device
      type: device_posture
      signal: device.posture.status
      operator: eq
      value: healthy
  enforcementConstraints:
    allowDelegation: false
    allowSemanticDowngrade: false
    minimumFidelity: high
  effect: allow
```

Here, `AccessPolicy`, `access`, `allow`, the condition types and operators come
from the Policy registries. `user` and `application_group` come from the Entity
Taxonomy. If the policy uses `role`, `group`, or `any` as a subject selector
instead, Forge must map it through IdP/CMDB context or wildcard behavior to
canonical subjects. The three `signal` values come from Context Keys. `low` and
the risk quality rules come from the Risk Taxonomy. Forge checks the constraints
against support, gap, and granularity rules.

## Natural Policy Intent

Operators may write a more natural intent language, but Forge must compile it
to the canonical YAML model before planning or bundling. Example authoring text:

```text
protect "ziti-controller"
allow group "kernloom-admins" to access "ziti-controller"
require "subject.risk.level" eq "low"
require "session.authentication.strength" in ["mfa", "phishing_resistant_mfa"]
default deny access to "ziti-controller"
when denied access to "ziti-controller" exceeds 5 within 15m then alert
never auto_block group "kernloom-admins"
```

Quotes around variable values are optional for simple IDs. They are useful
because they make subjects, resources and environments visible and allow names
with spaces without changing the policy meaning.

The compiler should split this into canonical artifacts:

- `protect` and `allow` become an `AccessPolicy`.
- `require` becomes `AccessPolicy.spec.conditions[]`. Multiple `require` lines
  are conjunctive by default: all must hold for the allow intent.
- `default deny` becomes target policy default behavior or a runtime pack
  `default_effect`, depending on target support.
- each `when ... then ...` becomes one runtime response rule or report action.
- `never` and `max action` become safety guardrails or enforcement constraints.

Multiple `when ... then ...` statements are allowed. In the current runtime
contract, each compiled `RuntimePolicyPack.spec.rules[]` has one `when` and one
`then`. If one natural rule names several actions, Forge should expand it into
several runtime rules with the same `when`, unless a future ordered action group
contract is added.

Do not use `when ... then ...` for access requirements such as low risk or MFA.
Use `require` for access conditions and `when ... then ...` for response
behavior after something is observed.

`alert` is only a natural alias. The standard response vocabulary comes from
`registries/actions/runtime-action-contracts.yaml` and
`registries/capabilities/canonical-capabilities.yaml`.

| Natural action | Canonical ID | Notes |
|---|---|---|
| `alert` | `observe.signal.emit` | Emit/raise an observable signal. |
| `finding` | `export.finding` | Export a reportable finding. |
| `rate_limit` | `enforce.network.rate_limit` | Soft, reversible runtime restriction. |
| `connection_limit` | `enforce.traffic.connection_limit` | Soft traffic restriction. |
| `bandwidth_limit` | `enforce.traffic.bandwidth_limit` | Soft traffic restriction. |
| `syn_protect` | `enforce.network.syn_protect` | Soft network protection. |
| `deny` or `block` | `enforce.access.deny` | Block-level access restriction. |
| `network_deny` | `enforce.network.deny` | Block-level network restriction. |
| `drop` | `enforce.traffic.drop` | Block-level traffic restriction. |
| `tarpit` | `enforce.traffic.tarpit` | Hard traffic restriction. |
| `quarantine` | `enforce.network.quarantine` | Block-level isolation/quarantine. |

Natural intent may also use canonical IDs directly, for example
`then enforce.traffic.drop for 5m`. Granting actions such as
`enforce.network.allow` are config/proposal-only, not local runtime response
actions.

## Which Registries Do I Use When Writing A Target?

A `TargetIntegrationProfile` describes what a concrete target may do. The most
important registries are Runtime Actions, Capabilities, Support, Granularity,
and Trust Boundaries.

Example target questions:

| Question | Registry |
|---|---|
| May this target do runtime enforcement? | `capabilities`, `runtime-actions`, `runtime-action-contracts`, `trust-boundaries` |
| May it block, or only rate limit? | `runtime-actions`, `runtime-action-contracts` |
| Does the action grant access or restrict access? | `capabilities`, `runtime-action-contracts` |
| At what level can it enforce: IP, five-tuple, user, service? | `granularity`, `scopes` |
| Is a mapping full, partial, delegated, or unsupported? | `support-taxonomy`, `gap-taxonomy` |
| May an adapter provide a given signal or metric? | `signals`, `metrics`, `label-policies`, `adapter-sdk` |

## Where The IDs End Up

| Output | Registry Link |
|---|---|
| `EnforcementPlan` | Support level, fidelity, downgrades, gaps, capabilities, target coverage |
| `RuntimePolicyPack.spec.capabilities_required` | Canonical runtime capabilities such as `enforce.access.deny` |
| `RuntimePolicyPack.spec.rules[].when` | Runtime facts derived from context, risk, or signals |
| `RuntimePolicyPack.spec.rules[].then.capability` | Canonical runtime capability/action, for example `enforce.traffic.rate_limit` |
| `RuntimeBundle.spec.registry_snapshot` | Pinned version/digest and snapshot of runtime-relevant registries |
| KLIQ activation | Compatibility, action, capability, signal, metric, and label rules |

## Practical Rules

- Keep policy intent business-focused: subject, resource, conditions, effect,
  constraints.
- Put runtime actions near policy only when you intentionally limit bounds or
  allowed actions.
- Adapters may extend vendor semantics, but must map them to canonical
  registries or declare delegation/downgrade.
- `unknown` is a real state, not "low risk" and not "healthy".
- Missing or unknown context must be reported as missing evidence or a context
  gap. It must not silently become a hard runtime block.
- Grant actions such as Allow/Network Allow belong in config/proposal paths,
  not in local runtime enforcement packs.
- Managed `RuntimeBundle`s must include a `registry_snapshot`. KLIQ must reject
  unknown or disallowed runtime semantics.
