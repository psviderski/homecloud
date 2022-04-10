VERSION 0.6
FROM alpine:3.15.3
RUN apk add --no-cache \
    curl

ubuntu-builder:
    FROM ubuntu:20.04
    # Speed up package download in Australia.
    RUN sed -i s/archive.ubuntu/au.archive.ubuntu/ /etc/apt/sources.list
    RUN apt-get update
    RUN apt-get install -y \
        bison \
        build-essential \
        curl \
        flex \
        gcc-aarch64-linux-gnu \
        libssl-dev
    WORKDIR /homecloud

u-boot:  # Only RPi4 is targeted for now.
    FROM +ubuntu-builder
    ARG UBOOT_VERSION=v2022.04

    RUN mkdir u-boot
    RUN curl -fsSL "https://github.com/u-boot/u-boot/archive/refs/tags/${UBOOT_VERSION}.tar.gz" \
        | tar -xzf - --strip-components 1 -C u-boot
    WORKDIR u-boot
    RUN make rpi_4_defconfig
    RUN CROSS_COMPILE=aarch64-linux-gnu- make u-boot.bin
    SAVE ARTIFACT u-boot.bin

rpi4-firmware:
    # See https://github.com/raspberrypi/firmware/tags
    ARG FIRMWARE_VERSION=1.20220331
    RUN curl -fsSL "https://github.com/raspberrypi/firmware/archive/refs/tags/${FIRMWARE_VERSION}.tar.gz" \
        | tar -xzf - --strip-components 1 "firmware-${FIRMWARE_VERSION}/boot"
    SAVE ARTIFACT boot/bcm2711* /boot/
    SAVE ARTIFACT boot/fixup4.dat /boot/
    SAVE ARTIFACT boot/start4.elf /boot/
    SAVE ARTIFACT boot/overlays /boot/

rpi4-cos-image:
    FROM DOCKERFILE -f Dockerfile.rpi4 .
    COPY +u-boot/u-boot.bin /.system-boot/
    COPY --dir +rpi4-firmware/boot/* /.system-boot/
    SAVE IMAGE hcos-rpi4:latest

rpi4-image-deps:
    ENV LUET_VERSION=0.30.3
    RUN apk --no-cache add \
        bash \
        coreutils \
        curl \
        dosfstools \
        e2fsprogs-extra \
        multipath-tools \
        rsync \
        sgdisk \
        util-linux
    RUN wget -O - https://raw.githubusercontent.com/rancher-sandbox/cOS-toolkit/0fff8c1a642ff9a51f00aa4354715319c406a84a/scripts/get_luet.sh | sh
    RUN luet install -y toolchain/elemental-cli

rpi4-image:
    FROM +rpi4-image-deps

    COPY build_image_rpi4.sh .
    WITH DOCKER --load hcos-rpi4:latest=+rpi4-cos-image
        RUN --privileged ./build_image_rpi4.sh --docker-image hcos-rpi4:latest hcos-rpi4.img
    END
    SAVE ARTIFACT hcos-rpi4.img* AS LOCAL ./build/
