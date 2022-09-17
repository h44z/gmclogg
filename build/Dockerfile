FROM golang:latest

RUN set -x \
    && dpkg --add-architecture arm64 \
    && apt-get update \
    && apt-get install -y -q --no-install-recommends \
           xz-utils \
           libusb-1.0-0-dev:arm64

RUN wget "https://ziglang.org/download/0.9.1/zig-linux-x86_64-0.9.1.tar.xz" \
    && tar xf zig-linux-x86_64-0.9.1.tar.xz \
    && ln -s $PWD/zig-linux-x86_64-0.9.1/zig /usr/bin/ \
    && zig version

RUN mkdir /build

WORKDIR /build

ENTRYPOINT ["make", "build-cross-plat"]