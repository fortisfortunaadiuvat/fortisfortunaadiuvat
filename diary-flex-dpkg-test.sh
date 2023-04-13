#!/usr/bin/env bash

# Title       : cloud-platform-diary SDK for Flex 1d Collector
# Author      : Jason Alexander Dasuki
# Date        : 07-March-2023
# Description : Script used for send flex data to Cloud Platform Diary

#Collector Script URLs
OS_UNAME_COLLECTOR="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex/diary-flex-collector/os-uname-collector.sh"
OS_USER_COLLECTOR="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex/diary-flex-collector/os-user-collector.sh"
OS_LSB_RELEASE_COLLECTOR="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex/diary-flex-collector/os-lsb-release-collector.sh"

diary_report() {
    MODULE_FILENAME="cloud-platform-diary.sh"
    MODULE_PATH="/tmp/.diary"
    MODULE_URL="https://tokopedia-dpkg.s3.ap-southeast-1.amazonaws.com/cloudplatform/cloud-platform-diary/diary-flex/$MODULE_FILENAME"
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

collect_os_user() {
	OS_USER_STR="$(fetch_module $OS_USER_COLLECTOR)"
	OS_USER_DATA=( $OS_USER_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex" \
                "diaryEventSourceType=diary_flex_os_user" \
                "diaryEventActor=diary-flex-1d-collector.sh" \
                "listUsers=${OS_USER_DATA}" 
}

collect_os_lsb_release() {
        OS_LSB_RELEASE_STR="$(fetch_module $OS_LSB_RELEASE_COLLECTOR)"
        OS_LSB_RELEASE_DATA=( $OS_LSB_RELEASE_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex" \
                "diaryEventSourceType=diary_flex_os_lsb_release" \
                "diaryEventActor=diary-flex-1d-collector.sh" \
                "distributorId=${OS_LSB_RELEASE_DATA[2]}" \
                "description=${OS_LSB_RELEASE_DATA[4]}_${OS_LSB_RELEASE_DATA[5]}_${OS_LSB_RELEASE_DATA[6]}" \
                "release=${OS_LSB_RELEASE_DATA[8]}" \
                "codename=${OS_LSB_RELEASE_DATA[10]}"
}

collect_os_dpkg() {
        dpkg_output="$(dpkg -l | tail -n +6)"
        hostname="$(hostname)"
        # Use awk to parse the output of dpkg -l and select the columns you want to include in the JSON output
        #awk '{print "{\"Package\":\""$2"\",\"Version\":\""$3"\",\"Architecture\":\""$4"\",\"Description\":\""$5"\"}"}' <<< "$dpkg_output" |
        # Use jq to convert the awk output to JSON format
	packages=$(awk '{print "{\"Package\":\""$2"\",\"Version\":\""$3"\",\"Architecture\":\""$4"\",\"Description\":\""$5"\"}"}' <<< "$dpkg_output" | jq -s '{Packages: .}')
        #packages="$(jq -s '{Hostname: "'"$HOSTNAME"'", Packages: .}' -)"
        #OS_LSB_RELEASE_STR="$(fetch_module $OS_LSB_RELEASE_COLLECTOR)"
        #OS_LSB_RELEASE_DATA=( $OS_LSB_RELEASE_STR )
        diary_report \
                "diaryEventStatus=$?" \
                "diaryEventType=diary_flex" \
                "diaryEventSourceType=diary_flex_os_dpkg" \
                "diaryEventActor=diary-flex-1d-collector.sh" \
                "listUsers=${packages}"
}

main() {
	#Collect uname
	collect_os_uname

	#Collect user from /etc/passwd
        collect_os_user

	#Collect os version from lsb_release
	collect_os_lsb_release
	
	collect_os_dpkg
}

main "$@"
