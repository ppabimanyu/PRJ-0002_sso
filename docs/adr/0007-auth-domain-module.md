# Auth domain module instead of handler-only logic

## Context

Boilerplate menetapkan tidak ada layer service: business logic ditulis di handler, dan hanya logic lintas handler yang masuk `pkg`. Alur auth (login, MFA, magic link, social linking, session) multi-step, security-critical, dan perlu diuji tanpa HTTP.

## Decision

Tambahkan `internal/auth` sebagai modul domain untuk logika alur auth dan penerbitan token. `httphandler` hanya decode, validasi, dan serialisasi response; akses data tetap lewat port di `domain`. Ini bukan layer service generik — hanya modul auth.

## Consequences

Menyimpang dari konvensi boilerplate, jadi guardrail import harus eksplisit: `auth` boleh mengimpor `domain` dan `pkg`; `httphandler` boleh mengimpor `auth`, `domain`, dan `pkg`; `repository` mengimplementasikan port. Logic auth bisa diuji tanpa HTTP, dengan biaya satu package tambahan yang harus dijaga dari kebocoran tanggung jawab.
