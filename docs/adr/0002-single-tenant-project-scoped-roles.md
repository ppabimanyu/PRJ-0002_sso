# Single-tenant identity model with project-scoped roles

## Context

Multi-tenancy adalah sumber kompleksitas terbesar Zitadel: instance, organization, granted project, delegated administration, dan policy per organization. Project pribadi tidak punya pelanggan B2B yang perlu mengelola user sendiri.

## Decision

Satu Instance dengan satu user pool global. `Project` adalah produk, `Application` adalah OIDC client, dan `User` adalah identitas global. Akses diberikan lewat `Membership` dan `Role` per Project. Tidak ada Organization atau Tenant.

## Consequences

User dari satu Project bisa login ke Project lain yang mengizinkannya (SSO). Tidak ada isolasi user per Project; privasi dijaga lewat Membership. Blast radius adalah satu user pool. Tidak ada delegated admin per organisasi.
