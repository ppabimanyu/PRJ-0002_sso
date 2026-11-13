# OIDC and OAuth 2.0 only

## Context

Semua aplikasi yang akan memakai SSO adalah web, SPA, atau API yang kita kendalikan. SAML, LDAP, dan SCIM dibutuhkan untuk federation enterprise dan provisioning organisasi, yang tidak ada dalam kasus ini.

## Decision

Hanya OIDC/OAuth 2.0: Authorization Code + PKCE, Client Credentials, refresh rotation, discovery, JWKS, introspection, dan revocation. Tidak ada SAML, LDAP, SCIM, device flow, atau token exchange di v1.

## Consequences

Tidak bisa menerima tenant enterprise yang mewajibkan SAML. Kalau nanti dibutuhkan, tambahkan sebagai adapter brokering, bukan dengan mengubah model inti.
