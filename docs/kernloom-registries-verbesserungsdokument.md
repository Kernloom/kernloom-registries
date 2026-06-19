# Kernloom Registries – konkretes Verbesserungsdokument

**Stand:** 2026-06-19  
**Scope:** Nur Optimierung der Kernloom Registries. Kein Natural Policy Model, keine Controlled Natural Language, kein LLM-Layer.  
**Ausgangsbasis:** `kernloom-registries.zip` mit `context`, `risk`, `capabilities`, `actions`, `signals`, `metrics`, `label-policies`, `taxonomy`, `snapshot.go`, `STANDARD.md` und `integration-plan.md`.

---

## 1. Kurzfazit

Die aktuelle Registry-Basis ist gut und zeigt bereits die richtige Architektur:

- Registries sind von Contracts getrennt.
- Forge soll Registry-Version/Digest pinnen.
- KLIQ soll im Managed Mode unbekannte Runtime-Semantik ablehnen.
- Kontext, Risiko, Capabilities, Actions, Signals, Metrics und Label-Policies sind bereits separat modelliert.
- Das Standard-Dokument formuliert wichtige Nicht-Ziele sauber: keine Kundenpolicy, keine Vendor-Konfiguration, keine durable Runtime-Learnings ohne Approval.

Für ein produktiv belastbares Kernloom-Standard-Repository fehlen aber noch mehrere Bausteine:

1. **Explizite Granularity-Registry** für Semantic Downgrade / Granularity Gap.
2. **Saubereres Runtime-Action-Modell**: aktuell gibt es Action-Level, aber zu wenig konkrete Action Contracts.
3. **Getrennte Scope-Semantik**: Entity Scope und Visibility Scope sind aktuell teilweise vermischt.
4. **Mehr Signal-Semantik**: Signals haben IDs und TTL, aber kaum Evidenz-, Confidence-, Source- und Enforcement-Semantik.
5. **Kontext-Snapshot für KLIQ**: `snapshot.go` übernimmt aktuell nur `ContextVersion`, aber nicht die Context Keys selbst.
6. **Registry-Schemas und Linter**: Es fehlen maschinenprüfbare JSON/YAML Schemas und CI-Checks.
7. **Naming-Konsistenz**: z. B. `signals.http.*` innerhalb der Signal Registry sollte bereinigt werden.
8. **Release-Hygiene**: `.git` sollte nicht im Release-ZIP landen; `replace ../kernloom-contracts` sollte nicht in veröffentlichter Form bleiben.

---

## 2. Zielbild

Das Registry-Repository soll die normative Semantik für Kernloom liefern.

Es soll nicht nur sagen:

```text
Diese ID existiert.
```

Sondern auch:

```text
Diese ID bedeutet X.
Diese ID darf in diesem Scope verwendet werden.
Diese ID darf nur mit dieser Qualität / TTL / Granularität / Wirkung verwendet werden.
Diese ID kann im Runtime Path verwendet werden oder nur im Config Path.
Diese ID darf von KLIQ lokal ausgeführt werden oder braucht Correlate / Forge / Approval.
```

Das Ziel ist ein Repository, gegen das Forge, KLIQ, Correlate und Adapter deterministisch validieren können.

---

## 3. Zielstruktur des Repositories

Vorgeschlagene Zielstruktur:

```text
kernloom-registries/
  README.md
  STANDARD.md
  CHANGELOG.md
  registry.lock.yaml
  go.mod
  snapshot.go

  registries/
    context/
      canonical-keys.yaml
    risk/
      risk-taxonomy.yaml
    capabilities/
      canonical-capabilities.yaml
    actions/
      runtime-actions.yaml
      runtime-action-contracts.yaml
    signals/
      canonical-signals.yaml
    metrics/
      canonical-metrics.yaml
      label-policies.yaml
    taxonomy/
      entities.yaml
    granularity/
      canonical-granularities.yaml
    scopes/
      canonical-scopes.yaml
    mappings/
      gap-taxonomy.yaml
      support-taxonomy.yaml

  schemas/
    context-key-registry.schema.json
    risk-taxonomy.schema.json
    capability-registry.schema.json
    runtime-action-registry.schema.json
    runtime-action-contracts.schema.json
    signal-registry.schema.json
    metric-registry.schema.json
    label-policy-registry.schema.json
    taxonomy-registry.schema.json
    granularity-registry.schema.json
    scope-registry.schema.json
    registry-snapshot.schema.json

  cmd/
    kernloom-registry-lint/
      main.go

  tests/
    fixtures/
      valid/
      invalid/
    golden/
      snapshot.json

  docs/
    integration-plan.md
    registry-authoring-guide.md
    compatibility-policy.md
```

---

## 4. Priorisierte Umsetzung

## P0 – Sofortige Hygiene und Stabilisierung

Diese Punkte sollten vor der nächsten Verwendung als Standard-Artefakt umgesetzt werden.

### 4.1 `.git` aus ZIP / Release-Artefakten entfernen

**Problem:**  
Das aktuelle ZIP enthält `.git/` inklusive Hooks und Git-Konfiguration. Das ist für ein Release-Artefakt unnötig und kann verwirren.

**Änderung:**

- Release-ZIP ohne `.git/` erzeugen.
- Optional zusätzlich Source-Archiv über GitHub/Gitea Release bereitstellen.

**Akzeptanzkriterium:**

```text
unzip -l kernloom-registries.zip | grep '.git/'
```

muss leer zurückkommen.

---

### 4.2 `go.mod replace` produktionsfähig machen

**Problem:**

Aktuell enthält `go.mod`:

```go
replace github.com/kernloom/kernloom-contracts => ../kernloom-contracts
```

Das ist für lokale Entwicklung okay, aber nicht für ein versioniertes Release.

