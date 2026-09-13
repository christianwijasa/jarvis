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

## Warning-Item Parity Plan

Three ⚠️ gaps between Sept 2026 deck pitch and current implementation. Close these to make Fase 1 web demo-credible against the PPTX.

### 1. Firm Self-Registration with Admin Verification

**Deck promise:** Law Firm wajib mendaftar lalu Admin DPC verifikasi sebelum akses talent pool / buka lowongan (slides 3/4/6).

**Current state:** Law Firm accounts are DPC-provisioned only. No self-registration flow.

**Proposed approach:**
- Add public `/register/law-firm` route with registration form: firm name, address, license number, contact person, email
- New `LawFirmRegistration` model: `status` enum (pending, approved, rejected), `submitted_at`, `reviewed_at`, `reviewed_by` (admin user), `rejection_reason` (nullable text)
- On submit → status `pending`, email to DPC admin queue
- Admin dashboard: "Pending Firm Registrations" list with approve/reject actions
- Approve → provision `LawFirm` + initial `User` (admin role for firm), send credentials email
- Reject → log reason, email applicant with feedback
- Block unapproved firms from login/lowongan access (middleware check)

**Acceptance criteria:**
- ✅ Law Firm can self-register with complete profile form
- ✅ Registration enters `pending` state; firm cannot log in yet
- ✅ DPC admin sees pending list with approve/reject buttons
- ✅ Approve → firm receives login credentials and can post Lowongan
- ✅ Reject → firm sees reason and can resubmit
- ✅ Demo script: show firm registers → admin reviews → firm logs in and creates first Lowongan

**Routes/models:**
```php
// routes/web.php
Route::get('/register/law-firm', [LawFirmRegistrationController::class, 'create'])->name('law-firm.register');
Route::post('/register/law-firm', [LawFirmRegistrationController::class, 'store']);

// app/Models/LawFirmRegistration.php
enum RegistrationStatus: string {
    case Pending = 'pending';
    case Approved = 'approved';
    case Rejected = 'rejected';
}

// Admin routes (auth:admin middleware)
Route::prefix('admin')->group(function () {
    Route::get('firm-registrations', [AdminFirmRegistrationController::class, 'index']);
    Route::post('firm-registrations/{id}/approve', [AdminFirmRegistrationController::class, 'approve']);
    Route::post('firm-registrations/{id}/reject', [AdminFirmRegistrationController::class, 'reject']);
});
```

**Demo script delta:**
1. Open public homepage → "Daftar sebagai Law Firm" button
2. Fill form (Kantor Hukum ABC, license 12345, etc.) → Submit
3. Show "Pendaftaran Anda sedang diproses" confirmation page
4. Switch to DPC admin view → "1 Pending Registration"
5. Review firm details → Approve with comment
6. Firm receives email with login link → logs in → creates first Lowongan

---

### 2. E-Signature Credibility for Logbook TTD

**Deck promise:** Otorisasi digital / e-signature pendamping on logbook monthly rekap (slide 6 tahap 2: "tanda tangan digital pendamping").

**Current state:** Status-based TTD (admin clicks approve/bulk-sign button). No demo-credible signature artifact: no signer identity stamp, no timestamp, no immutable signed snapshot.

**Proposed approach:**
- When Advokat Pendamping approves monthly rekap, capture:
  - Signer name + NPA (advocate license number) from pendamping profile
  - Signature timestamp (UTC + Jakarta timezone display)
  - Optional: drawn signature (canvas) or typed signature text (name in cursive font)
  - Signed content hash: SHA-256 of logbook entries + metadata for that month
- Store in `LogbookRekapSignature` model: `rekap_id`, `signer_id`, `signed_at`, `signature_image_path` (nullable), `content_hash`, `ip_address`
- Display signature block on rekap PDF/view: "Ditandatangani oleh [Name], NPA [license], pada [timestamp]" with drawn/typed signature image if present
- Immutable: once signed, rekap cannot be edited (enforce at controller/policy level)
- Document in FAQ: "E-signature ini adalah tanda persetujuan digital dengan identitas + timestamp. Full PKI crypto signature (if required by regulation) is Fase 2 / future enhancement."

**Acceptance criteria:**
- ✅ Advokat Pendamping can review monthly rekap and click "Tanda Tangan" button
- ✅ Signature capture modal: name + NPA pre-filled, optional drawn/typed signature, "Confirm" button
- ✅ On confirm: store signature record with timestamp, hash, and signer identity
- ✅ Signed rekap displays signature block with all metadata
- ✅ Signed rekap is read-only (cannot edit entries or re-sign)
- ✅ PDF export includes signature block with visual signature if drawn
- ✅ Demo script: pendamping reviews rekap → signs with drawn signature → show PDF with signature block

