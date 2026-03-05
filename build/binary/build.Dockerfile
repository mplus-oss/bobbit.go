FROM docker.io/library/golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache gcc musl-dev bash;
COPY ./go.* .
RUN go mod download -x;
COPY . .
RUN ./build/binary/compile.sh;

FROM scratch AS artifacts
COPY --from=build /app/build/dist/ /
