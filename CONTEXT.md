# Context — Shared Glossary

Domain vocabulary for projects tracked in this hub. Terms are listed alphabetically within sections.

## PERADI / Kantong Magang Advokat

### Advokat Pendamping
A supervising advocate (mentor) from a Law Firm who guides a Calon Advokat through their internship. Provisioned by DPC PERADI Jakarta Barat.

### Audit Akhir
Final audit conducted at the end of the 24-month internship period to verify completion requirements before recommending the Calon Advokat for oath ceremony.

### Berkas Sumpah
"Oath package" — the complete documentation bundle prepared after successful audit, required for the advocate's swearing-in ceremony.

### Bursa Magang
Alternative term for Kantong Magang Advokat. Refers to the internship exchange/marketplace platform. _Avoid:_ "job board" (too generic).

### Calon Advokat
"Advocate candidate" — a law graduate registered to complete the mandatory 24-month internship (magang) with a Law Firm before becoming a licensed advocate. Self-registers via `/register` in the platform.

### DPC (Dewan Pimpinan Cabang)
Branch Executive Board of PERADI. The DPC PERADI Jakarta Barat admin role provisions Law Firms and Advokat Pendamping accounts, monitors internships, and conducts final audits.

### Fase 1
Current implementation scope for Kantong Magang Advokat: internship posting, application, logbook tracking, dual verification, and audit workflow. Does not include PKPA or UPA admisi (Fase 2).

### Fase 2
Future scope (teased in Sept 2026 deck): PKPA and UPA admisi functionality. Out of scope for current graphify effort.

### Kantong Magang Advokat
"Advocate Internship Pool" — the digital platform for managing the 24-month mandatory advocate internship at DPC PERADI Jakarta Barat. Built on Laravel 11/MySQL from mockup Platform_Magang_Fase_1.

### Kuota
Quota for internship positions. Each Law Firm has a maximum number of Calon Advokat they can accept simultaneously. Referenced in deck gaps as needing hard enforcement.

### Lamaran
Internship application submitted by a Calon Advokat to a specific Lowongan. Tracked through statuses: submitted, reviewed, accepted, rejected, etc.

### Law Firm
Legal practice entity registered with DPC PERADI Jakarta Barat that offers internship positions. Provisioned by DPC admin (Fase 1); self-registration mentioned in gap analysis (future).

### Logbook
Daily work journal maintained by Calon Advokat during their internship. Entries are reviewed monthly by Advokat Pendamping. Platform supports submission, revision requests, and resubmission workflow (with gaps noted in resubmit-after-revisi flow).

### Logbook Rekap Bulanan
Monthly logbook summary requiring TTD (signature/approval) from the Advokat Pendamping. Part of the verification trail for the 24-month period.

### Lowongan
"Opening" or internship position posted by a Law Firm. Calon Advokat search and apply to Lowongan. Fase 1: DPC-created; gap analysis mentions firm-created UI for Fase 2+.

### Magang Fiktif
"Fake internship" — fraudulent claim of completing internship requirements. The platform's dual verification system (daily logbook + monthly TTD + final audit) is designed to prevent this.

### PERADI (Perhimpunan Advokat Indonesia)
Indonesian Advocates Association. National organization for licensed advocates; manages mandatory internship requirements through regional DPC branches.

### PKPA (Pendidikan Khusus Profesi Advokat)
Special Professional Education for Advocates. A prerequisite certification course before or during internship. Mentioned as Fase 2 scope.

### TTD (Tanda Tangan Digital)
Digital signature, specifically the monthly approval signature from Advokat Pendamping on Logbook Rekap Bulanan.

### UPA (Ujian Profesi Advokat)
Advocate Professional Exam. Another certification gate. Mentioned as Fase 2 scope alongside PKPA admisi.

### Verifikasi Checklist
Dual verification system: daily logbook verification + monthly rekap approval. Core anti-fraud mechanism in Fase 1 build.

---

**Last updated:** 2026-09-13  
**Projects:** peradijakartabarat
