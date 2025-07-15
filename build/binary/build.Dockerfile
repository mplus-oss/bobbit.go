FROM docker.io/library/golang:1.24 AS build
WORKDIR /app
COPY . .
RUN ./build/binary/compile.sh

FROM scratch AS artifacts
COPY --from=build /app/build/dist /dist
