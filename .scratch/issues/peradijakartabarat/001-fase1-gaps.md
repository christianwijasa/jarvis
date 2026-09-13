---
status: open
created: 2026-09-13
plan: docs/plans/peradijakartabarat/
repos: peradijakartabarat
---

# Complete Fase 1 High-Priority Gaps

## Problem

Kantong Magang Advokat (peradijakartabarat) has core internship workflow built but is missing key features from Sept 2026 deck "Digitalisasi Bursa Magang Advokat":

1. **Firm self-registration** — Law Firms currently DPC-provisioned only
2. **Firm UI: Create Lowongan** — Firms can't post their own internship openings
3. **File upload/download** — No real document handling (logbook attachments, berkas sumpah)
4. **Admin audit actions + export** — Missing structured audit workflow and Pengadilan Tinggi export

## Goal

Close these four gaps to reach Fase 1 completeness per deck vision.

## Notes

- See [design.md](../../docs/plans/peradijakartabarat/design.md) for full gap backlog and medium-priority items
- Verify Pengadilan Tinggi export format with DPC stakeholder before building
- Document file storage strategy (local/S3/Laravel Filesystem) before implementing uploads

## Links

- Origin: https://cursor.com/codebase/fchr/peradijakartabarat
- Plan: [docs/plans/peradijakartabarat/design.md](../../docs/plans/peradijakartabarat/design.md)
- ADR: [0001-origin-sot.md](../../docs/adr/peradijakartabarat/0001-origin-sot.md)
