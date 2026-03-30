# Specs Index

## Migration Specs

- **[terraform-removal-migration.md](terraform-removal-migration.md)** — Master design spec for removing the Terraform layer. Covers the full process: Phase 0 infrastructure, per-service migration cycle, batch ordering, plan skill design, executor adaptations, agent verification, cutover process, and final TF stack removal. Start here.

## Catalogs (to be created in Phase 0)

These catalogs will be generated as Phase 0 infrastructure tickets:

- `external-name-catalog.json` — Per-resource external name strategies (extracted from `config/externalname.go`)
- `tf-business-logic-catalog.md` — Resources with TerraformConfigurationInjector/TerraformCustomDiff and native equivalents
- `move-to-status-catalog.json` — Fields moved from spec to status per resource
- `connection-details-catalog.json` — Connection detail keys and sources per resource
- `native-controller-pattern.md` — Comprehensive guide for building native controllers (reference for executor)
