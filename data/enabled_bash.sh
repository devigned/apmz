#!/usr/bin/env bash

__TMP_APMZ_BATCH_FILE="$(mktemp /tmp/apmz.XXXXXX)"
__SCRIPT_START_TIME=$(apmz time unixnano)
__SCRIPT_NAME="${__SCRIPT_NAME:-{{.ScriptName}}}"
__APP_INSIGHTS_KEY="${__APP_INSIGHTS_KEY:-{{.AppInsightsKey}}}"
__DEFAULT_TAGS="${__DEFAULT_TAGS:-{{.DefaultTags}}}"

# trace_err will log an error level trace event to the tmp batch file in $TMP_APMZ_BATCH_FILE
#
# should be invoked in the following way: `trace_err "trace_name" "tag1,tag2,tag3"`
trace_err() {
  name=$1
  tags=$2
  tags=$(append_default_tags "${tags}")
  if [[ -z "${tags}" ]]; then
    apmz trace -n "${name}" -l 3 -o >>"${__TMP_APMZ_BATCH_FILE}"
  else
    apmz trace -n "${name}" -l 3 -t "${tags}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  fi
}

# trace_info will log an info level trace event to the tmp batch file in $TMP_APMZ_BATCH_FILE
#
# should be invoked in the following way: `trace_info "trace_name" "tag1,tag2,tag3"`
trace_info() {
  name=$1
  tags=$2
  tags=$(append_default_tags "${tags}")
  if [[ -z "${tags}" ]]; then
    apmz trace -n "${name}" -l 0 -o >>"${__TMP_APMZ_BATCH_FILE}"
  else
    apmz trace -n "${name}" -t "${tags}" -l 0 -o >>"${__TMP_APMZ_BATCH_FILE}"
  fi
}

# time_metric will log a custom metric event to the tmp batch file in $TMP_APMZ_BATCH_FILE
#
# should be invoked in the following way: `time_metric "metric_name" fuction_to_time(...)`
time_metric() {
  name=$1
  shift
  start=$(apmz time unixnano)
  "$@"
  end=$(apmz time unixnano)
  diff=$(apmz time diff -a "${start}" -b "${end}")
  tags=$(append_default_tags)
  if [[ -z "${tags}" ]]; then
    apmz metric -n "${name}" -v "${diff}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  else
    apmz metric -n "${name}" -v "${diff}" -t "${tags}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  fi
}

# time_metric_with_tags will log a custom metric event to the tmp batch file in $TMP_APMZ_BATCH_FILE
#
# should be invoked in the following way: `time_metric "metric_name" "tag1=value,tag2=value" fuction_to_time(...)`
time_metric_with_tags() {
  name=$1
  shift
  tags=$1
  shift
  start=$(apmz time unixnano)
  "$@"
  end=$(apmz time unixnano)
  diff=$(apmz time diff -a "${start}" -b "${end}")
  tags=$(append_default_tags "${tags}")
  if [[ -z "${tags}" ]]; then
    apmz metric -n "${name}" -v "${diff}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  else
    apmz metric -n "${name}" -v "${diff}" -t "${tags}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  fi
}

# append_default_tags will append default_apmz_tags to the input tags string
#
# should be invoked in the following way: `append_default_tags "${tags}"`
append_default_tags() {
  input_tags=$1
  if [[ -z "${input_tags}" && -z "$(default_apmz_tags)" ]]; then
    echo ""
  elif [[ -z "${input_tags}" ]]; then
    default_apmz_tags
  else
    echo "${input_tags},${__DEFAULT_TAGS}"
  fi
}

# default_apmz_tags will return the __DEFAULT_TAGS and is called from append_default_tags
#
# feel free to override this function with your own
default_apmz_tags() {
  echo "${__DEFAULT_TAGS}"
}

exitAndFlush() {
  tags=$(append_default_tags "code=$?")
  if [[ "$?" == "0" ]]; then
    trace_info "$__SCRIPT_NAME-exit" "${tags}"
  else
    trace_err "$__SCRIPT_NAME-exit" "${tags}"
  fi

  script_end=$(apmz time unixnano)
  duration=$(apmz time diff -a "$__SCRIPT_START_TIME" -b "$script_end")
  tags=$(default_apmz_tags)
  if [[ -z "${tags}" ]]; then
    apmz metric -n "script-duration" -v "${duration}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  else
    apmz metric -n "script-duration" -v "${duration}" -t "${tags}" -o >>"${__TMP_APMZ_BATCH_FILE}"
  fi

  if [[ -n "${__APP_INSIGHTS_KEY}" && -z "${__DRY_RUN}" ]]; then
    apmz batch -f "${__TMP_APMZ_BATCH_FILE}" --api-key "${__APP_INSIGHTS_KEY}"
  fi

  if [[ -z "${__PRESERVE_TMP_FILE}" ]]; then
    echo deleted
    rm "${__TMP_APMZ_BATCH_FILE}"
  fi
}

trap exitAndFlush EXIT
