# SSO Service

Identity provider untuk project pribadi: satu akun untuk semua aplikasi, baik user internal maupun publik.

## Language

### Identitas

**Instance**:
Satu deployment SSO service dengan satu kumpulan User.
_Avoid_: Tenant, realm, organization

**User**:
Identitas global satu orang, dikenali lewat alamat email.
_Avoid_: Account, member, customer

**Service Account**:
Identitas non-manusia yang dipakai Application untuk mengakses API.
_Avoid_: Machine user, bot, API user

**Project**:
Produk atau aplikasi pribadi yang memakai SSO; berisi Application dan Role.
_Avoid_: Tenant, client, app

**Application**:
Aplikasi terdaftar di dalam Project yang memakai SSO sebagai OIDC client.
_Avoid_: Client, app, service

**Trusted Application**:
Application yang sudah dipercaya Instance sehingga Consent dilewati.
_Avoid_: First-party app, internal app

**Membership**:
Hubungan antara User dan Project yang memberikan satu atau lebih Role.
_Avoid_: Grant, authorization, access

**Role**:
Label bermakna milik satu Project yang muncul di token; artinya ditentukan oleh Application.
_Avoid_: Permission, scope, group

**Admin**:
User yang boleh mengelola Instance dan seluruh Project di dalamnya.
_Avoid_: Superuser, owner

### Protokol

**Auth Request**:
Permintaan otorisasi OIDC yang menunggu autentikasi atau Consent dari User.
_Avoid_: Login attempt, flow

**Session**:
Sesi terautentikasi User di Instance; sumber SSO antar Application.
_Avoid_: Login, token

**Access Token**:
Token berumur pendek yang membawa identitas User dan Role untuk satu audience.
_Avoid_: API key, session token

**Refresh Token**:
Token berumur panjang yang ditukar dengan Access Token baru; setiap pertukaran membatalkan token sebelumnya.
_Avoid_: Session, API key

**Consent**:
Persetujuan User agar Application mendapat akses ke scope tertentu.
_Avoid_: Approval, permission

**Recovery Code**:
Kode sekali pakai untuk masuk saat metode MFA utama tidak tersedia.
_Avoid_: Backup code, emergency code

**Upstream Identity Provider**:
Identity provider eksternal yang dipakai User untuk login.
_Avoid_: IdP, social provider

**Provider Link**:
Hubungan antara User dan Upstream Identity Provider berdasarkan subject identifier.
_Avoid_: Social account, connection

### Operasi

**Audit Event**:
Catatan permanen tentang aksi yang menyentuh identitas atau akses.
_Avoid_: Log, history

**Email Token**:
Token sekali pakai yang dikirim lewat email untuk membuktikan kepemilikan alamat email.
_Avoid_: OTP, code

**Magic Link**:
URL berisi Email Token yang langsung mengautentikasi User saat dibuka.
_Avoid_: Passwordless login, email login
