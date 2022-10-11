#!/bin/bash
# Build an HCOS image for Raspberry Pi 4 from the provided container image. This script is a rewritten and simplified
# version of https://github.com/rancher-sandbox/cOS-toolkit/blob/95430e481977c0bd69316f6e5427f2b7d455e1ed/images/arm-img-builder.sh
# Some of the main differences:
#  * active.img is not duplicated as passive.img and recovery.img to not increase the size of the the final image.
#    Instead, they are going to be copied on the first system boot.
#  * The RPi system boot and recovery partitions are populated with files from the container image. The image build
#    process contains all the logic of how these files are prepared, not this script.
set -e

load_vars() {
  # Img creation options. Size is in MB for all of the vars below
  image_size="${IMAGE_SIZE:-2560}"
  state_part_size="${STATE_PART_SIZE:-1536}"
  recovery_part_size="${RECOVERY_PART_SIZE:-512}"
  active_img_size="${ACTIVE_IMG_SIZE:-512}"

  # Warning: these default values must be aligned with the values provided
  # in 'packages/cos-config/cos-config', provide an environment file using the
  # --cos-config flag if different values are needed.
  : "${OEM_LABEL:=COS_OEM}"
  : "${RECOVERY_LABEL:=COS_RECOVERY}"
  : "${ACTIVE_LABEL:=COS_ACTIVE}"
  : "${PASSIVE_LABEL:=COS_PASSIVE}"
  : "${PERSISTENT_LABEL:=COS_PERSISTENT}"
  : "${SYSTEM_LABEL:=COS_SYSTEM}"
  : "${STATE_LABEL:=COS_STATE}"
}

cleanup() {
  if [ -n "$active_img_mnt" ]; then
    umount $active_img_mnt &> /dev/zero || true
  fi
  if [ -n "$system_boot_part_mnt" ]; then
    umount $system_boot_part_mnt &> /dev/zero || true
  fi
  if [ -n "$state_part_mnt" ]; then
    umount $state_part_mnt &> /dev/zero || true
  fi
  if [ -n "$recovery_part_mnt" ]; then
    umount $recovery_part_mnt &> /dev/zero || true
  fi
  if [ -n "$persistent_part_mnt" ]; then
    umount $persistent_part_mnt &> /dev/zero || true
  fi
  if [ -n "$WORKDIR" ]; then
    rm -rf $WORKDIR
  fi
  if [ -n "$image_loop_dev" ]; then
    kpartx -dv "$image_loop_dev"
  fi
  losetup -D || true
}

ensure_dir_structure() {
    local target=$1
    for mnt in /sys /proc /dev /tmp /boot /usr/local /oem
    do
        if [ ! -d "${target}${mnt}" ]; then
          mkdir -p "${target}${mnt}"
        fi
    done
}

usage()
{
    echo "Usage: $0 [options] image.img"
    echo ""
    echo "Example: $0 --cos-config cos-config --docker-image <image> output.img"
    echo ""
    echo "Flags:"
    echo " --docker-image: A container image which will be used for active/passive/recovery_part system"
    echo " --cos-config: (optional) Specifies a cos-config file for required environment variables"
    echo " --config: (optional) Specify a cloud-init config file to embed into the final image"
    exit 1
}

get_url()
{
    FROM=$1
    TO=$2
    case $FROM in
        ftp*|http*|tftp*)
            n=0
            attempts=5
            until [ "$n" -ge "$attempts" ]
            do
                curl -o $TO -fL ${FROM} && break
                n=$((n+1))
                echo "Failed to download, retry attempt ${n} out of ${attempts}"
                sleep 2
            done
            ;;
        *)
            cp -f $FROM $TO
            ;;
    esac
}

trap "cleanup" 1 2 3 6 9 14 15 EXIT

load_vars

while [ "$#" -gt 0 ]; do
    case $1 in
        --cos-config)
            shift 1
            cos_config=$1
            ;;
        --config)
            shift 1
            config=$1
            ;;
        --docker-image)
            shift 1
            container_image=$1
            ;;
        -h)
            usage
            ;;
        --help)
            usage
            ;;
        *)
            if [ "$#" -gt 2 ]; then
                usage
            fi
            output_image=$1
            break
            ;;
    esac
    shift 1
done

if [ -n "$cos_config"] && [ -e "$cos_config" ]; then
  source "$cos_config"
fi

if [ -z "$container_image" ]; then
  echo "Container image (--docker-image) must be specified."
  exit 1
fi

if [ -z "$output_image" ]; then
  echo "No image file specified"
  exit 1
fi

echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "Image size: $image_size MB."
echo "Recovery partition: $recovery_part_size MB."
echo "State partition: $state_part_size MB."
echo "Images (active/passive/recovery_part.img) size: $active_img_size MB."
echo "Container image: $container_image"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"

