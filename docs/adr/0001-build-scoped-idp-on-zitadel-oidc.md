# Build a scoped IdP on the Zitadel OIDC library

## Context

Kita butuh SSO untuk beberapa project pribadi, dengan user internal dan publik. Zitadel adalah referensi fitur, tetapi self-host Zitadel membawa seluruh platform (instance, organization, event store, console, login UI) yang sebagian besar tidak terpakai.

## Decision

Bangun IdP sendiri di atas `github.com/zitadel/oidc/v3`, library OIDC OP/RP yang dipakai Zitadel. Protokol, token, JWKS, dan discovery memakai library; yang kita bangun adalah user store, login UI, MFA, session, dan admin. Semua Application hanya melihat kontrak OIDC standar.

## Considered Options

- **Self-host Zitadel**: fitur lengkap, tetapi operasional berat dan scope besar.
- **Keycloak**: matang, tetapi Java dan kompleksitas realm yang tidak dibutuhkan.
- **Ory Hydra + Kratos**: dua service terpisah, integrasi lebih rumit.
- **Menulis protokol OIDC sendiri**: risiko keamanan tinggi, tidak sepadan.

## Consequences

Kita memiliki password hashing, session, WebAuthn, audit, dan bug keamanannya. Tidak ada SAML/LDAP/SCIM. Karena kontraknya standar, Zitadel atau Keycloak tetap bisa menggantikan service ini tanpa mengubah Application.
