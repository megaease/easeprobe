#!/usr/bin/env bash

pushd "$(dirname $0)" > /dev/null || exit
SCRIPT_PATH=$(pwd -P)
popd > /dev/null || exit

build_image() {
    local img_name=$1
    local img_ver=$2
    local docker_context_path=$3
    local docker_file=$4

    DOCKER_BUILDKIT=1 docker build -t "${img_name}":"${img_ver}" -f "${docker_file}" "${docker_context_path}"
}

show_usage() {
   printf "usage: build_image.sh [-i IMAGE_NAME] [-v IMAGE_VERSION] -d ROOT_REPOSITORY_DIR [-f DOCKER_FILE] \n"
   printf "\t-i IMAGE_NAME the name of building image, default is megaease/easestash\n"
   printf "\t-v IMAGE_VERSION the version of building image, default is 0.1.0-alpine\n"
   printf "\t-d ROOT_REPOSITORY_DIR the root directory of repository\n"
   printf "\t-f DOCKER_FILE the location of Dockerfile, if it's omitted, default location is ROOT_REPOSITORY_DIR/resources/rootfs/Dockerfile\n"
   printf ""
}

if [ "${IMAGE_NAME:-x}" == "x" ];
then
   IMAGE_NAME=megaease/easeprobe
fi

if [ "${IMAGE_VERSION:-alpine-0.0.1}" != "alpine-0.0.1" ];
then
   IMAGE_VERSION="alpine-${IMAGE_VERSION}"
fi

IMAGE_VERSION=${IMAGE_VERSION:-alpine-0.0.1}


while getopts 'hi:v:d:f:' flag; do
  case "${flag}" in
    h) show_usage; exit ;;
    i) IMAGE_NAME=${OPTARG};;
    v) IMAGE_VERSION=${OPTARG};;
    d) REPOSITORY_DIR=${OPTARG};;
    f) DOCKER_FILE=${OPTARG};;
    *) show_usage; exit ;;
  esac
done

if [ -z "${REPOSITORY_DIR}" ]; then
   show_usage
   exit 1
fi

if [ -z "${DOCKER_FILE}" ];
then
    DOCKER_FILE=${REPOSITORY_DIR}/Dockerfile
fi


build_image "${IMAGE_NAME}" \
	"${IMAGE_VERSION}" \
	"${REPOSITORY_DIR}/" \
   "${DOCKER_FILE}"