# Build stage
FROM node:alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

    # Install dependencies
    # Use npm install to generate/update package-lock.json if needed
    RUN npm install

# Copy source code
COPY . .

# Build the application with relative API URL for nginx proxy
ARG VITE_API_URL=/api
ENV VITE_API_URL=${VITE_API_URL}
RUN npm run build

# Production stage
FROM nginx:alpine

# Install envsubst for template substitution
RUN apk add --no-cache gettext

# Ensure conf.d directory exists and is writable
RUN mkdir -p /etc/nginx/conf.d && chmod 755 /etc/nginx/conf.d

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration template
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Copy startup script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]

