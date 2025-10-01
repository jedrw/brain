FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build

FROM python:slim
RUN pip install mkdocs mkdocs-material
RUN useradd --create-home --shell /bin/bash brain
USER brain
WORKDIR /brain
COPY --from=build --chown=brain:brain /app/brain .
CMD ["./brain", "serve", "-c", "config.yaml"]