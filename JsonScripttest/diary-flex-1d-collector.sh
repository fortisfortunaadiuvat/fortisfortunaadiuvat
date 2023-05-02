#!/usr/bin/env bash

# Title       : cloud-platform-diary SDK for Flex 1d Collector
# Author      : Jason Alexander Dasuki
# Date        : 07-March-2023
# Description : Script used for send flex data to Cloud Platform Diary

#Collector Script URLs
OS_UNAME_COLLECTOR="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex/diary-flex-collector/os-uname-collector.sh"

diary_report() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/tmp/.diary"
    MODULE_URL="https://raw.githubusercontent.com/fortisfortunaadiuvat/fortisfortunaadiuvat/main/JsonScripttest/$MODULE_FILENAME"
    MODULE_PARAMS=$(echo "$@")

    if [ $(( ( $(date +%s) - $(stat -L --format %Y $MODULE_PATH/$MODULE_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$MODULE_URL" -P "$MODULE_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary-flex-1d-collector] Fetch module %s error on URL %s: diary-flex-1d-collector-error-fetch-module-failed\n" "$MODULE_FILENAME" "$MODULE_URL" >&2
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

collect_os_uname() {
	OS_UNAME_STR="$(fetch_module $OS_UNAME_COLLECTOR)"
	OS_UNAME_DATA=( $OS_UNAME_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex" \
                "diaryEventSourceType=diary_flex_os_uname" \
                "diaryEventActor=diary-flex-1d-collector.sh" \
                "kernelName=${OS_UNAME_DATA[0]}" \
                "kernelRelease=${OS_UNAME_DATA[2]}" \
		"kernelVersion=${OS_UNAME_DATA[3]}_${OS_UNAME_DATA[4]}_${OS_UNAME_DATA[5]}_${OS_UNAME_DATA[6]}_${OS_UNAME_DATA[7]}_${OS_UNAME_DATA[8]}_${OS_UNAME_DATA[9]}_${OS_UNAME_DATA[10]}" \
		"machine=${OS_UNAME_DATA[11]}" \
		"processor=${OS_UNAME_DATA[12]}" \
		"hardwarePlatform=${OS_UNAME_DATA[13]}" \
		"operatingSystem=${OS_UNAME_DATA[14]}"
}

collect_os_dpkg() {
        dpkg_output="$(dpkg -l | grep collec | tail -n +6)"
        hostname="$(hostname)"
	OS_DPKG_STR=$(awk '{print "{\"Package\":\""$2"\",\"Version\":\""$3"\",\"Architecture\":\""$4"\"}"}' <<< "$dpkg_output" | jq '.' -c )
        OS_DPKG_DATA=( $OS_DPKG_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex" \
                "diaryEventSourceType=diary_flex_os_dpkg" \
                "diaryEventActor=diary-flex-1d-collector.sh" \
                "$OS_DPKG_STR"
}

main() {
	#Collect uname
	collect_os_uname

        #Collect DPKH
        collect_os_dpkg
}

main "$@"
