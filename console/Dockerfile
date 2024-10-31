FROM golang:1.16.5 as backend-builder

WORKDIR /workspace

COPY . .
ENV GOPROXY=https://goproxy.cn,direct
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -o backend-server console/backend/main.go

FROM node:12.14.1 as frontend-builder

WORKDIR /workspace

COPY console/frontend/ .
RUN rm -rf ./dist && rm -rf ./node_modules && rm -f ./package-lock.json
RUN npm --registry=https://registry.npmmirror.com install --no-optional
RUN npm run build

FROM alpine:latest
WORKDIR /

ARG ARG_TZ=Etc/UTC
COPY --from=frontend-builder /workspace/dist ./console/frontend/dist
COPY --from=backend-builder /workspace/backend-server ./backend-server
ENV TZ=$ARG_TZ
RUN apk add -U tzdata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
#RUN chmod +x ./startup.sh

ENTRYPOINT ["/backend-server"]
