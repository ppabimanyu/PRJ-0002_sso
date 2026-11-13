# Audience-scoped role claims

## Context

Seorang User bisa menjadi anggota beberapa Project. Satu claim `roles` global akan membocorkan role lintas Project ke setiap token dan memaksa resource server memfilter sendiri.

## Decision

Setiap token memiliki `aud` = client_id Application, dan `roles` hanya berisi Role dari Project milik Application tersebut. Admin console adalah audience terpisah dengan role global `admin`. Role adalah string; tidak ada model permission.

## Consequences

User dengan role di beberapa Project mendapat set role berbeda per token. Resource server bisa mempercayai `roles` tanpa mengetahui topologi Project. Menambahkan role lintas Project nanti membutuhkan claim baru, bukan mengubah arti `roles`.
