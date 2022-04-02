VERSION 0.6
FROM alpine:3.15.3

rpi4-cos-image:
    FROM DOCKERFILE -f Dockerfile.rpi4 .
    SAVE IMAGE hcos-rpi4:latest

rpi4-image-deps:
    ENV LUET_VERSION=0.30.3
    RUN apk --no-cache add \
        bash \
        coreutils \
        curl \
        dosfstools \
        e2fsprogs-extra \
        git \
        multipath-tools \
        parted \
        sgdisk \
        util-linux
    RUN wget -O - https://raw.githubusercontent.com/rancher-sandbox/cOS-toolkit/0fff8c1a642ff9a51f00aa4354715319c406a84a/scripts/get_luet.sh | sh
    RUN luet install -y \
        toolchain/elemental-cli \
        toolchain/yq \
        utils/jq

rpi4-image:
    FROM +rpi4-image-deps

    COPY --dir cOS-toolkit .
    # arm-img-builder.sh script expects it is run within a git repo so create a fake one.
    RUN git init ./cOS-toolkit
    WITH DOCKER --load hcos-rpi4:latest=+rpi4-cos-image
        RUN --privileged cd cOS-toolkit && ./images/arm-img-builder.sh --model rpi64 --docker-image hcos-rpi4:latest hcos-rpi4.img
    END
    SAVE ARTIFACT cOS-toolkit/hcos-rpi4.img* AS LOCAL .
