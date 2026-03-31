# Specs Index

## Migration Specs

- **[terraform-removal-migration.md](terraform-removal-migration.md)** — Master design spec for removing the Terraform layer. Covers the full process: Phase 0 infrastructure, per-service migration cycle, batch ordering, plan skill design, executor adaptations, agent verification, cutover process, and final TF stack removal. Start here.
- **[native-controller-pattern.md](native-controller-pattern.md)** — Comprehensive implementation guide for building native controllers. Package map, file layout, controller setup, CRUD templates, helper libraries, external name strategies, common pitfalls, pre-commit checklist.

## Agent Documentation (`.agents/docs/`)

- **[../docs/architecture.md](../docs/architecture.md)** — Provider architecture, family structure, dual scope, resource lifecycle, authentication, code generation pipeline, generated file patterns
- **[../docs/build-system.md](../docs/build-system.md)** — Makefile targets, CI/CD workflows, tool versions, code generation flow, buildtagger, linting
- **[../docs/config-patterns.md](../docs/config-patterns.md)** — Resource configuration patterns: external names, references, late init, custom diff, connection details, overrides
- **[../docs/testing.md](../docs/testing.md)** — Unit tests, E2E (Uptest), CI pipeline, test patterns, test gaps
- **[../docs/native-controller-guide.md](../docs/native-controller-guide.md)** — Provider-template patterns, ExternalClient interface, migration file layout, CRUD implementation

## Catalogs (to be created in Phase 0)

These catalogs will be generated as Phase 0 infrastructure tickets:

- `external-name-catalog.json` — Per-resource external name strategies (extracted from `config/externalname.go`)
- `tf-business-logic-catalog.md` — Resources with TerraformConfigurationInjector/TerraformCustomDiff and native equivalents
- `move-to-status-catalog.json` — Fields moved from spec to status per resource
- `connection-details-catalog.json` — Connection detail keys and sources per resource
