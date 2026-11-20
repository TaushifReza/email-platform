FROM node:22-alpine AS build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM nginx:alpine

COPY --from=build /app/out /usr/share/nginx/html
COPY infra/nginx/web.conf /etc/nginx/conf.d/default.conf
