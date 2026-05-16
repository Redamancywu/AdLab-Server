ARG NODE_BASE_IMAGE=node:20-alpine
ARG ADMIN_NGINX_IMAGE=nginx:1.27-alpine

FROM ${NODE_BASE_IMAGE} AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM ${ADMIN_NGINX_IMAGE}

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
