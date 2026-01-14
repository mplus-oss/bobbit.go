FROM docker.io/library/golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN apk add --no-cache gcc musl-dev bash;
RUN CONTAINERIZED=1 ./build/binary/compile.sh;

FROM scratch AS artifacts
COPY --from=build /app/build/dist /dist