**Empfehlung:**

Für Entwicklung:

```text
go.work
```

Für Releases:

```go
require github.com/kernloom/kernloom-contracts v0.1.0
```

**Akzeptanzkriterium:**

- Release-Branch enthält kein lokales `replace`.
- Lokale Entwicklung nutzt `go.work` oder ein separates `go.mod.dev`-Pattern.

---

### 4.3 Einheitliches Naming für Signals

**Problem:**

In `registries/signals/canonical-signals.yaml` gibt es IDs wie:

```yaml
- id: source.pps_high
- id: graph.new_edge_after_freeze
- id: signals.http.credential_stuffing_suspected
```

Das Prefix `signals.` ist innerhalb der Signal Registry redundant und inkonsistent.

**Änderung:**

Vorher:

```yaml
- id: signals.http.credential_stuffing_suspected
- id: signals.http.path_scan_suspected
- id: signals.http.status_5xx_source_pressure
- id: signals.http.auth_fail_rate_high
```

Nachher:

```yaml
- id: http.credential_stuffing_suspected
- id: http.path_scan_suspected
- id: http.status_5xx_source_pressure
- id: http.auth_fail_rate_high
```

**Akzeptanzkriterium:**

- Keine Signal-ID beginnt mit `signals.`.
- Deprecation Mapping ist vorhanden, falls alte IDs schon genutzt wurden:

```yaml
- id: signals.http.credential_stuffing_suspected
  deprecated: true
  replacedBy: http.credential_stuffing_suspected
```

Wenn noch keine Abhängigkeiten bestehen, direkt umbenennen ohne Deprecated Alias.

---

### 4.4 Registry-Linter einführen

**Problem:**

Aktuell gibt es keine sichtbaren CI-Regeln, die Registry-Konsistenz prüfen.

**Neue Komponente:**

```text
cmd/kernloom-registry-lint/main.go
```

**Linter muss prüfen:**

```text
- jede Registry-Datei ist gültiges YAML
- apiVersion ist erlaubt
- kind passt zum erwarteten Schema
- metadata.name ist gesetzt
- metadata.version ist semver-kompatibel
- alle IDs sind eindeutig innerhalb einer Registry
- IDs entsprechen Naming-Konvention
- Cross-References existieren
- deprecated Einträge haben reason und optional replacedBy
- keine unbekannten Scope-Werte
- keine unbekannten Severity-Werte
- keine verbotenen Labels als Metric Label
- keine Runtime Capability ohne gültigen Action Contract
- keine restrictive Runtime Action ohne TTL-Anforderung
```

**Akzeptanzkriterium:**

```bash
go test ./...
go run ./cmd/kernloom-registry-lint ./registries
```

muss erfolgreich sein.

---

## P1 – Semantik vervollständigen

Diese Punkte machen die Registries deutlich produktionsfähiger.

---

## 5. Neue Granularity Registry

### 5.1 Problem

Kernloom braucht eine klare Sprache für Granularity Gaps und Semantic Downgrades.

Beispiel:

```text
Policy meint workload_identity.
Adapter kann nur src_ip.
=> Granularity Gap und Semantic Downgrade.
```

Aktuell gibt es in `taxonomy/entities.yaml` Support- und Fidelity-Level, aber keine explizite Granularity-Definition.

### 5.2 Neue Datei

```text
registries/granularity/canonical-granularities.yaml
```

### 5.3 Vorschlag

```yaml
apiVersion: kernloom.io/registry/v1alpha1
kind: GranularityRegistry
metadata:
  name: kernloom-canonical-granularities
  version: "0.1.0"
spec:
  semanticLevels:
    - id: network
      order: 10
    - id: transport
      order: 20
    - id: session
      order: 30
    - id: identity
      order: 40
    - id: workload
      order: 50
    - id: service
      order: 60
    - id: application
      order: 70

  granularities:
    - id: ip
      semanticLevel: network
      entityScope: src_ip
      description: Single IP address. No identity semantics.

    - id: ip_port
      semanticLevel: transport
      entityScope: endpoint
      description: IP and port tuple. No user or workload identity.

    - id: five_tuple
      semanticLevel: transport
      entityScope: flow
      description: Source IP, source port, destination IP, destination port and protocol.

    - id: user_identity
      semanticLevel: identity
      entityScope: subject
      description: Human or external partner identity.

    - id: device_identity
      semanticLevel: identity
      entityScope: device
      description: Managed or attested device identity.

    - id: workload_identity
      semanticLevel: workload
      entityScope: workload
      description: Workload, service account, SPIFFE ID, Kubernetes workload or equivalent.

    - id: service_identity
      semanticLevel: service
      entityScope: resource
      description: Logical service identity independent of IP address.

    - id: application_segment
      semanticLevel: application
      entityScope: resource
      description: Application segment or application group as represented by a vendor platform.

  gapTypes:
    - id: granularity_gap
      description: Target adapter supports a weaker granularity than the policy requires.

    - id: semantic_downgrade
      description: Policy meaning is preserved only partially because identity/context is reduced to a lower-level representation.

    - id: enforcement_gap
      description: Required enforcement effect is not available on the selected adapter.

    - id: context_gap
      description: Required context fact is missing, stale or unsupported.
```

### 5.4 Anpassung in Adapter-Manifests

Adapter sollen später deklarieren:

```yaml
provides:
  granularities:
    subject: [ip, five_tuple]
    resource: [ip_port]
```

oder:

```yaml
provides:
  granularities:
    subject: [user_identity]
    resource: [application_segment]
```

### 5.5 Akzeptanzkriterien

