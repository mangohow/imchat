FROM golang:1.20-alpine AS builder

ARG SERVER_NAME

RUN echo ${SERVER_NAME}

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY="https://goproxy.cn,direct" \
    GO111MODULE=on

WORKDIR /build

COPY . .

RUN go mod download
RUN go build -ldflags="-s -w" -o /app/${SERVER_NAME} ./cmd/${SERVER_NAME}/main.go


# 该镜像用于运行web程序
FROM alpine

ARG SERVER_NAME
ARG SERVER_PORT

RUN echo ${SERVER_NAME}
RUN     echo ${SERVER_PORT}

ENV SERVER_APP=${SERVER_NAME}

RUN echo ${SERVER_APP}

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/${SERVER_NAME} /app/
COPY --from=builder /build/conf/${SERVER_NAME}.yaml /app/conf/

EXPOSE ${SERVER_PORT}

CMD -conf conf/${SERVER_APP}.yaml
ENTRYPOINT ./${SERVER_APP}
