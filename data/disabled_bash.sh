#!/usr/bin/env bash

trace_err() {
    return
}

trace_info() {
    return
}

time_metric() {
    shift
    "$@"
}

time_metric_with_tags() {
  shift
  shift
  "$@"
}

append_default_tags() {
  return
}

default_apmz_tags() {
  return
}