- Forge kann pro Policy Required Granularity gegen Adapter Granularity vergleichen.
- Reports können `granularity_gap`, `semantic_downgrade`, `enforcement_gap`, `context_gap` standardisiert ausgeben.
- KLIQ kann Runtime Packs ablehnen, wenn Required Granularity nicht durch lokalen Adapter gedeckt ist und keine explizite Downgrade-Erlaubnis vorliegt.

---

## 6. Scope-Modell auftrennen

### 6.1 Problem

Aktuell werden Scopes vermischt:

In Metrics:

```yaml
allowedScopes: [src_ip, node, global]
```

In Signals:

```yaml
allowedScopes: [local, site, tenant]
```

Das sind unterschiedliche Dimensionen:

```text
src_ip, node, service = worauf bezieht sich das Signal / die Metrik?
local, site, tenant, global = wie weit reicht die Sichtbarkeit / Aussagekraft?
```

### 6.2 Neue Datei

```text
registries/scopes/canonical-scopes.yaml
```

### 6.3 Vorschlag

```yaml
apiVersion: kernloom.io/registry/v1alpha1
kind: ScopeRegistry
metadata:
  name: kernloom-canonical-scopes
  version: "0.1.0"
spec:
  entityScopes:
    - id: src_ip
      entityType: network
    - id: dst_ip
      entityType: network
    - id: endpoint
      entityType: network
    - id: five_tuple
      entityType: communication_edge
    - id: node
      entityType: component
    - id: service
      entityType: resource
    - id: workload
      entityType: workload
    - id: subject
      entityType: subject
    - id: device
      entityType: device
    - id: session
      entityType: session
    - id: communication_edge
      entityType: communication_edge

  visibilityScopes:
    - id: local
      description: Single node or local KLIQ view.
    - id: site
      description: Site, cluster or local environment view.
    - id: tenant
      description: Tenant-wide view.
    - id: global
      description: Cross-tenant or global view. Usually not allowed for customer policy decisions.
```

### 6.4 Anpassung in Metrics

Vorher:

```yaml
allowedScopes: [src_ip, node, global]
```

Nachher:

```yaml
entityScopes: [src_ip, node]
visibilityScopes: [local, site]
```

### 6.5 Anpassung in Signals

Vorher:

```yaml
allowedScopes: [local, site, tenant]
```

Nachher:

```yaml
entityScopes: [src_ip]
visibilityScopes: [local, site, tenant]
```

### 6.6 Akzeptanzkriterien

- Keine Registry verwendet mehr `allowedScopes` ohne klare Bedeutung.
- Metrics und Signals unterscheiden `entityScopes` und `visibilityScopes`.
- Linter lehnt unbekannte Scope-Werte ab.

---

## 7. Runtime Actions verbessern

### 7.1 Problem

`registries/actions/runtime-actions.yaml` definiert aktuell Level:

```yaml
levels:
  - id: observe
  - id: soft
  - id: hard
  - id: block
```

Das ist nützlich, aber noch kein vollständiger Action Contract.

Es fehlt pro konkreter Runtime Action:

```text
- capability ID
- action level
- effect
- monotonicity
- requires TTL?
- requires lease?
- requires audit?
- reversible?
- auto-revert strategy
- allowed decision source
- allowed runtime owner
- can grant access?
- default max TTL
- max TTL
```

### 7.2 Neue Datei

```text
registries/actions/runtime-action-contracts.yaml
```

### 7.3 Vorschlag

```yaml
apiVersion: kernloom.io/registry/v1alpha1
kind: RuntimeActionContractRegistry
metadata:
  name: kernloom-runtime-action-contracts
  version: "0.1.0"
spec:
  contracts:
    - id: enforce.traffic.rate_limit
      level: soft
      effect: restrictive
      monotonicity: restrict_only
      runtimeAllowed: true
      requiresTTL: true
      defaultTTL: 10m
      maxTTL: 1h
      requiresLease: true
      requiresAudit: true
      autoRevert: required
      reversible: true
      allowedDecisionSources: [kliq, correlate]
      requiredConfidence: medium
      canGrantAccess: false

    - id: enforce.traffic.connection_limit
      level: soft
      effect: restrictive
      monotonicity: restrict_only
      runtimeAllowed: true
      requiresTTL: true
      defaultTTL: 10m
      maxTTL: 1h
      requiresLease: true
      requiresAudit: true
      autoRevert: required
      reversible: true
      allowedDecisionSources: [kliq, correlate]
      requiredConfidence: medium
      canGrantAccess: false

    - id: enforce.traffic.drop
      level: block
      effect: restrictive
      monotonicity: restrict_only
      runtimeAllowed: true
      requiresTTL: true
      defaultTTL: 5m
      maxTTL: 30m
      requiresLease: true
      requiresAudit: true
      autoRevert: required
      reversible: true
      allowedDecisionSources: [correlate, kliq]
      requiredConfidence: high
      canGrantAccess: false

    - id: enforce.network.quarantine
      level: block
      effect: restrictive
      monotonicity: restrict_only
      runtimeAllowed: true
      requiresTTL: true
      defaultTTL: 10m
      maxTTL: 2h
      requiresLease: true
      requiresAudit: true
      autoRevert: required
      reversible: true
      allowedDecisionSources: [correlate]
      requiredConfidence: high
      canGrantAccess: false

    - id: enforce.access.allow
      level: observe
      effect: context_dependent
      monotonicity: may_grant_access
      runtimeAllowed: false
      configPathOnly: true
      requiresApproval: true
      canGrantAccess: true

    - id: enforce.access.deny
      level: block
      effect: restrictive
      monotonicity: restrict_only
      runtimeAllowed: true
      requiresTTL: true
      defaultTTL: 5m
      maxTTL: 30m
      requiresLease: true
      requiresAudit: true
      autoRevert: required
      reversible: true
      allowedDecisionSources: [correlate]
      requiredConfidence: high
      canGrantAccess: false
```

