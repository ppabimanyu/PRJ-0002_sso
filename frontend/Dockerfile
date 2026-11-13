# syntax=docker/dockerfile:1

FROM node:22-alpine AS builder

WORKDIR /app

# Install dependencies from the lockfile before copying the rest of the source.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Same-origin deployments leave this empty. Set a full origin only when the
# frontend and API are served from different origins.
ARG VITE_API_URL=
ARG VITE_APP_NAME=SSO
ARG VITE_MAINTENANCE_MODE=false
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_APP_NAME=${VITE_APP_NAME}
ENV VITE_MAINTENANCE_MODE=${VITE_MAINTENANCE_MODE}

RUN npm run build

FROM nginx:alpine-slim

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
