# JWT access tokens with rotating refresh tokens

## Context

Application harus bisa memverifikasi token tanpa memanggil SSO service di setiap request, tetapi revocation tetap harus terbatas. Session SSO juga perlu bertahan antar Application tanpa membawa token di cookie.

## Decision

Access token adalah JWT RS256 berumur 15 menit, diverifikasi offline lewat JWKS, dengan `kid` untuk rotasi key. Refresh token opaque berumur 30 hari, di-hash di database, dirotasi setiap dipakai, dan reuse detection mencabut seluruh family. Session IdP adalah cookie opaque `__Host-sso_session` dengan idle 30 hari dan absolut 90 hari. Introspection tersedia untuk resource server yang tidak bisa memakai JWKS. Logout memakai RP-initiated logout; tanpa back-channel.

## Consequences

Revocation access token punya jeda hingga 15 menit; revocation instan hanya berlaku untuk refresh token dan session. Refresh token family harus disimpan untuk mendeteksi reuse. Back-channel logout tidak tersedia di v1.