# Temp dir used during build
WORKDIR=$(mktemp -d --tmpdir image_build.XXXXXXXXXX)

echo ">> Creating image and partition table"
dd if=/dev/zero of="${output_image}" bs=1M count="${image_size}" || exit 1
sgdisk -n 1:8192:+96M -c 1:system-boot -t 1:0c00 "${output_image}"
sgdisk -n 2:0:+${state_part_size}M -c 2:state -t 2:8300 "${output_image}"
sgdisk -n 3:0:+${recovery_part_size}M -c 3:recovery -t 3:8300 "${output_image}"
sgdisk -n 4:0:+64M -c 4:persistent -t 4:8300 "${output_image}"
# Convert GPT to MBR for backwards compatibility. GPT partition table is supported out of box on RPi4 after upgrading
# the EEPROM.
sgdisk -m 1:2:3:4 "${output_image}"
# Change the first partition (system-boot) type to W95 FAT32 (LBA).
sfdisk --part-type "${output_image}" 1 c

# Prepare the image and copy over the files.
image_loop_dev=$(losetup -f "${output_image}" --show)
if [ -z "${image_loop_dev}" ]; then
	echo "Cannot execute losetup for $output_image"
	exit 1
fi
image_loop_dev_basename="${image_loop_dev/\/dev\//}"
# Add partition mappings for the image loop device.
kpartx -va "$image_loop_dev"
mapped_image_loop_dev="/dev/mapper/${image_loop_dev_basename}"

system_boot_part="${mapped_image_loop_dev}p1"
state_part="${mapped_image_loop_dev}p2"
recovery_part="${mapped_image_loop_dev}p3"
persistent_part="${mapped_image_loop_dev}p4"

# Format partitions (BOOT, STATE, RECOVERY, PERSISTENT).
mkfs.vfat -F 32 "${system_boot_part}"
fatlabel "${system_boot_part}" HCOS_BOOT
mkfs.ext4 -F -L "${STATE_LABEL}" "$state_part"
mkfs.ext4 -F -L "${RECOVERY_LABEL}" "$recovery_part"
mkfs.ext4 -F -L "${PERSISTENT_LABEL}" "$persistent_part"

echo ">> Populating partitions"
system_boot_part_mnt="$WORKDIR/system-boot"
state_part_mnt="$WORKDIR/state"
recovery_part_mnt="$WORKDIR/recovery"
active_img_mnt="$WORKDIR/active"
mkdir "$system_boot_part_mnt"
mkdir "$state_part_mnt"
mkdir "$recovery_part_mnt"
mkdir "$active_img_mnt"

mount "$system_boot_part" "$system_boot_part_mnt"
mount "$state_part" "$state_part_mnt"
mount "$recovery_part" "$recovery_part_mnt"

echo ">> Preparing active.img"
mkdir -p "${state_part_mnt}/cOS"
active_img="${state_part_mnt}/cOS/active.img"
dd if=/dev/zero of="$active_img" bs=1M count="$active_img_size"
mkfs.ext2 -L "${ACTIVE_LABEL}" "$active_img"
mount -o loop -t ext2 "$active_img" "$active_img_mnt"
ensure_dir_structure "$active_img_mnt"

echo ">>> Downloading container image"
elemental pull-image --local "$container_image" "$active_img_mnt"

echo ">> Preparing GRUB (system-boot) partition"
rsync -av "${active_img_mnt}/.system-boot/" "$system_boot_part_mnt/"
rm -rf "${active_img_mnt}/.system-boot"

echo ">> Preparing recovery partition"
mkdir -p "${recovery_part_mnt}/cOS"
# Do not duplicate active.img as recovery.img to not increase the size of the the final image. Instead, do the copy
# on the first system boot.

# Install real GRUB config to recovery partition.
rsync -av "${active_img_mnt}/.cos-recovery/" "$recovery_part_mnt/"
rm -rf "${active_img_mnt}/.cos-recovery"

# Set an OEM config file if specified.
if [ -n "$config" ]; then
  echo ">> Copying $config OEM config file"
  persistent_part_mnt="$WORKDIR/persistent"
  mkdir "$persistent_part_mnt"
  mount "$persistent_part" "$persistent_part_mnt"
  mkdir "$persistent_part_mnt/cloud-config"
  get_url $config "$persistent_part_mnt/cloud-config/99_custom.yaml"
  umount "$persistent_part_mnt"
fi

umount "$active_img_mnt"
umount "$system_boot_part_mnt"
umount "$state_part_mnt"
umount "$recovery_part_mnt"
sync

# Delete partition mappings /dev/mapper/loopXpY for the image loop device.
kpartx -dv "$image_loop_dev"
# Detach the image loop device /dev/loopX.
losetup -d "$image_loop_dev"

echo ">> Done writing $output_image"
echo ">> Creating SHA256 sum"
sha256sum "$output_image" > "$output_image.sha256"
