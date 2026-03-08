FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ENV USER_SERVICE_URL=http://user-service:8080
ENV PRODUCT_SERVICE_URL=http://product-service:8080
ENV ORDER_SERVICE_URL=http://order-service:8080
ENV PAYMENT_SERVICE_URL=http://payment-service:8080

RUN npm run build

FROM node:22-alpine

WORKDIR /app

ENV NODE_ENV=production

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
