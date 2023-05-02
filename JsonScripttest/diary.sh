#!/usr/bin/env bash

# Title       : cloud-platform-diary
# Author      : Jason Alexander Dasuki
# Date        : 07-March-2023
# Description : SDK for shell script for infrastructure analytics

PLATFORM_DIARY_PROD_URL="http://diary.tokopedia.net"
PLATFORM_DIARY_STAG_URL="http://diary-staging.tokopedia.net"
PLATFORM_DIARY_DEV_URL="http://localhost:9000"
PLATFORM_DIARY_URL=""
PLATFORM_DIARY_SOURCE="$(hostname)"
PLATFORM_DIARY_SOURCE_SDK="shell"
PLATFORM_DIARY_DATACENTER=""
PLATFORM_DIARY_ENVIRONMENT="${PLATFORM_DIARY_ENVIRONMENT:-}"
PLATFORM_DIARY_TYPE="${1:-"null"}"
PLATFORM_DIARY_SOURCE_TYPE="${2:-"null"}"
PLATFORM_DIARY_ACTOR="${3:-"null"}"
PLATFORM_DIARY_STATUS="${4:-"null"}"
PLATFORM_OPERATION_COMPANY="${5:-""}"
PLATFORM_DIARY_MESH="${6:-"default"}"
PLATFORM_DIARY_DETAILS="${7:-"{}"}"
PLATFORM_DIARY_TIME_DURATION=${8:-}

if [ -f /etc/ansible-host-info/datacenter ]; then
    PLATFORM_DIARY_DATACENTER="$(cat /etc/ansible-host-info/datacenter)"
else
    PLATFORM_DIARY_DATACENTER="null"
fi

if [ -z "$PLATFORM_DIARY_ENVIRONMENT" ]; then
    if [ -f /etc/ansible-host-info/env ]; then
        PLATFORM_DIARY_ENVIRONMENT="$(cat /etc/ansible-host-info/env)"
    elif [ -n "$TKPENV" ]; then
        PLATFORM_DIARY_ENVIRONMENT="$TKPENV"
    fi
fi

case $PLATFORM_DIARY_ENVIRONMENT in
  production)
    PLATFORM_DIARY_URL="$PLATFORM_DIARY_PROD_URL" ;;
  staging)
    PLATFORM_DIARY_URL="$PLATFORM_DIARY_STAG_URL" ;;
  development)
    PLATFORM_DIARY_URL="$PLATFORM_DIARY_DEV_URL" ;;
  *)
    PLATFORM_DIARY_URL="$PLATFORM_DIARY_PROD_URL" ;;
esac

if [ $PLATFORM_DIARY_TIME_DURATION == "null" ]; then
    PLATFORM_DIARY_TIME_DURATION=0
else
    PLATFORM_DIARY_TIME_DURATION=$PLATFORM_DIARY_TIME_DURATION
fi

EVENT=$( jq -n \
        --arg diary_event_type "$PLATFORM_DIARY_TYPE" \
        --arg diary_event_source "$PLATFORM_DIARY_SOURCE" \
        --arg diary_event_source_type "$PLATFORM_DIARY_SOURCE_TYPE" \
        --arg diary_event_source_sdk "$PLATFORM_DIARY_SOURCE_SDK" \
        --arg diary_event_actor "$PLATFORM_DIARY_ACTOR" \
        --arg diary_event_status "$PLATFORM_DIARY_STATUS" \
        --arg diary_event_operating_company "$PLATFORM_OPERATION_COMPANY" \
        --arg diary_event_environment "$PLATFORM_DIARY_ENVIRONMENT" \
        --arg diary_event_datacenter "$PLATFORM_DIARY_DATACENTER" \
        --arg diary_event_mesh "$PLATFORM_DIARY_MESH" \
        --argjson diary_event_details "$PLATFORM_DIARY_DETAILS" \
        --argjson diary_event_time_duration $PLATFORM_DIARY_TIME_DURATION \
        '{
            "diaryEventType": $diary_event_type,
            "diaryEventSource": $diary_event_source,
            "diaryEventSourceType": $diary_event_source_type,
            "diaryEventSourceSdk": $diary_event_source_sdk,
            "diaryEventActor": $diary_event_actor,
            "diaryEventStatus": $diary_event_status,
            "diaryEventOperatingCompany": $diary_event_operating_company,
            "diaryEventEnvironment": $diary_event_environment,
            "diaryEventDatacenter": $diary_event_datacenter,
            "diaryEventMesh": $diary_event_mesh,
            "diaryEventDetails": $diary_event_details,
            "diaryEventTimeDuration": $diary_event_time_duration,
        }' )

send_event() {
    echo "[$(date)] Pushed Event to Cloud Platform Diary"

    curl \
        --retry 2 \
        --retry-max-time 5 \
        --connect-timeout 1 \
        --request POST $PLATFORM_DIARY_URL/api/diary \
        --header 'Content-Type: application/json' \
        --data "$EVENT" &
}

debug_event_params() {
    echo "[$(date)] Debug Event parameter Cloud Platform Diary"

    echo $EVENT
}