### 7.4 Anpassung in `canonical-capabilities.yaml`

Capabilities sollten weiterhin definieren, was ein Adapter kann. Runtime Action Contracts definieren, unter welchen Bedingungen es ausgeführt werden darf.

Beispiel Capability:

```yaml
- id: enforce.traffic.rate_limit
  category: enforce
  domain: traffic
  effect: restrictive
  runtimeAction: true
  severity: 1
  actionContract: enforce.traffic.rate_limit
```

### 7.5 Wichtige Designentscheidung

`allow` sollte im produktiven Managed Runtime Path standardmässig **nicht** erlaubt sein.

Begründung:

```text
Runtime darf kurzfristig einschränken oder wiederherstellen.
Runtime soll nicht dauerhaft oder stillschweigend Zugriff erweitern.
Allow / Grant gehört in den Config Path mit Review.
```

### 7.6 Akzeptanzkriterien

- Jede `runtimeAction: true` Capability hat einen Action Contract.
- Jede restrictive Runtime Action verlangt TTL, Lease, Audit und Auto-Revert.
- Jede Action mit `canGrantAccess: true` ist `runtimeAllowed: false`, ausser es gibt eine explizite Ausnahme mit Approval-Modell.
- KLIQ kann anhand des Snapshot prüfen, ob eine Action lokal erlaubt ist.

---

## 8. Capabilities präzisieren

### 8.1 Problem

Die Capability Registry ist gut als Start, aber noch zu knapp für produktive Delegation.

Aktuell fehlen pro Capability u. a.:

```text
- supported path: runtime/config/both
- required granularity
- reversible / irreversible
- local safe default
- adapter requirement category
- whether it is a grant, restriction, observation or assessment
```

### 8.2 Vorschlag für zusätzliche Felder

```yaml
- id: enforce.traffic.rate_limit
  category: enforce
  domain: traffic
  effect: restrictive
  runtimeAction: true
  severity: 1
  actionContract: enforce.traffic.rate_limit
  allowedPaths: [runtime]
  requiredGranularity:
    subject: [src_ip, workload_identity]
    resource: [ip_port, service_identity]
  reversible: true
  safeLocalDefault: true
```

Für `enforce.access.allow`:

```yaml
- id: enforce.access.allow
  category: enforce
  domain: access
  effect: grant
  runtimeAction: false
  allowedPaths: [config]
  requiresApproval: true
  reversible: true
  safeLocalDefault: false
```

### 8.3 Akzeptanzkriterien

- Jede Capability hat `allowedPaths`.
- Granting Capabilities sind im Runtime Path verboten oder require Approval.
- Jede Enforcement Capability referenziert eine Granularity oder erklärt, warum keine nötig ist.

---

## 9. Signals semantisch anreichern

### 9.1 Problem

Signals sind aktuell sehr schlank:

```yaml
- id: source.pps_high
  domain: source
  allowedScopes: [local, site]
  defaultTTL: 15m
```

Für produktive Policy-Entscheidungen braucht Forge/KLIQ mehr Informationen:

```text
- aus welchen Metrics entsteht das Signal?
- ist es lokal oder korreliert?
- darf es Enforcement auslösen?
- welche Confidence ist nötig?
- was ist der empfohlene Response-Typ?
- welche TTL gilt?
- welche Evidence muss mitgeliefert werden?
```

### 9.2 Vorschlag

```yaml
- id: source.pps_high
  domain: network
  entityScopes: [src_ip]
  visibilityScopes: [local, site]
  defaultTTL: 15m
  derivedFromMetrics:
    - network.packets_per_second
  evidence:
    required:
      - observed_value
      - baseline_upper_bound
      - observation_window
      - entity_scope
    optional:
      - packet_sample_hash
      - top_destination_ports
  confidence:
    minimumForAlert: low
    minimumForRateLimit: medium
    minimumForBlock: high
  enforcementEligible: true
  suggestedResponses:
    - enforce.traffic.rate_limit
    - enforce.traffic.drop
```

Für Graph-Signale:

```yaml
- id: graph.new_edge_after_freeze
  domain: graph
  entityScopes: [communication_edge]
  visibilityScopes: [local, site, tenant]
  defaultTTL: 1h
  evidence:
    required:
      - source_entity
      - target_entity
      - first_seen_at
      - graph_freeze_version
  enforcementEligible: conditional
  requiredContext:
    - communication.edge.status
  suggestedResponses:
    - observe.signal.emit
    - enforce.traffic.rate_limit
```

### 9.3 Akzeptanzkriterien

- Jedes Signal hat `entityScopes` und `visibilityScopes`.
- Jedes Signal deklariert, ob es enforcement-eligible ist.
- Wenn `enforcementEligible: true`, muss mindestens eine Suggested Response existieren.
- Wenn ein Signal von Metriken abgeleitet ist, müssen alle Metriken in `canonical-metrics.yaml` existieren.

---

## 10. Metrics und Label Policies verbessern

### 10.1 Problem

Metrics haben IDs, Unit, Scope und Cardinality Risk. Label Policies sind separat vorhanden. Es fehlt aber die Verbindung:

```text
Welche Labels darf diese konkrete Metric verwenden?
Welche Aggregationen sind erlaubt?
Welche Retention ist empfohlen?
Welche Metriken dürfen für Baselines genutzt werden?
```

### 10.2 Vorschlag Metric-Format

