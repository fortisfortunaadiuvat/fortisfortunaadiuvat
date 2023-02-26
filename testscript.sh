#!/usr/bin/env bash

set -euo pipefail
mkdir -p /opt/diary-flex-agent

diary-flex-agent-1d() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/opt/diary-flex-agent/diary-flex-agent-1d"
    MODULE_URL="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/$MODULE_FILENAME"
    MODULE_PARAMS=$(echo "$@")

    if [ $(( ( $(date +%s) - $(stat -L --format %Y $MODULE_PATH/$MODULE_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$MODULE_URL" -P "$MODULE_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary-report] Fetch module %s error on URL %s: diary-error-fetch-module-failed\n" "$MODULE_FILENAME" "$MODULE_URL" >&2
        fi
		chmod 0755 $MODULE_PATH/$MODULE_FILENAME
    fi

#	bash -c "$MODULE_PATH/$MODULE_FILENAME $MODULE_PARAMS"
}

diary-flex-agent-1h() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/opt/diary-flex-agent/diary-flex-agent-1h"
    MODULE_URL="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/$MODULE_FILENAME"
    MODULE_PARAMS=$(echo "$@")

    if [ $(( ( $(date +%s) - $(stat -L --format %Y $MODULE_PATH/$MODULE_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$MODULE_URL" -P "$MODULE_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary-report] Fetch module %s error on URL %s: diary-error-fetch-module-failed\n" "$MODULE_FILENAME" "$MODULE_URL" >&2
        fi
		chmod 0755 $MODULE_PATH/$MODULE_FILENAME
    fi

#	bash -c "$MODULE_PATH/$MODULE_FILENAME $MODULE_PARAMS"
}

diary-flex-agent-1m() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/opt/diary-flex-agent/diary-flex-agent-1m"
    MODULE_URL="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/$MODULE_FILENAME"
    MODULE_PARAMS=$(echo "$@")

    if [ $(( ( $(date +%s) - $(stat -L --format %Y $MODULE_PATH/$MODULE_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$MODULE_URL" -P "$MODULE_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary-report] Fetch module %s error on URL %s: diary-error-fetch-module-failed\n" "$MODULE_FILENAME" "$MODULE_URL" >&2
        fi
		chmod 0755 $MODULE_PATH/$MODULE_FILENAME
    fi

#	bash -c "$MODULE_PATH/$MODULE_FILENAME $MODULE_PARAMS"
}



main() {
    diary-flex-agent-1m

    diary-flex-agent-1h

    diary-flex-agent-1d

    
}
