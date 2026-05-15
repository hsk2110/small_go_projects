# syntax=docker/dockerfile:1
FROM golang:1.26
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/rest_api .

FROM scratch
COPY --from=0 /bin/rest_api /bin/rest_api
CMD ["/bin/rest_api"]