```yaml
- id: http.requests_per_second
  domain: http
  valueType: rate
  unit: requests/s
  entityScopes: [src_ip, service]
  visibilityScopes: [local, site]
  baselineAllowed: true
  highCardinalityRisk: low
  allowedLabels:
    - service
    - route_group
    - status_class
  forbiddenLabels:
    - path
    - uri
    - full_url
    - authorization
    - cookie
  aggregations:
    - sum
    - rate
    - p95
  defaultWindow: 1m
  retentionClass: operational
```

### 10.3 Neue Retention-Klassen

Optional in `label-policies.yaml` oder eigener Datei:

```yaml
retentionClasses:
  - id: ephemeral
    defaultRetention: 1h
  - id: operational
    defaultRetention: 14d
  - id: baseline
    defaultRetention: 90d
  - id: audit
    defaultRetention: 1y
```

### 10.4 Akzeptanzkriterien

- Linter prüft, dass `allowedLabels` in `label-policies.yaml` existieren und `allowed: true` sind.
- Linter verbietet Labels mit `allowed: false` pro Metric.
- Baseline-Metrics müssen `baselineAllowed: true` haben.
- Metrics mit `highCardinalityRisk: medium` oder höher brauchen explizite Normalisierung oder Aggregation.

---

## 11. Context Keys für Runtime Snapshot erweitern

### 11.1 Problem

In `snapshot.go` wird aktuell nur die Context Registry Version übernommen:

```go
ContextVersion: contextReg.Metadata.Version,
```

Wenn KLIQ im Managed Mode unbekannte Context Keys ablehnen soll, braucht es mindestens eine kompakte Runtime-Sicht auf die erlaubten Context Keys.

### 11.2 Vorschlag

Contracts sollten `RegistrySnapshot` um Context Keys erweitern:

```go
type RegistrySnapshot struct {
    Ref            RegistryRef
    ContextVersion string
    ContextKeys    []ContextKeyEntry
    RiskLevels     []RiskLevelEntry
    Capabilities   []CapabilityEntry
    ActionLevels   []ActionLevelEntry
    ActionContracts []RuntimeActionContractEntry
    Signals        []SignalEntry
    Metrics        []MetricEntry
    LabelPolicies  []LabelPolicyEntry
    Granularities  []GranularityEntry
    Scopes         ScopeRegistryView
}
```

Dann in `snapshot.go`:

```go
snap := contracts.RegistrySnapshot{
    Ref: contracts.RegistryRef{
        Name:     Name,
        Version:  Version,
        Revision: revision,
    },
    ContextVersion: contextReg.Metadata.Version,
    ContextKeys:     contextReg.Spec.Keys,
    RiskLevels:      riskReg.Spec.Levels,
    Capabilities:    capReg.Spec.Capabilities,
    ActionLevels:    actionReg.Spec.Levels,
    ActionContracts: actionContractReg.Spec.Contracts,
    Signals:         signalReg.Spec.Signals,
    Metrics:         metricReg.Spec.Metrics,
    LabelPolicies:   labelReg.Spec.Labels,
    Granularities:   granularityReg.Spec.Granularities,
    Scopes:          scopeReg.Spec,
}
```

### 11.3 Akzeptanzkriterien

- KLIQ kann `subject.risk.level` als bekannten Key erkennen.
- KLIQ kann `subject.magic.foo` im Managed Mode ablehnen.
- Snapshot-Digest ändert sich, wenn Context Keys geändert werden.

---

## 12. Risk Taxonomy verbessern

### 12.1 Problem

Risk Taxonomy ist gut als Anfang, aber für Runtime-Entscheidungen fehlen:

```text
- staleness semantics
- unknown handling
- source type
- normalization requirement
- minimum confidence/completeness by enforcement level
```

### 12.2 Vorschlag

```yaml
spec:
  scopes:
    - subject
    - device
    - workload
    - session
    - resource
    - application
    - communication_edge
    - network_source
    - target_platform
    - environment

  levels:
    - id: low
      scoreRange: [0, 29]
    - id: medium
      scoreRange: [30, 59]
    - id: high
      scoreRange: [60, 79]
    - id: critical
      scoreRange: [80, 100]
    - id: unknown
      scoreRange: null
      enforcementMeaning: insufficient_evidence

  quality:
    confidence:
      type: float
      range: [0.0, 1.0]
    completeness:
      type: float
      range: [0.0, 1.0]
    freshness:
      type: duration
      meaning: maximum accepted age of all critical inputs

  minimumQualityByActionLevel:
    observe:
      confidence: 0.0
      completeness: 0.0
    soft:
      confidence: 0.5
      completeness: 0.4
    hard:
      confidence: 0.7
      completeness: 0.6
    block:
      confidence: 0.85
      completeness: 0.7

  sourceTypes:
    - id: kernloom_native
    - id: vendor_native
    - id: normalized_vendor
    - id: manual_override
    - id: inherited_context
```

### 12.3 Akzeptanzkriterien

- High-impact enforcement kann gegen `minimumQualityByActionLevel` geprüft werden.
- Vendor-native Risk darf nicht direkt als Kernloom Risk verwendet werden ohne `sourceType: normalized_vendor` oder explizites Mapping.
- `unknown` wird nicht als `low` interpretiert.

---

## 13. Taxonomy erweitern

### 13.1 Problem

`taxonomy/entities.yaml` enthält bereits wichtige Kategorien. Für Gap Reports und Delegation fehlen aber noch konkrete Gap-Typen und Mapping-Klassen.

### 13.2 Vorschlag

Entweder in `taxonomy/entities.yaml` ergänzen oder besser auslagern nach:

```text
registries/mappings/gap-taxonomy.yaml
registries/mappings/support-taxonomy.yaml
```

### 13.3 Beispiel

