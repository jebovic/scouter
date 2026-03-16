FROM node:22-alpine AS builder

ARG VITE_DISABLE_SW=false
ENV VITE_DISABLE_SW=$VITE_DISABLE_SW

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY . .
RUN npm run build

FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
