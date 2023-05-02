#!/usr/bin/env bash

# Title       : cloud-platform-diary
# Author      : Jason Alexander Dasuki
# Date        : 07-March-2023
# Description : Script used for send event log to Cloud Platform Diary

sync_diary() {
    # parameter variable
    JSON_DATA=($@)
    JSON_DATA_SOURCE="${JSON_DATA[@]:-'null=null'}"
    JSON_CUSTOM_DATA=$(jq -R 'split(" ") | map( index("=") as $i | {(.[0:$i]) : .[$i+1:]}) | add' <<< "$JSON_DATA_SOURCE")

 #   if [[ "$JSON_DATA_SOURCE" == *"="* ]]; then
 #       # format 1: space-separated key-value pairs
 #       JSON_CUSTOM_DATA=$(jq -R 'split(" ") | map( index("=") as $i | {(.[0:$i]) : .[$i+1:]}) | add' <<< "$JSON_DATA_SOURCE")
 #   else
 #       # format 2: JSON object
 #       JSON_CUSTOM_DATA=$(jq -c '.' <<< "$JSON_DATA_SOURCE")
 #   fi

    DIARY_EVENT_TIME_DURATION=$(jq -r '.diaryEventTimeDuration // "0" | tonumber' <<< "$JSON_CUSTOM_DATA")
    DIARY_EVENT_STATUS=$(jq -r '.diaryEventStatus' <<< "$JSON_CUSTOM_DATA")
    if [ "$DIARY_EVENT_STATUS" == 0 ]; then
        JSON_RESULT_DATA=$(jq --argjson DURATION $DIARY_EVENT_TIME_DURATION '.diaryEventStatus = "success" | .diaryEventTimeDuration = $DURATION' <<< "$JSON_CUSTOM_DATA")
    else
        JSON_RESULT_DATA=$(jq --argjson DURATION $DIARY_EVENT_TIME_DURATION '.diaryEventStatus = "failed" | .diaryEventTimeDuration = $DURATION' <<< "$JSON_CUSTOM_DATA")
    fi

    JSON_DIARY_CUSTOM_DATA=$(echo "$JSON_RESULT_DATA" | jq -c .)
    JSON_DIARY_TYPE=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventType' | sed -r 's/[/.]+/_/g')
    JSON_DIARY_SOURCE_TYPE=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventSourceType')
    JSON_DIARY_ACTOR=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventActor')
    JSON_DIARY_STATUS=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventStatus')
    JSON_DIARY_OPERATION_COMPANY=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventOperatingCompany')
    JSON_DIARY_MESH=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventMesh')
    JSON_DIARY_TIME_DURATION=$(echo "$JSON_RESULT_DATA" | jq -r '.diaryEventTimeDuration')
    SDK_FILENAME="diary.sh"
    SDK_PATH="/tmp/.diary"
    SDK_URL="https://raw.githubusercontent.com/fortisfortunaadiuvat/fortisfortunaadiuvat/main/JsonScripttest/$SDK_FILENAME"

    # set new args for send event to cloud-platform-diary
    set -- $JSON_DIARY_TYPE $JSON_DIARY_SOURCE_TYPE $JSON_DIARY_ACTOR $JSON_DIARY_STATUS $JSON_DIARY_OPERATION_COMPANY $JSON_DIARY_MESH $JSON_DIARY_CUSTOM_DATA $JSON_DIARY_TIME_DURATION
    EVENT_DIARY_PARAMS="$@"

    # source diary-sdk-shell
    if [ $(( ( $(date +%s) - $(stat -L --format %Y $SDK_PATH/$SDK_FILENAME 2>/dev/null || printf '0') ) > 60 )) -eq 1 ]; then
        wget --wait=3 --tries=2 -N "$SDK_URL" -P "$SDK_PATH" 2>/dev/null
        if [ $? -ne 0 ]; then
            printf "[diary] Fetch module %s error on URL %s: diary-error-fetch-sdk-failed\n" "$SDK_FILENAME" "$SDK_URL" >&2
        fi
        sudo chmod 0755 $SDK_PATH/$SDK_FILENAME
    fi
    source "$SDK_PATH/$SDK_FILENAME"

    ## send event to cloud-platform-diary
    debug_event_params 
}

sync_diary "$@"