```yaml
apiVersion: kernloom.io/registry/v1alpha1
kind: GapTaxonomy
metadata:
  name: kernloom-gap-taxonomy
  version: "0.1.0"
spec:
  gapTypes:
    - id: granularity_gap
      severityDefault: medium
    - id: semantic_downgrade
      severityDefault: high
    - id: enforcement_gap
      severityDefault: high
    - id: context_gap
      severityDefault: medium
    - id: fidelity_gap
      severityDefault: medium
    - id: delegation_gap
      severityDefault: medium
    - id: observability_gap
      severityDefault: medium
    - id: revert_gap
      severityDefault: high
    - id: audit_gap
      severityDefault: high

  supportLevels:
    - id: full
    - id: partial
    - id: conditional
    - id: full_vendor_native
    - id: not_supported

  fidelityLevels:
    - id: high
    - id: medium
    - id: low
    - id: none
```

### 13.4 Akzeptanzkriterien

- Forge Reports verwenden nur standardisierte Gap IDs.
- Adapter-Mappings verwenden nur standardisierte Support/Fidelity/Delegation IDs.
- Linter prüft diese Werte.

---

## 14. Registry-Schemas ergänzen

### 14.1 Problem

Ohne JSON/YAML Schemas werden Änderungen zu wenig kontrolliert.

### 14.2 Neue Schemas

Mindestens:

```text
schemas/context-key-registry.schema.json
schemas/risk-taxonomy.schema.json
schemas/capability-registry.schema.json
schemas/runtime-action-registry.schema.json
schemas/runtime-action-contracts.schema.json
schemas/signal-registry.schema.json
schemas/metric-registry.schema.json
schemas/label-policy-registry.schema.json
schemas/taxonomy-registry.schema.json
schemas/granularity-registry.schema.json
schemas/scope-registry.schema.json
```

### 14.3 Wichtige Schema-Regeln

```text
- id muss regex erfüllen: ^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$
- version muss semver-kompatibel sein
- defaultTTL muss Duration-Format erfüllen
- severity ist 0..3
- runtimeAction ist boolean
- effect ist enum: restrictive, grant, observe, assess, context_dependent
- allowedPaths ist enum-list: runtime, config, report, proposal
- deprecated braucht reason
- replacedBy muss existierende ID referenzieren
```

### 14.4 Akzeptanzkriterien

- Jede Registry-Datei referenziert oder wird gegen ein Schema geprüft.
- CI schlägt fehl bei ungültigen IDs, unbekannten Feldern oder falschen Typen.

---

## 15. Snapshot und Digest stabilisieren

### 15.1 Problem

`DigestSnapshot` nutzt `json.Marshal`. Das kann ausreichend sein, wenn die Structs stabil sind. Langfristig sollte der Digest aber klar kanonisch und reproduzierbar sein.

### 15.2 Vorschlag

- Alle Snapshot-Listen vor Digest nach ID sortieren.
- Unbekannte Map-Ordering-Probleme vermeiden.
- Registry Digest über kanonisches JSON bilden.
- Optional zusätzlich Digest über rohe Registry-Dateien mit deterministischer File-Reihenfolge.

### 15.3 Vorschlag Code-Verhalten

```text
Load all registries
Validate all registries
Normalize ordering
Build snapshot
Marshal canonical JSON
Hash SHA-256
```

### 15.4 Akzeptanzkriterien

- Zwei Builds aus gleichem Registry-Stand erzeugen denselben Digest.
- Änderung an Context Keys ändert Digest.
- Änderung an README ändert Digest nicht, sofern der Digest nur Registry-Semantik abbildet.

---

## 16. Registry Lockfile einführen

### 16.1 Neue Datei

```text
registry.lock.yaml
```

### 16.2 Zweck

Das Lockfile beschreibt, welche Registry-Dateien Teil des Standards sind und welche Digests sie haben.

### 16.3 Beispiel

```yaml
apiVersion: kernloom.io/registry/v1alpha1
kind: RegistryLock
metadata:
  name: kernloom-standard
  version: "0.1.0"
spec:
  files:
    - path: registries/context/canonical-keys.yaml
      digest: sha256:...
    - path: registries/risk/risk-taxonomy.yaml
      digest: sha256:...
    - path: registries/capabilities/canonical-capabilities.yaml
      digest: sha256:...
    - path: registries/actions/runtime-actions.yaml
      digest: sha256:...
    - path: registries/actions/runtime-action-contracts.yaml
      digest: sha256:...
    - path: registries/signals/canonical-signals.yaml
      digest: sha256:...
    - path: registries/metrics/canonical-metrics.yaml
      digest: sha256:...
    - path: registries/metrics/label-policies.yaml
      digest: sha256:...
    - path: registries/taxonomy/entities.yaml
      digest: sha256:...
    - path: registries/granularity/canonical-granularities.yaml
      digest: sha256:...
    - path: registries/scopes/canonical-scopes.yaml
      digest: sha256:...
```

### 16.4 Akzeptanzkriterien

- `kernloom-registry-lint` kann Lockfile verifizieren.
- Forge kann Lockfile pinnen.
- KLIQ kann Snapshot-Digest gegen Bundle-Metadata prüfen.

---

## 17. Konkrete Dateiänderungen nach Priorität

## PR 1 – Release Hygiene und Naming

**Ziel:** Repository sauber paketierbar machen.

Änderungen:

```text
- .git aus ZIP entfernen
- go.mod replace entfernen oder go.work einführen
- signals.http.* umbenennen zu http.*
- README um Release-Artefakt-Regeln ergänzen
- CHANGELOG.md hinzufügen
```

Akzeptanz:

```text
- go test ./... läuft
- kein .git im Release-ZIP
- keine Signal-ID beginnt mit signals.
```

---

## PR 2 – Linter und Schemas Skeleton

**Ziel:** Registry-Änderungen maschinell absichern.

Änderungen:

