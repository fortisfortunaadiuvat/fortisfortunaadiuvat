#!/usr/bin/env bash

# Title       : cloud-platform-diary SDK for Flex 1d Collector
# Author      : Jason Alexander Dasuki
# Date        : 07-March-2023
# Description : Script used for send flex data to Cloud Platform Diary

#Collector Script URLs
UNAME_COLLECTOR="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex-agent/diary-flex-agent-collector/uname-collector.sh"

diary_report() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/tmp/.diary"
    MODULE_URL="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex-agent/$MODULE_FILENAME"
    MODULE_PARAMS=$(echo "$@")

    if [ $(( ( $(date +%s) - $(stat -L --format %Y $MODULE_PATH/$MODULE_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$MODULE_URL" -P "$MODULE_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary-flex-agent-1d-collector] Fetch module %s error on URL %s: diary-error-fetch-module-failed\n" "$MODULE_FILENAME" "$MODULE_URL" >&2
        fi
                chmod 0755 $MODULE_PATH/$MODULE_FILENAME
    fi

        bash -c "$MODULE_PATH/$MODULE_FILENAME $MODULE_PARAMS"
}

fetch_module() {
        if [ $# -lt 1 ]; then
                printf "Fetch module usage: %s MODULE_URL [MODULE_ARGS...]" "$0"
                return 1
        fi

        MODULE_URL="$1"
        MODULE_FILENAME="${MODULE_URL##*/}"
        MODULE_PARAMS=$(echo "$@" | cut -s -f 2- -d ' ')

        wget --retry-connrefused --wait=15 --tries=15 --retry-on-http-error=404,503 -N "$MODULE_URL"
        bash ./$MODULE_FILENAME $MODULE_PARAMS
}

collect_uname() {
	UNAME_STR="$(fetch_module $UNAME_COLLECTOR)"
	UNAME_DATA=( $UNAME_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex_agent_run" \
                "diaryEventSourceType=diary_flex_agent" \
                "diaryEventActor=diary-flex-agent-1d-collector.sh" \
                "kernelName=${UNAME_DATA[0]}" \
                "kernelRelease=${UNAME_DATA[2]}" \
		"kernelVersion=${UNAME_DATA[3]}_${UNAME_DATA[4]}_${UNAME_DATA[5]}_${UNAME_DATA[6]}_${UNAME_DATA[7]}_${UNAME_DATA[8]}_${UNAME_DATA[9]}_${UNAME_DATA[10]}" \
		"machine=${UNAME_DATA[11]}" \
		"processor=${UNAME_DATA[12]}" \
		"hardwarePlatform=${UNAME_DATA[13]}" \
		"operatingSystem=${UNAME_DATA[14]}"
}

main() {
	#Collect uname data
	collect_uname
}

main "$@"
