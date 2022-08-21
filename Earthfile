VERSION 0.6
FROM alpine:3.15.3
RUN apk add --no-cache \
    curl

ubuntu-builder:
    FROM ubuntu:20.04@sha256:af5efa9c28de78b754777af9b4d850112cad01899a5d37d2617bb94dc63a49aa
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

luet:
    ARG LUET_VERSION=0.32.1
    # Use the native platform image which is a hacky workaround to speed up the arm64 container image build when run
    # on an amd64 host. The `luet install` commands below will run without emulation but this requires the luet config
    # to explicitly enable the arm64 repository.
    FROM quay.io/luet/base:$LUET_VERSION
    SAVE ARTIFACT /usr/bin/luet

opensuse-leap-base:
    FROM --platform=linux/arm64 opensuse/leap:15.4@sha256:495697747909d2f5830c3d55257fa27a61b339c1505f5a5164f5945f82bb16e4
    ENV LUET_NOLOCK=true

    COPY +luet/luet /usr/bin/luet
    COPY luet_arm64.yaml /etc/luet/luet.yaml

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

kernel-arm64:
    FROM +opensuse-leap-base

    RUN luet install -y \
        meta/cos-core \
        && luet cleanup \
        # Patch the elemental immutable-rootfs Dracut module to include missing utils.
        # TODO: contribute the fix to upstream.
        && sed -i "s/ cut/ basename cut/" /usr/lib/dracut/modules.d/30cos-immutable-rootfs/module-setup.sh
    RUN zypper install -y --no-recommends \
        curl \
        device-mapper \
        dracut \
        e2fsprogs \
        iproute2 \
        kernel-default \
        # parted is used to partition a disk or expand partitions in the cloud-config rootfs stage.
        parted \
        rsync \
        tar \
        wicked \
        xz \
        && zypper clean --all

    # TODO: looks like after changing the base to Tumbleweed cos-setup package has to be changed to use /usr/lib/systemd
    RUN ln -s /usr/lib/systemd /lib/systemd && mkinitrd
    # elemental-toolkit expects the kernel at /boot/vmlinuz.
    RUN ln -s Image /boot/vmlinuz

    SAVE ARTIFACT --symlink-no-follow /boot/Image* /boot/
    SAVE ARTIFACT /boot/System.map-* /boot/
    SAVE ARTIFACT /boot/config-* /boot/
    SAVE ARTIFACT --symlink-no-follow /boot/initrd* /boot/
    SAVE ARTIFACT --symlink-no-follow /boot/vmlinuz /boot/
    SAVE ARTIFACT /lib/modules /lib/modules

rpi4-firmware:
    # See https://github.com/raspberrypi/firmware/tags
    ARG FIRMWARE_VERSION=1.20220331
    RUN curl -fsSL "https://github.com/raspberrypi/firmware/archive/refs/tags/${FIRMWARE_VERSION}.tar.gz" \
        | tar -xzf - --strip-components 1 "firmware-${FIRMWARE_VERSION}/boot"
    SAVE ARTIFACT boot/bcm2711* /boot/
    SAVE ARTIFACT boot/fixup4.dat /boot/
    SAVE ARTIFACT boot/start4.elf /boot/
    SAVE ARTIFACT boot/overlays /boot/

rpi4-elemental-image:
    FROM DOCKERFILE -f Dockerfile.rpi4 .
    ARG K3S_VERSION=v1.24.3+k3s1

    # Install k3s.
    RUN curl -fsSL "https://raw.githubusercontent.com/k3s-io/k3s/${K3S_VERSION}/install.sh" | \
        INSTALL_K3S_BIN_DIR="/usr/sbin" \
        INSTALL_K3S_EXEC="" \
        INSTALL_K3S_SKIP_ENABLE="true" \
        INSTALL_K3S_VERSION="${K3S_VERSION}" \
        sh - \
        && echo -e '# This file will be overwritten by hcos-agent.\ncommand_args=""' > /etc/rancher/k3s/k3s.env \
        && rm /usr/sbin/k3s-killall.sh /usr/sbin/k3s-uninstall.sh
    # Create default hcos user.
    RUN adduser --uid 2000 --disabled-password --gecos hcos hcos \
        && adduser hcos wheel

    # Patch elemental-toolkit services to redirect their output to logfiles.
    FOR stage IN boot network
        RUN sed -i \
            "s/elemental.*/ebegin \"Running elemental $stage stage\"\n  \0 >> \/var\/log\/\${RC_SVCNAME}.log 2>\&1/" \
            /etc/init.d/cos-setup-$stage
    END

    COPY --dir +kernel-arm64/* /
    COPY +hcos-linux-arm64/hcos /usr/sbin/hcos
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
    WITH DOCKER --load hcos-rpi4:latest=+rpi4-elemental-image
        RUN --privileged ./build_image_rpi4.sh --docker-image hcos-rpi4:latest hcos-rpi4.img
    END
    SAVE ARTIFACT hcos-rpi4.img* AS LOCAL ./build/

go-deps:
    FROM golang:1.18.4-alpine3.16
    WORKDIR /build
    COPY go.mod go.sum ./
    RUN go mod download && go mod verify
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

hcos:
    FROM +go-deps
    ARG --required GOARCH
    ARG --required GOOS
    COPY --dir cmd internal pkg ./
    RUN go build -o hcos ./cmd/hcos/main.go
    SAVE ARTIFACT hcos

hcos-darwin-arm64:
    COPY (+hcos/hcos --GOARCH=arm64 --GOOS=darwin) .
    SAVE ARTIFACT hcos AS LOCAL ./build/hcos-darwin-arm64

hcos-linux-arm64:
    COPY (+hcos/hcos --GOARCH=arm64 --GOOS=linux) .
    SAVE ARTIFACT hcos AS LOCAL ./build/hcos-linux-arm64

hcos-linux-amd64:
    COPY (+hcos/hcos --GOARCH=amd64 --GOOS=linux) .
    SAVE ARTIFACT hcos AS LOCAL ./build/hcos-linux-amd64