```text
- cmd/kernloom-registry-lint/main.go
- schemas/*.schema.json Skeletons
- tests/fixtures/valid
- tests/fixtures/invalid
- CI Workflow
```

Akzeptanz:

```text
- doppelte IDs werden erkannt
- ungültige ID-Namen werden erkannt
- unbekannte Kind-Werte werden erkannt
- fehlende metadata.version wird erkannt
```

---

## PR 3 – Scopes und Granularity

**Ziel:** Basis für Granularity Gap und Semantic Downgrade schaffen.

Änderungen:

```text
- registries/scopes/canonical-scopes.yaml
- registries/granularity/canonical-granularities.yaml
- Metrics: allowedScopes -> entityScopes + visibilityScopes
- Signals: allowedScopes -> entityScopes + visibilityScopes
- snapshot.go lädt Scopes und Granularities
```

Akzeptanz:

```text
- Linter prüft Scope-Werte
- Forge kann Required Granularity gegen Adapter Granularity vergleichen
- Snapshot enthält Granularities und Scopes
```

---

## PR 4 – Runtime Action Contracts

**Ziel:** Runtime Enforcement sicher und begrenzt beschreiben.

Änderungen:

```text
- registries/actions/runtime-action-contracts.yaml
- capabilities referenzieren actionContract
- snapshot.go lädt Action Contracts
- Contracts-Modul wird erweitert
```

Akzeptanz:

```text
- jede runtimeAction Capability hat actionContract
- block/drop/quarantine brauchen TTL, Lease, Audit, Auto-Revert
- allow/grant ist nicht runtimeAllowed
```

---

## PR 5 – Signal Evidence Semantics

**Ziel:** Signals produktionsfähig für Risiko- und Enforcement-Entscheidungen machen.

Änderungen:

```text
- canonical-signals.yaml erweitert um derivedFromMetrics, evidence, confidence, enforcementEligible, suggestedResponses
- Cross-Reference Checks zu Metrics und Runtime Actions
```

Akzeptanz:

```text
- enforcementEligible=true verlangt suggestedResponses
- derivedFromMetrics müssen existieren
- suggestedResponses müssen Capability/Action Contract referenzieren
```

---

## PR 6 – Context Keys im Runtime Snapshot

**Ziel:** KLIQ kann unbekannte Context Keys im Managed Mode ablehnen.

Änderungen:

```text
- contracts.RegistrySnapshot erweitert um ContextKeys
- snapshot.go füllt ContextKeys
- Digest berücksichtigt ContextKeys
- KLIQ Managed Mode validiert Context Keys gegen Snapshot
```

Akzeptanz:

```text
- subject.risk.level wird akzeptiert
- subject.magic.foo wird abgelehnt
- Änderung an canonical-keys.yaml ändert Snapshot Digest
```

---

## PR 7 – Gap / Mapping Taxonomy

**Ziel:** Reports zwischen Forge, KLIQ, Correlate und Adaptern einheitlich machen.

Änderungen:

```text
- registries/mappings/gap-taxonomy.yaml
- registries/mappings/support-taxonomy.yaml
- taxonomy/entities.yaml ggf. verschlanken
```

Akzeptanz:

```text
- Forge Reports verwenden nur standardisierte Gap IDs
- Adapter Manifests verwenden nur standardisierte Support/Fidelity/Delegation IDs
```

---

## 18. Beispiel: überarbeitete Signal Registry

Aktuell:

```yaml
- id: source.pps_high
  domain: source
  allowedScopes: [local, site]
  defaultTTL: 15m
```

Vorschlag:

```yaml
- id: source.pps_high
  domain: network
  entityScopes: [src_ip]
  visibilityScopes: [local, site]
  defaultTTL: 15m
  derivedFromMetrics:
    - network.packets_per_second
  evidence:
    required:
      - observed_value
      - baseline_upper_bound
      - observation_window
      - entity_scope
    optional:
      - top_destination_ports
  confidence:
    minimumForAlert: low
    minimumForRateLimit: medium
    minimumForBlock: high
  enforcementEligible: true
  suggestedResponses:
    - enforce.traffic.rate_limit
    - enforce.traffic.drop
```

---

## 19. Beispiel: überarbeitete Metric Registry

Aktuell:

```yaml
- id: http.requests_per_second
  domain: http
  valueType: rate
  unit: requests/s
  allowedScopes: [src_ip, service]
  baselineAllowed: true
  highCardinalityRisk: low
```

Vorschlag:

```yaml
- id: http.requests_per_second
  domain: http
  valueType: rate
  unit: requests/s
  entityScopes: [src_ip, service]
  visibilityScopes: [local, site]
  baselineAllowed: true
  highCardinalityRisk: low
  allowedLabels:
    - service
    - route_group
    - status_class
  forbiddenLabels:
    - path
    - uri
    - full_url
    - authorization
    - cookie
  aggregations:
    - sum
    - rate
  defaultWindow: 1m
  retentionClass: operational
```

---

## 20. Beispiel: überarbeitete Capability Registry

Aktuell:

```yaml
- id: enforce.traffic.drop
  category: enforce
  domain: traffic
  effect: restrictive
  runtimeAction: true
  severity: 3
```

Vorschlag:

```yaml
- id: enforce.traffic.drop
  category: enforce
  domain: traffic
  effect: restrictive
  runtimeAction: true
  severity: 3
  actionContract: enforce.traffic.drop
  allowedPaths: [runtime]
  requiredGranularity:
    subject: [src_ip, workload_identity]
    resource: [ip_port, service_identity]
  reversible: true
  safeLocalDefault: false
```

---

## 21. Beispiel: Runtime Action Contract

