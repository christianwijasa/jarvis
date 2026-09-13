# Plan: peradijakartabarat (Kantong Magang Advokat)

**Project:** Kantong Magang Advokat — DPC PERADI Jakarta Barat  
**Tech stack:** Laravel 11, MySQL  
**Repos:**
- Origin (primary): https://cursor.com/codebase/fchr/peradijakartabarat
- GitHub sibling: https://github.com/christianwijasa/peradijakartabarat

**Related ADRs:** [ADR-0001: Origin as Source of Truth for peradijakartabarat](../../adr/peradijakartabarat/0001-origin-sot.md)

---

## Problem

DPC PERADI Jakarta Barat needs a digital platform to manage the mandatory 24-month advocate internship program (magang) and prevent "magang fiktif" (fraudulent internship claims). The platform connects Calon Advokat (advocate candidates), Law Firms, Advokat Pendamping (supervising advocates), and DPC admins through a verified internship lifecycle.

**Source material:** Platform_Magang_Fase_1 mockup, Sept 2026 deck "Digitalisasi Bursa Magang Advokat"

## Scope: Fase 1

Current implementation covers the core internship workflow:

### Built (verified in codebase)
- **User roles:** Calon Advokat (self-register), Law Firm & Advokat Pendamping (DPC-provisioned), Admin DPC
- **Lowongan (internship postings):** Search/browse, application submission
- **Lamaran (applications):** Status tracking (submitted → reviewed → accepted/rejected)
- **Logbook:** Daily entry submission, monthly Rekap Bulanan with TTD (digital signature) from Advokat Pendamping
- **Verifikasi:** Dual verification system (daily + monthly) to prevent magang fiktif
- **Berkas tracking UI:** Monitor documentation preparation for oath package
- **Admin monitoring:** DPC dashboard for overseeing all internships
- **24-month progress tracker:** Timeline visualization
- **Quota reference:** Data model includes kuota concept

### Gaps vs Sept 2026 Deck
Priority backlog for completing Fase 1 vision:

#### High Priority
1. **Firm self-registration** — Allow Law Firms to register directly (currently DPC-provisioned only)
2. **Firm UI: Create Lowongan** — Law Firms create their own postings (currently admin-created)
3. **File upload/download** — Real document handling for logbook attachments, verification berkas, and oath package generation
4. **Admin audit actions + export** — Structured audit workflow and export to Pengadilan Tinggi (High Court)

#### Medium Priority
5. **Lamaran mid-status transitions** — Additional workflow states between submitted and final decision (e.g., "under review", "interview scheduled")
6. **Logbook resubmit after revisi** — Explicit resubmission flow when Advokat Pendamping requests revisions
7. **Notifications** — Email/in-app alerts for status changes, approvals, deadlines
8. **Hard quota enforcement** — Prevent Law Firms from exceeding kuota limits (currently soft reference)

#### Optional / Deferred
9. **Mobile app** — Responsive mobile interface (Phase 2 teaser in deck)

### Out of Scope (Fase 2)
- **PKPA admisi:** Special Professional Education integration
- **UPA admisi:** Professional Exam integration

See [CONTEXT.md](../../../CONTEXT.md) for domain vocabulary (PERADI, DPC, UPA, PKPA, Calon Advokat, etc.).

---

## Architecture Overview

**Laravel 11 MVC** with key entities:

- `LawFirm` — Legal practice hosting internships
- `AdvokatPendamping` — Supervising advocate (belongs to LawFirm)
- `CalonAdvokat` — Advocate candidate user account
- `Lowongan` — Internship posting
- `Lamaran` — Application linking CalonAdvokat → Lowongan
- `LogbookEntry` — Daily work journal entry
- `LogbookRekapBulanan` — Monthly summary + TTD approval
- `VerifikasiChecklist` — Dual verification state
- `BerkasSumpah` — Oath package documentation
- `AuditAkhir` — Final 24-month completion audit

**Core workflow:**
1. Law Firm posts Lowongan (via DPC admin in Fase 1)
2. Calon Advokat searches and applies (Lamaran)
3. Firm reviews and accepts
4. Daily Logbook entries + monthly Rekap with Advokat Pendamping TTD
5. Verification checklist tracking over 24 months
6. Final Audit Akhir by DPC
7. Berkas Sumpah package generation for oath ceremony

---

## Recommended Work Streams

### Stream 1: Close Fase 1 Gaps
**Goal:** Complete the Sept 2026 deck vision  
**Priorities:** File handling, firm self-reg, lowongan creation, admin audit export

**Why first:** These are documented requirements with mockups; highest confidence, highest delivery value.

### Stream 2: Workflow Polish
**Goal:** Improve user experience for existing flows  
**Focus:** Notification system, lamaran transitions, logbook resubmit UX, quota enforcement

**Why second:** Enhances core flows that are already built; lower risk than new features.

### Stream 3: Monitoring & Analytics
**Goal:** DPC admin visibility into program health  
**Focus:** Aggregate dashboards (completion rates, firm performance, magang fiktif detection signals)

**Why third:** Operational improvement after core workflows are solid.

---

## Integration Points

### Pengadilan Tinggi Export
Final audit data and Berkas Sumpah summaries must export to the High Court (mentioned in deck). Format TBD — confirm with DPC stakeholder before building.

### Future Fase 2 Hooks
- PKPA enrollment/certificate validation
- UPA exam results integration
- Mobile app API endpoints (if mobile is greenlit)

---

## Deployment Notes

- **Current hosting:** TBD (check repo for deployment scripts or .env.example)
- **DPC admin seed:** Ensure first admin account provisioning is documented
- **Data retention:** Logbook and verification records must persist 24+ months; confirm backup/archival policy

---

## Next Steps

1. **Verify GitHub/Origin sync strategy** — Clarify if both repos stay in sync or if Origin is canonical (see ADR-0001)
2. **Prioritize gap backlog** — Confirm High Priority items with DPC stakeholder
3. **Document file storage approach** — Local disk, S3, or Laravel Filesystem config for Berkas and Logbook attachments
4. **Define Pengadilan Tinggi export format** — Get sample or spec from DPC

---

**Last updated:** 2026-09-13  
**Status:** Fase 1 partially complete, gap backlog identified
