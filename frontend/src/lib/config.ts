// Generated OpenAPI paths already include /api/v1. Keep the base empty for
// same-origin deployments; set VITE_API_URL to a full origin only when the API
// lives elsewhere.
const DEFAULT_API_URL = "";

export const apiBaseUrl = import.meta.env.VITE_API_URL || DEFAULT_API_URL;
