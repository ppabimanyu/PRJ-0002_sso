# Spec — SSO Service v1

**Status:** Draft — published as [GitHub issue #1](https://github.com/ppabimanyu/PRJ-0002_sso/issues/1)
**Sumber:** `docs/architecture.md`, `GLOSSARY.md`, ADR 0001–0007

## Problem Statement

Saya memiliki beberapa project pribadi, masing-masing dengan user internal dan publik. Saat ini setiap project harus memiliki login, session, reset password, dan manajemen user sendiri. Itu berarti kredensial dan data pribadi tersebar di banyak aplikasi, user harus membuat akun berulang kali, dan setiap project menanggung risiko keamanan yang sama.

Saya tidak ingin mengoperasikan platform IAM besar seperti Zitadel: sebagian besar fiturnya tidak terpakai, dan biaya operasionalnya lebih besar dari manfaatnya. Saya juga tidak ingin mengunci aplikasi ke satu vendor. Yang saya butuhkan adalah satu layanan SSO kecil yang berbicara OIDC standar, bisa digantikan Zitadel atau Keycloak di kemudian hari, dan cukup aman untuk user publik.

## Solution

Satu SSO service ringan yang menjadi satu-satunya identity provider untuk semua project pribadi:

- Satu akun global untuk semua Project, dengan akses dan Role per Project.
- Hosted login yang mendukung password, passkey, Magic Link, Google, GitHub, TOTP, dan Recovery Code.
- OIDC/OAuth 2.0 standar: Authorization Code + PKCE, Client Credentials, refresh rotation, discovery, JWKS, introspection, revocation, userinfo, dan RP-initiated logout.
- Account console untuk self-service User dan admin console untuk mengelola Project, Application, Role, Membership, User, Service Account, dan Audit Event.
- Satu deployment Go + React dengan PostgreSQL dan Redis, dilayani dari satu origin `auth.<domain>`.
- Batas yang jelas: tanpa SAML, LDAP, SCIM, organization, Actions JavaScript, atau multi-region.

## User Stories

### Autentikasi dan SSO

1. As a Visitor, I want to log in with email and password, so that I can access any Project that trusts the SSO service.
2. As a Visitor, I want to log in with a passkey, so that I don't have to remember a password.
3. As a Visitor, I want to request a Magic Link, so that I can log in without a password.
4. As a Visitor, I want to log in with Google, so that I can reuse my existing identity.
5. As a Visitor, I want to log in with GitHub, so that I can reuse my existing identity.
6. As a User, I want to move between Projects without logging in again, so that SSO feels seamless.
7. As a User, I want to see which Application is requesting access and which scopes it wants, so that I can make an informed Consent decision.
8. As a User, I want to allow or deny Consent, so that I control access to my identity.
9. As a User, I want my Consent to be remembered for an Application, so that I am not asked on every login.
10. As a User, I want to be redirected back to the Application after login, so that the original request completes.
11. As a User, I want to log out, so that my Session ends on the auth domain.
12. As a User, I want to see active Sessions and revoke them, so that I can remove access from a lost device.
13. As a User, I want my Session to expire after inactivity, so that an unattended browser does not stay logged in.
14. As a User, I want clear but non-revealing error messages, so that failed logins don't expose whether an email is registered.
15. As a disabled User, I want to be blocked from logging in, so that an admin can suspend access immediately.
16. As a User, I want an expired or reused authorization code to be rejected, so that my Session cannot be hijacked.
17. As a User, I want an invalid redirect URI to be rejected, so that my credentials are not sent to an attacker.
18. As an Application developer, I want the provider to return `state` exactly as sent, so that my client can validate it.
19. As an Application developer, I want an invalid `nonce` to be rejected, so that replay is prevented.

### MFA

20. As a User, I want to enroll TOTP, so that I can add a second factor.
21. As a User, I want to register a passkey, so that I can use phishing-resistant authentication.
22. As a User, I want to name and remove passkeys, so that I can manage my devices.
23. As a User, I want to generate Recovery Codes, so that I can recover access if I lose my MFA device.
24. As a User, I want to use a Recovery Code, so that I can log in when my primary MFA method is unavailable.
25. As a User, I want to see which MFA methods are active, so that I know how my account is protected.
26. As an Admin, I want MFA to be required for my account, so that admin actions are protected.
27. As a User, I want TOTP replay to be rejected, so that an intercepted code cannot be reused.
28. As a User, I want an MFA challenge to expire quickly, so that a half-finished login cannot be resumed by someone else.

### Registrasi dan Membership

29. As a Visitor, I want to register on an Application with open access, so that I can create an account.
30. As a Visitor, I want to verify my email, so that my account is trusted.
31. As an Admin, I want to invite a User to an Application with invite-only access, so that only approved people can join.
32. As an Admin, I want to assign Roles to a User in a Project, so that the Application receives the right authorization.
33. As a User, I want to see my Projects and Roles, so that I know where I have access.
34. As a Visitor, I want to be rejected from an Application with closed access, so that the access policy is enforced.
35. As a User, I want to change my email with verification, so that I can keep my account current.
36. As a User, I want to reset my password by email, so that I can recover access.
37. As a User, I want to change my password, so that I can rotate a compromised credential.
38. As a User, I want to delete my account, so that my PII is removed after the grace period.
39. As a User, I want to be notified when my password or email changes, so that I can detect unauthorized changes.
40. As a User, I want a password reset link to be single-use and short-lived, so that it cannot be replayed.
41. As a User, I want an unverified email to be rejected where verification is required, so that I cannot bypass the policy.
42. As a User, I want a deleted account to be recoverable during the grace period, so that accidental deletion is reversible.

### Social linking

43. As a User, I want my social account to be linked automatically only when the provider verifies my email, so that account takeover is prevented.
44. As a User, I want to link a social account explicitly when auto-linking is not safe, so that I can still use social login.
45. As a User, I want to unlink a social account only when I have another login method, so that I don't lock myself out.
46. As a User, I want to see which Upstream Identity Providers are linked, so that I can manage them.
47. As a User, I want a social signup to create a User with the provider's verified email, so that my account reflects reality.
48. As a User, I want a GitHub login without a verified primary email to require explicit linking, so that an unverified address cannot take over an account.

### Admin

49. As an Admin, I want to create a Project, so that I can group Applications and Roles.
50. As an Admin, I want to update or delete a Project, so that I can keep the registry accurate.
51. As an Admin, I want to register an Application with exact redirect URIs, so that OIDC clients can integrate safely.
52. As an Admin, I want to update an Application's redirect URIs and access mode, so that I can change its behavior without recreating it.
53. As an Admin, I want to disable an Application, so that a compromised client stops working.
54. As an Admin, I want to rotate a client secret, so that I can recover from a leak.
55. As an Admin, I want to define Roles per Project, so that Applications can authorize Users.
56. As an Admin, I want to disable or enable a User, so that I can suspend access.
57. As an Admin, I want to add or remove Memberships and Roles, so that I can manage access.
58. As an Admin, I want to force a password reset, so that I can help a User regain access.
59. As an Admin, I want to view and revoke Sessions, so that I can respond to incidents.
60. As an Admin, I want to view Audit Events, so that I can investigate security-relevant actions.
61. As an Admin, I want to create a Service Account and rotate its secret, so that machine clients can authenticate.
62. As an Admin, I want to mark an Application as trusted or untrusted, so that Consent is only shown when needed.
63. As an Admin, I want to see a User's Project Memberships and Roles, so that I can verify access.
64. As an Admin, I want the first Admin to be bootstrapped from configuration, so that a fresh deployment is manageable.

### OIDC dan integrasi

65. As an Application developer, I want OIDC discovery, so that standard libraries can configure themselves.
66. As an Application developer, I want a JWKS endpoint, so that resource servers can verify tokens offline.
67. As an Application developer, I want Authorization Code + PKCE, so that web and SPA clients can authenticate safely.
68. As an Application developer, I want an ID Token with `nonce`, so that I can prevent replay.
69. As an Application developer, I want an Access Token with audience-scoped Roles, so that my API can authorize requests without knowing other Projects.
70. As an Application developer, I want refresh token rotation with reuse detection, so that stolen refresh tokens are contained.
71. As an Application developer, I want introspection and revocation, so that resource servers that cannot use JWKS still work.
72. As an Application developer, I want a userinfo endpoint, so that I can fetch profile claims with a valid Access Token.
73. As an Application developer, I want RP-initiated logout, so that Users can end their Session cleanly.
74. As an Application developer, I want standard OAuth error responses, so that my client library can handle failures.
75. As a Service Account, I want a client credentials token with my Roles, so that my service can call APIs.
76. As an Application developer, I want a demo client, so that I can see the full flow working before integrating my own app.
77. As an Application developer, I want tokens to include `sid` and `amr`, so that I can correlate Sessions and authentication methods.
78. As an Application developer, I want the provider to reject unsupported flows and scopes, so that misconfiguration fails loudly.

### Operasi dan keamanan

79. As an Operator, I want rate limiting and temporary lockout, so that brute force is bounded.
80. As an Operator, I want Audit Events for security-relevant actions, so that I can investigate incidents.
81. As an Operator, I want email delivery through an outbox with retries, so that Magic Links and verification emails are reliable.
82. As an Operator, I want health and readiness endpoints, so that deployments and monitoring work.
83. As an Operator, I want OTel traces, metrics, and structured logs, so that I can debug and observe the service.
84. As an Operator, I want a single-origin deployment with TLS, so that cookies and redirects stay safe.
85. As an Operator, I want signing key rotation without downtime, so that I can rotate keys safely.
86. As an Operator, I want database migrations before deploy, so that schema changes are safe.
87. As an Operator, I want backup and restore procedures, so that I can recover from data loss.
88. As an Operator, I want OpenID conformance validation, so that I know the provider interoperates.
89. As an Operator, I want PII retention and deletion rules, so that the service respects privacy commitments.
90. As an Operator, I want auth rate limiting to stay bounded when Redis is unavailable, so that brute force is not unbounded.

## Implementation Decisions

### Architecture

- Satu deployable Go service dengan PostgreSQL dan Redis; frontend React SPA dilayani dari origin yang sama.
- Protokol OIDC memakai library `github.com/zitadel/oidc/v3`; service ini mengimplementasikan storage adapter dan alur auth, bukan protokolnya dari nol. Lihat ADR 0001.
- Single tenant: satu Instance, satu user pool global, `Project` sebagai produk, `Application` sebagai OIDC client, akses lewat `Membership` dan `Role`. Tidak ada Organization atau Tenant. Lihat ADR 0002.
- Hanya OIDC/OAuth 2.0. Tidak ada SAML, LDAP, SCIM, device flow, atau token exchange. Lihat ADR 0003.
- Modul auth domain memegang logika multi-step (login, MFA, Magic Link, social linking, session, token); handler HTTP hanya decode, validasi, dan serialisasi. Ini deviasi yang disengaja dari aturan boilerplate "tanpa layer service". Lihat ADR 0007.
- Repository mengimplementasikan port di domain; modul auth tidak mengimpor handler atau repository secara langsung.

### Domain model

- `Instance`, `User`, `Project`, `Application`, `Service Account`, `Membership`, `Role`, `Session`, `Auth Request`, `Consent`, `Email Token`, `Magic Link`, `Audit Event`, `Trusted Application`, `Recovery Code`, `Upstream Identity Provider`, `Provider Link` mengikuti `GLOSSARY.md`.
- Satu User boleh menjadi anggota banyak Project; satu Application hanya milik satu Project; Role hanya berlaku di Project tempat ia didefinisikan.
- `is_admin` adalah satu-satunya role global dan wajib MFA.
- `access_mode` Application: `open` (registrasi terbuka, semua User terautentikasi), `invite` (hanya undangan dan Membership), `membership` (hanya User dengan Membership).
- Registrasi publik tidak otomatis membuat Membership; Role diberikan lewat admin console atau undangan.

### Token dan Session

- Access Token JWT RS256, 15 menit, `aud` = client_id Application, `roles` hanya dari Project Application tersebut, plus `sid`, `amr`, `scope`, dan claim profil sesuai scope. Lihat ADR 0004 dan ADR 0006.
- Refresh Token opaque 256-bit, 30 hari, di-hash SHA-256, dirotasi setiap dipakai; reuse mencabut seluruh family dan Session terkait.
- ID Token standar OIDC dengan `nonce`.
- Session cookie opaque `__Host-sso_session`, HttpOnly, Secure, SameSite=Lax, host-only; idle 30 hari, absolut 90 hari.
- Session ID dirotasi saat login dan privilege change; ganti password dan reset password mencabut Session lain.
- Signing key berupa PEM di file/secret mount; JWKS menyajikan key aktif dan key sebelumnya; rotasi manual tanpa downtime.
- Admin console memakai Session same-origin, bukan Access Token. Audience `sso-admin` hanya relevan jika console dipisah ke origin lain.
- RP-initiated logout; tidak ada back-channel logout di v1.

### Data

- PostgreSQL dengan goose dan sqlc; UUIDv7 dibuat aplikasi; semua timestamp `timestamptz` UTC; email `citext`.
- Tabel: `users`, `password_credentials`, `totp_credentials`, `webauthn_credentials`, `recovery_codes`, `provider_links`, `projects`, `roles`, `applications`, `service_accounts`, `service_account_roles`, `memberships`, `membership_roles`, `consents`, `sessions`, `auth_requests`, `auth_codes`, `email_tokens`, `webauthn_challenges`, `audit_events`, `outbox_emails`.
- Index unik pada email, token hash, dan client_id; index retensi pada expiry dan `outbox_emails(status, next_attempt_at)`.
- Retensi: Session kedaluwarsa 30 hari, Auth Request/Code 1 hari, Email Token 7 hari, email terkirim 30 hari, Audit Event 1 tahun, PII User soft-deleted 30 hari.

### API

- OIDC: `/.well-known/openid-configuration`, `/oidc/jwks`, `/oauth2/authorize`, `/oauth2/token`, `/oauth2/userinfo`, `/oauth2/introspect`, `/oauth2/revoke`, `/oauth2/end_session`.
- Auth API: konteks Auth Request, login password, Magic Link request/consume, WebAuthn begin/finish, TOTP verify, Recovery Code verify, social start/callback, Consent, Session, logout.
- Account API: profil, ganti email, ganti password, MFA, Session, Provider Link, Membership, hapus akun.
- Admin API: CRUD Project, Application, Role, User, Membership, Service Account; rotasi secret; disable/enable; reset password; Audit Event; Session.
- Konvensi boilerplate: REST `/api/v1`, JSON `camelCase`, error RFC 7807, pagination `limit/offset`, OpenAPI dari anotasi Go.
- OIDC error memakai format OAuth standar, bukan RFC 7807.

### Auth flows

- Authorization Code + PKCE: validasi client, redirect URI exact match, scope, PKCE `S256`; cek Session; redirect ke login UI dengan Auth Request; setelah login lanjut ke Consent atau code; code 60 detik dan sekali pakai.
- Magic Link: response selalu sama untuk mencegah enumerasi; Email Token 10 menit, sekali pakai, terikat Auth Request; hanya untuk User yang sudah ada.
- Social: Google dan GitHub; `state` + PKCE; auto-link hanya jika provider menyatakan email terverifikasi dan email lokal sudah terverifikasi; selain itu link eksplisit; unlink hanya jika ada metode login lain.
- MFA: TOTP wajib untuk Admin; passkey bisa menjadi faktor pertama dan kedua; Recovery Code 10 kode sekali pakai; MFA ticket 5 menit, sekali pakai, disimpan sebagai hash dan cookie HttpOnly.
- Client Credentials: Service Account sebagai OAuth client; `sub` = Service Account, `aud` = audience terkonfigurasi, `roles` = Role Service Account; tanpa refresh token.

### Keamanan

- Password Argon2id (memory 64 MiB, iterations 3, parallelism 2, salt 16 byte, hash 32 byte); panjang 12–128; tanpa aturan komposisi; tolak password yang sama dengan email.
- PKCE wajib untuk public client; redirect URI exact match, HTTPS kecuali localhost, tanpa wildcard; `state` dan `nonce` divalidasi.
- Session cookie `__Host-`; state-changing request memvalidasi Origin/Sec-Fetch-Site.
- Rate limit Redis + fallback in-process; login 10 kegagalan/15 menit per akun + IP, ditahan 15 menit; Magic Link/reset 5 request/15 menit; token endpoint 60 request/menit per client.
- Client secret dan Recovery Code di-hash SHA-256; TOTP secret dienkripsi AES-GCM dengan `APP_ENCRYPTION_KEY`; Session, refresh, code, dan Email Token disimpan sebagai hash.
- Header HSTS, CSP `default-src 'self'; frame-ancestors 'none'`, nosniff, Referrer-Policy, Permissions-Policy.
- Audit Event untuk login sukses/gagal, logout, registrasi, verifikasi, reset, ganti email, MFA, client secret, Role/Membership, disable/enable, dan akses admin.
- Log tidak memuat token, password, TOTP secret, atau isi email.

### Email

- Outbox pattern; worker in-process; retry exponential 1m, 5m, 30m, 2h, 6h, maksimum 10 attempt.
- SMTP generik; Mailpit di development.
- Template: verifikasi, reset password, Magic Link, ganti email, undangan, notifikasi perubahan password.
- TTL: verifikasi 24 jam, reset 30 menit, Magic Link 10 menit, undangan 7 hari.

### Frontend

- Satu React app dengan tiga area: login/consent, account, admin. Route mengikuti `docs/architecture.md`.
- Bahasa UI dan template email: English; dokumentasi internal: Indonesian.
- Tidak ada SSR; tidak ada token di `localStorage`/`sessionStorage`; console memakai Session cookie.
- Login memakai alur email-first: masukkan email, lalu tampilkan metode yang tersedia.
- Error problem+json dipetakan ke UI tanpa membocorkan apakah email terdaftar.

### Deployment

- Satu origin `https://auth.<domain>`; Caddy menyajikan frontend dan mem-proxy path backend.
- Docker Compose: db, redis, migrasi, backend, frontend, caddy; profil dev menambah Mailpit; profil observability menambah OTel Collector/Jaeger/Prometheus.
- Migrasi sebelum backend baru menerima traffic; backup PostgreSQL harian, retensi 14 hari; Redis tidak durable; single instance tanpa HA.
- Konfigurasi via environment; secret tidak di-commit.

### Milestones

- **M0** — fork boilerplate, rename module, CI, compose + Mailpit; `make dev` jalan, CI hijau.
- **M1** — OIDC core, password login, registry Project/Application, demo client; satu app login end-to-end dan menerima token dengan Role.
- **M2** — registrasi, verifikasi email, reset password, ganti email, account console, email outbox.
- **M3** — TOTP, passkey, Recovery Code, Session management, social login, Magic Link.
- **M4** — admin console lengkap, audit viewer, hardening, conformance suite.
- **M5** — production: TLS, backup, monitoring, runbook; domain auth live dan Project pertama memakai SSO.

## Testing Decisions

### What makes a good test

Test harus menguji perilaku yang bisa diamati dari luar, bukan detail implementasi. Untuk backend, artinya status HTTP, header, body, cookie, redirect, dan claim token; state yang tersimpan diverifikasi lewat request API berikutnya, bukan query langsung ke tabel atau pemanggilan fungsi private. Untuk frontend, artinya UI yang dirender dan interaksi user, bukan state internal komponen.

### Seams

- **Backend — satu seam utama: HTTP API.** Semua perilaku backend diuji lewat server HTTP nyata dengan PostgreSQL dan Redis nyata (testcontainers). Ini seam tertinggi dan menutup OIDC, login, Session, refresh, revocation, account, admin, dan Service Account. Gunakan HTTP client biasa dan satu OIDC client standar untuk flow end-to-end.
- **Frontend — satu seam: page/feature dengan MSW.** Semua perilaku UI diuji lewat halaman atau feature component dengan network di-mock di boundary HTTP.
- **Tidak ada seam repository terpisah.** Perilaku repository diuji lewat HTTP; ini menghindari duplikasi dan menjaga jumlah seam tetap minimal.
- **Pengecualian sempit:** helper murni seperti hashing Argon2id, jendela TOTP, dan pembentukan claim boleh punya table test unit. Ini bukan seam perilaku utama dan tidak boleh menjadi tempat mayoritas test.
- **Conformance:** OpenID basic OP suite berjalan black-box terhadap HTTP API yang sama; ini gate sebelum v1, bukan gate setiap PR.

### Prior art

Boilerplate sudah memakai testify + testcontainers untuk integration test repository, dan Vitest + React Testing Library + MSW untuk frontend. Test baru mengikuti pola tersebut. Belum ada test auth di codebase; spesifikasi ini menetapkan seam-nya.

## Out of Scope

- SAML 2.0, LDAP, SCIM, template Entra/Okta.
- Organization, tenant, granted project, delegated administration, custom domain per Project.
- Actions JavaScript, token exchange, impersonation, custom login API.
- Device authorization, DPoP, PAR, JARM, mTLS, back-channel logout.
- SMS, i18n, quota, per-Project branding/policy.
- Event sourcing/CQRS, CockroachDB, gRPC, HA multi-region.
- Billing, team management, user import.
- Webhook keluar, generic OIDC Upstream Identity Provider, dan consent revocation (P2).
- Menyimpan token di browser.

## Further Notes

- Spec ini mencakup v1 (M1–M5). Untuk implementasi agent, mulai dari M0 lalu M1; M1 adalah slice vertikal pertama yang harus berakhir di demo client yang benar-benar login.
- Admin console memakai Session same-origin, bukan Access Token terpisah. Ini refinement dari diskusi awal dan sudah tercatat di `docs/architecture.md`.
- Referensi lengkap: `docs/architecture.md`, `GLOSSARY.md`, dan ADR 0001–0007.
- Spec dipublikasikan sebagai GitHub issue dengan label `ready-for-agent`.
