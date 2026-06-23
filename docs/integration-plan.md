# Integration Plan

## v0.3.0 Status

- Forge loads `kernloom-registries` as the source of canonical policy/runtime
  vocabulary.
- Forge can include a registry snapshot in signed RuntimeBundles.
- KLIQ consumes registry snapshots on the managed runtime path.
- Policy-intent vocabulary now includes registries for detection evaluators,
  missing-context behavior, guardrail types, gap handling, and notification
  bindings.
- RuntimePDP facts and RuntimePolicyPacks have deterministic P1 context/risk
  fixtures; production PIP expansion remains a later phase.
- Local development may use workspace replaces. Release modules must not.

## Phase 1: Move Defaults Into Registries

- Move Forge `registries/context/canonical-keys.yaml` into this repo.
- Move KLIQ former embedded defaults from `pkg/registry/defaults.go` into YAML here:
  metrics, signals, capabilities and label policies.
- Remove independent embedded defaults. Any embedded registry data must come
  from a pinned `kernloom-registries` release.

## Phase 2: Forge Loader

Add a Forge registry loader package that reads this repo layout and produces:

- policy kind registry;
- AccessPolicy schema metadata;
- policy condition type registry;
- policy operator registry;
- context key registry;
- metric registry;
- signal registry;
- capability registry;
- runtime action registry;
- detection evaluator registry;
- missing-context behavior registry;
- guardrail type registry;
- gap-handling behavior registry;
- notification binding registry;
- risk taxonomy;
- taxonomy bundle;
- registry digest.

Forge compiler checks:

- policy kind and apiVersion are supported;
- subject/resource selectors are valid policy selectors;
- `any` is handled as wildcard selector, not as a canonical entity type;
- condition types and structured operators are known;
- policy condition keys exist;
- RequirementPolicy missing/freshness/confidence behavior is explicit;
- DetectionPolicy evaluator and state requirements are explicit;
- CapabilityRequirement gaps can make a target non-deployable;
- enum values are valid;
- required capabilities are canonical or explicitly mapped;
- runtime actions are restrictive and within profile bounds;
- risk levels and confidence thresholds use standard values;
- labels used for baselines are allowed.

## Phase 3: Runtime Snapshot

Add a minimal RegistrySnapshot to RuntimeBundle or an adjacent signed artifact.
KLIQ needs the managed runtime view:

- context keys and context registry version;
- metric IDs, entity scopes, visibility scopes, labels and retention class;
- signal IDs, evidence requirements and suggested responses;
- runtime action capability IDs, severity, max levels and action contracts;
- detection evaluator and missing-context behavior IDs relevant to runtime
  response/detection validation;
- label policies;
- risk levels and risk quality thresholds;
- scopes and granularities.

## Phase 4: Replace Hardcoded Runtime Lists

Replace KLIQ hardcoded lists such as action severity maps with the pinned
registry snapshot.

Local operation may use a linked or embedded registry snapshot from this repo,
but managed mode must reject unknown IDs and bundles without a signed snapshot.

## Phase 5: Adapter Conformance

All adapter manifests must validate against this standard:

- `provides.metrics` are known metric IDs;
- `provides.signals` are known signal IDs;
- `consumes.actions` are known runtime action capabilities, not config-only
  grant actions;
- selected metric labels are allowed;
- vendor-native mappings declare support, fidelity and delegation.
