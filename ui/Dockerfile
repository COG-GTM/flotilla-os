FROM node:22-alpine
WORKDIR /usr/src/app
COPY package.json package-lock.json ./
RUN npm ci
RUN npm install -g serve
COPY . .
ARG FLOTILLA_API
ARG DEFAULT_CLUSTER
RUN npm run build
ENTRYPOINT ["serve", "-s", "build"]
