# Single auth origin with a host-only session cookie

## Context

SSO melayani Project yang berada di domain berbeda. Cookie di parent domain (`.example.com`) bisa memberi silent SSO lintas subdomain, tetapi memperlebar blast radius cookie dan mengasumsikan semua Project berbagi registrable domain.

## Decision

Login UI dan API dilayani dari satu origin `https://auth.<domain>`. Session cookie `__Host-sso_session` bersifat host-only, HttpOnly, Secure, dan SameSite=Lax. Setiap Project tetap di domainnya sendiri dan redirect ke auth. Tidak ada custom domain per Project.

## Consequences

Tidak ada silent SSO lintas subdomain; setiap Project melewati redirect OIDC standar. Cookie tidak bisa dibaca oleh aplikasi Project. Berpindah ke parent-domain cookie nanti akan melemahkan isolasi dan sengaja ditolak.
