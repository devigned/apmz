// Code generated for package data by go-bindata DO NOT EDIT. (@generated)
// sources:
// data/disabled_bash.sh
// data/enabled_bash.sh
package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)
type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _dataDisabled_bashSh = []byte(`#!/usr/bin/env bash

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
}`)

func dataDisabled_bashShBytes() ([]byte, error) {
	return _dataDisabled_bashSh, nil
}

func dataDisabled_bashSh() (*asset, error) {
	bytes, err := dataDisabled_bashShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/disabled_bash.sh", size: 236, mode: os.FileMode(420), modTime: time.Unix(1578612908, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataEnabled_bashSh = []byte(`#!/usr/bin/env bash

__TMP_APMZ_BATCH_FILE="$(mktemp /tmp/apmz.XXXXXX)"
__SCRIPT_START_TIME=$(apmz time unixnano)
__SCRIPT_NAME="${__SCRIPT_NAME:-{{.ScriptName}}}"
__APP_INSIGHTS_KEY="${__APP_INSIGHTS_KEY:-{{.AppInsightsKey}}}"
__DEFAULT_TAGS="${__DEFAULT_TAGS:-{{.DefaultTags}}}"

# trace_err will log an error level trace event to the tmp batch file in $TMP_APMZ_BATCH_FILE
#
# should be invoked in the following way: `+"`"+`trace_err "trace_name" "tag1,tag2,tag3"`+"`"+`
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
# should be invoked in the following way: `+"`"+`trace_info "trace_name" "tag1,tag2,tag3"`+"`"+`
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
# should be invoked in the following way: `+"`"+`time_metric "metric_name" fuction_to_time(...)`+"`"+`
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
# should be invoked in the following way: `+"`"+`time_metric "metric_name" "tag1=value,tag2=value" fuction_to_time(...)`+"`"+`
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
# should be invoked in the following way: `+"`"+`append_default_tags "${tags}"`+"`"+`
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
    rm "${__TMP_APMZ_BATCH_FILE}"
  fi
}

trap exitAndFlush EXIT
`)

func dataEnabled_bashShBytes() ([]byte, error) {
	return _dataEnabled_bashSh, nil
}

func dataEnabled_bashSh() (*asset, error) {
	bytes, err := dataEnabled_bashShBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/enabled_bash.sh", size: 3779, mode: os.FileMode(420), modTime: time.Unix(1578699667, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"data/disabled_bash.sh": dataDisabled_bashSh,
	"data/enabled_bash.sh":  dataEnabled_bashSh,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"data": &bintree{nil, map[string]*bintree{
		"disabled_bash.sh": &bintree{dataDisabled_bashSh, map[string]*bintree{}},
		"enabled_bash.sh":  &bintree{dataEnabled_bashSh, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
