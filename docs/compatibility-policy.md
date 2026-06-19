# Compatibility Policy

The Kernloom registry standard follows semantic versioning for registry
semantics.

## Patch Releases

Patch releases may:

- clarify descriptions;
- add non-normative documentation;
- fix typos that do not change semantics;
- add schema or linter checks for already-invalid content.

## Minor Releases

Minor releases may:

- add new IDs;
- add optional fields;
- add stricter metadata for newly introduced IDs;
- deprecate old IDs without removing them;
- extend snapshots with optional runtime views.

Existing IDs are append-only within a minor line.

## Major Releases

Major releases are required when:

- an existing ID changes meaning;
- an existing runtime action changes effect or monotonicity;
- required fields are removed or renamed;
- a previously valid managed RuntimeBundle would be rejected for reasons other
  than an explicit safety invariant.

## Runtime Activation

Managed KLIQ nodes activate only signed RuntimeBundles whose registry snapshot
is trusted by name, version and digest. Unknown runtime actions, metrics,
signals, labels and context references fail closed.

Standalone mode may use an embedded or linked snapshot from this repository,
but it must not carry independent fallback defaults.