```yaml
- id: enforce.traffic.drop
  level: block
  effect: restrictive
  monotonicity: restrict_only
  runtimeAllowed: true
  requiresTTL: true
  defaultTTL: 5m
  maxTTL: 30m
  requiresLease: true
  requiresAudit: true
  autoRevert: required
  reversible: true
  allowedDecisionSources: [correlate, kliq]
  requiredConfidence: high
  canGrantAccess: false
```

---

## 22. Beispiel: Granularity Gap Report auf Registry-Basis

Wenn eine Policy verlangt:

```yaml
requiredGranularity:
  subject: workload_identity
  resource: service_identity
```

Und ein Adapter meldet:

```yaml
provides:
  granularities:
    subject: [ip]
    resource: [ip_port]
```

Dann muss Forge standardisiert erzeugen:

```yaml
gaps:
  - type: granularity_gap
    severity: high
    required:
      subject: workload_identity
      resource: service_identity
    provided:
      subject: ip
      resource: ip_port
    impact: workload identity is downgraded to IP-based enforcement

  - type: semantic_downgrade
    severity: high
    from: workload_identity
    to: ip
    approvalRequired: true
```

---

## 23. Was explizit nicht in dieses Registry-Repo gehört

Nicht aufnehmen:

```text
- konkrete Kundenpolicies
- konkrete Zscaler / OpenZiti / NGINX Zielkonfiguration
- konkrete Kundengruppen wie firma-admins
- freie Menschensprache / Natural Policy Model
- LLM-Prompts
- Runtime-gelernte Graph Edges als dauerhafte Policy
- Kunden-CMDB-Daten
- Secrets oder API Keys
```

Gehört stattdessen in andere Repos:

```text
kernloom-contracts:
  Wire Schemas, RuntimeBundle, RuntimeDecision, RegistrySnapshot Structs

kernloom-forge:
  Compiler, Validator, Coverage Reports, Gap Reports

kernloom-kliq:
  Runtime PDP, local enforcement, managed snapshot validation

adapter repos:
  Adapter manifests, vendor-native mappings, downgrade declarations

customer policy repos:
  konkrete Policy Intents und Zielsystem-Konfigurationen
```

---

## 24. Definition of Done für `kernloom-registries` v0.2.0

Eine Version `0.2.0` wäre aus meiner Sicht produktiv sinnvoll, wenn folgende Punkte erfüllt sind:

```text
[ ] Release-ZIP enthält keine .git-Dateien.
[ ] go.mod enthält kein lokales replace im Release-Zweig.
[ ] Alle Registry-Dateien haben JSON/YAML Schemas.
[ ] Registry-Linter läuft in CI.
[ ] Duplicate IDs werden verhindert.
[ ] ID-Naming ist konsistent.
[ ] Signal-IDs verwenden kein redundantes signals.* Prefix.
[ ] Granularity Registry existiert.
[ ] Scope Registry existiert.
[ ] Metrics und Signals trennen entityScopes und visibilityScopes.
[ ] Runtime Action Contracts existieren.
[ ] Runtime Capabilities referenzieren Action Contracts.
[ ] Restrictive Runtime Actions benötigen TTL, Lease, Audit und Auto-Revert.
[ ] Granting Actions sind im Runtime Path standardmässig verboten.
[ ] Signals deklarieren Evidence, Confidence und Enforcement Eligibility.
[ ] Metrics deklarieren allowedLabels und Retention/Aggregation.
[ ] Risk Taxonomy definiert Minimum Quality pro Action Level.
[ ] Snapshot enthält Context Keys, Granularities, Scopes und Action Contracts.
[ ] Snapshot Digest ist reproduzierbar.
[ ] Registry Lockfile existiert.
[ ] Forge kann Registry laden und validieren.
[ ] KLIQ kann Managed Mode gegen Snapshot prüfen.
```

---

## 25. Empfohlene Reihenfolge für Umsetzung durch eine KI / Entwicklerteam

1. **Repo reinigen und Release-Hygiene herstellen.**
2. **Linter Skeleton bauen.**
3. **Schemas für bestehende Registries ergänzen.**
4. **Signal-Naming normalisieren.**
5. **Scopes Registry einführen.**
6. **Granularity Registry einführen.**
7. **Metrics und Signals auf entityScopes / visibilityScopes migrieren.**
8. **Runtime Action Contracts einführen.**
9. **Capabilities mit Action Contracts und allowedPaths anreichern.**
10. **Signals mit Evidence / Confidence / Enforcement Eligibility anreichern.**
11. **Metrics mit allowedLabels / Retention / Aggregation anreichern.**
12. **Contracts RegistrySnapshot erweitern.**
13. **snapshot.go erweitern.**
14. **Golden Snapshot Test erstellen.**
15. **Forge Loader und KLIQ Managed Mode auf neue Snapshot-Struktur anpassen.**

---

## 26. Wichtigste Architekturentscheidung

Die Registries sollen nicht nur technische IDs sammeln, sondern **Sicherheitssemantik erzwingen**.

Deshalb gilt:

```text
Capability sagt: Ein Adapter kann X.
Action Contract sagt: X darf nur unter diesen Bedingungen runtime ausgeführt werden.
Granularity sagt: Auf welcher semantischen Tiefe gilt X.
Scope sagt: Worauf und mit welcher Sichtweite gilt X.
Signal sagt: Welche Evidenz liegt vor.
Risk sagt: Wie vertrauenswürdig und vollständig die Bewertung ist.
Forge sagt: Diese Kombination ist gültig oder nicht.
KLIQ sagt: Ich akzeptiere nur, was im Snapshot erlaubt ist.
```

Das ist der Kern, damit Kernloom später nicht einfach ein weiteres Policy-Tool wird, sondern ein echtes, capability-aware Zero-Trust Control-Plane-Modell.