**Routes/models:**
```php
// app/Models/LogbookRekapSignature.php
Schema::create('logbook_rekap_signatures', function (Blueprint $table) {
    $table->id();
    $table->foreignId('rekap_id')->constrained('logbook_rekap_bulanan')->onDelete('cascade');
    $table->foreignId('signer_id')->constrained('users'); // Advokat Pendamping
    $table->timestamp('signed_at');
    $table->string('signature_image_path')->nullable(); // drawn signature PNG
    $table->string('content_hash', 64); // SHA-256
    $table->ipAddress('ip_address');
    $table->timestamps();
});

// routes/web.php (pendamping routes)
Route::post('logbook/rekap/{id}/sign', [LogbookRekapController::class, 'sign'])->middleware('can:sign,rekap');
```

**Demo script delta:**
1. Pendamping views monthly rekap for Calon X → all entries look good
2. Click "Tanda Tangan Digital" → modal opens
3. Name/NPA pre-filled → draw signature on canvas → "Konfirmasi"
4. Rekap now shows signature block: "Ditandatangani oleh Advokat Y, NPA 54321, pada 13 Sep 2026 15:45 WIB" with drawn signature image
5. Download PDF → signature block appears at bottom
6. Try to edit signed rekap → blocked with message "Rekap yang sudah ditandatangani tidak dapat diubah"

---

### 3. Berkas Digital Completeness for Oath Path

**Deck promise:** Pemberkasan digital that eliminates lost physical logbooks and supports full sumpah (oath) path (slide 6 tahap 3-4: "pemberkasan digital" → "pelaksanaan sumpah").

**Current state:** Upload/download exists but oath package compilation thin vs pitch. Missing: consolidated berkas bundle (all signed logbook rekap PDFs + verification checklist + audit sign-off) for one-click download or export.

**Proposed approach:**
- After Audit Akhir passes, DPC admin can generate "Paket Berkas Sumpah" for the Calon:
  - ZIP bundle containing:
    - All 24 signed monthly logbook rekap PDFs (with e-signatures)
    - Verification checklist summary (PDF showing dual-verification completion)
    - Audit sign-off certificate (PDF with admin signature + audit date)
    - Cover sheet (metadata: calon name, NPA application number, internship period, firm name)
  - Store ZIP in `berkas_sumpah` storage disk, link via `BerkasSumpah` model: `calon_id`, `generated_at`, `generated_by` (admin), `file_path`, `download_count`
- Admin dashboard: after audit passes, "Generate Berkas Sumpah" button → background job creates ZIP → download link appears
- Calon and Pendamping can view/download the same package (read-only)
- Optional: email notification when package is ready

**Acceptance criteria:**
- ✅ After Audit Akhir passes, admin sees "Generate Berkas Sumpah" button for that Calon
- ✅ Click generate → background job compiles all 24 rekap PDFs + checklist + audit cert into ZIP
- ✅ ZIP download link appears on admin dashboard and calon profile
- ✅ ZIP contents: 24 signed rekap PDFs, verification_checklist.pdf, audit_certificate.pdf, cover_sheet.pdf
- ✅ Each PDF is properly formatted with signature blocks and DPC branding
- ✅ Calon receives email: "Berkas sumpah Anda sudah siap untuk diunduh"
- ✅ Demo script: admin completes audit → generates package → downloads ZIP → opens sample PDFs to show signatures

**Routes/models:**
```php
// app/Models/BerkasSumpah.php
Schema::create('berkas_sumpah', function (Blueprint $table) {
    $table->id();
    $table->foreignId('calon_id')->constrained('users');
    $table->timestamp('generated_at');
    $table->foreignId('generated_by')->constrained('users'); // admin
    $table->string('file_path'); // path to ZIP
    $table->integer('download_count')->default(0);
    $table->timestamps();
});

// routes/web.php (admin routes)
Route::post('admin/calon/{id}/generate-berkas-sumpah', [AdminAuditController::class, 'generateBerkasSumpah']);
Route::get('berkas-sumpah/{id}/download', [BerkasSumpahController::class, 'download'])->middleware('auth');

// app/Jobs/GenerateBerkasSumpahJob.php
// Compile PDFs, create ZIP, store in storage, update BerkasSumpah record
```

**Demo script delta:**
1. Admin completes final audit for Calon X → audit passes
2. Click "Generate Berkas Sumpah" button → spinner shows "Sedang menyusun berkas..."
3. After 5-10 seconds → "Berkas sumpah siap diunduh" with download icon
4. Click download → ZIP file `berkas_sumpah_calon_X_2026.zip` downloads
5. Extract ZIP → show folder contents: 24 rekap PDFs (named by month), verification_checklist.pdf, audit_certificate.pdf, cover_sheet.pdf
6. Open one rekap PDF → show signature block with pendamping signature
7. Open audit_certificate.pdf → show DPC admin signature + audit pass stamp

---

**Implementation priority:** 1 → 2 → 3 (firm self-reg unblocks demo flow, e-signature makes logbook credible, berkas package closes the oath path story).

**Out of scope for this parity plan:**
- ❌ Pengadilan Tinggi export (not a warning item; handle separately if needed)
- ❌ Mobile app (slide 5 opsional, Fase 2)
- ❌ PKPA/UPA admisi (Fase 2)

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
