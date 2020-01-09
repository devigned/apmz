#! /bin/bash
set -eu -o pipefail

echo "SOURCE ${SOURCE}"
ls -alh ${SOURCE}

az login --identity
az storage container create -n ${STORAGE_CONTAINER}
az storage blob upload-batch --source "${SOURCE}" --destination "${STORAGE_CONTAINER}" --pattern "${PATTERN}" --destination-path "${DESTINATION_PATH}"