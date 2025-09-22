FROM node:20.19-alpine AS build
WORKDIR /src
COPY frontend/datacat-ui/package*.json ./
RUN npm ci
COPY frontend/datacat-ui/ ./
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY .ops/docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist/datacat-ui/browser /usr/share/nginx/html
EXPOSE 8080
