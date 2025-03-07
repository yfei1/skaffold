/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"path/filepath"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestSupportedKubernetesFormats(t *testing.T) {
	var tests = []struct {
		description string
		in          string
		out         bool
	}{
		{
			description: "yaml",
			in:          "filename.yaml",
			out:         true,
		},
		{
			description: "yml",
			in:          "filename.yml",
			out:         true,
		},
		{
			description: "json",
			in:          "filename.json",
			out:         true,
		},
		{
			description: "txt",
			in:          "filename.txt",
			out:         false,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := IsSupportedKubernetesFormat(test.in)

			t.CheckDeepEqual(test.out, actual)
		})
	}
}

func TestExpandPathsGlob(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	tmpDir.Write("dir/sub_dir/file", "")
	tmpDir.Write("dir_b/sub_dir_b/file", "")

	var tests = []struct {
		description string
		in          []string
		out         []string
		shouldErr   bool
	}{
		{
			description: "match exact filename",
			in:          []string{"dir/sub_dir/file"},
			out:         []string{tmpDir.Path("dir/sub_dir/file")},
		},
		{
			description: "match leaf directory glob",
			in:          []string{"dir/sub_dir/*"},
			out:         []string{tmpDir.Path("dir/sub_dir/file")},
		},
		{
			description: "match top level glob",
			in:          []string{"dir*"},
			out:         []string{tmpDir.Path("dir/sub_dir/file"), tmpDir.Path("dir_b/sub_dir_b/file")},
		},
		{
			description: "invalid pattern",
			in:          []string{"[]"},
			shouldErr:   true,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual, err := ExpandPathsGlob(tmpDir.Root(), test.in)

			t.CheckErrorAndDeepEqual(test.shouldErr, err, test.out, actual)
		})
	}
}

func TestExpand(t *testing.T) {
	var tests = []struct {
		description string
		text        string
		key         string
		value       string
		expected    string
	}{
		{
			description: "${key} syntax",
			text:        "BEFORE[${key}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE]AFTER",
		},
		{
			description: "$key syntax",
			text:        "BEFORE[$key]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE]AFTER",
		},
		{
			description: "replace all",
			text:        "BEFORE[$key][${key}][$key][${key}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[VALUE][VALUE][VALUE][VALUE]AFTER",
		},
		{
			description: "ignore common prefix",
			text:        "BEFORE[$key1][${key1}]AFTER",
			key:         "key",
			value:       "VALUE",
			expected:    "BEFORE[$key1][${key1}]AFTER",
		},
		{
			description: "just the ${key} placeholder",
			text:        "${key}",
			key:         "key",
			value:       "VALUE",
			expected:    "VALUE",
		},
		{
			description: "just the $key placeholder",
			text:        "$key",
			key:         "key",
			value:       "VALUE",
			expected:    "VALUE",
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := Expand(test.text, test.key, test.value)

			t.CheckDeepEqual(test.expected, actual)
		})
	}
}

func TestAbsFile(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()
	tmpDir.Write("file", "")
	expectedFile, err := filepath.Abs(filepath.Join(tmpDir.Root(), "file"))
	testutil.CheckError(t, false, err)

	file, err := AbsFile(tmpDir.Root(), "file")
	testutil.CheckErrorAndDeepEqual(t, false, err, expectedFile, file)

	_, err = AbsFile(tmpDir.Root(), "")
	testutil.CheckErrorAndDeepEqual(t, true, err, tmpDir.Root()+" is a directory", err.Error())

	_, err = AbsFile(tmpDir.Root(), "does-not-exist")
	testutil.CheckError(t, true, err)
}

func TestNonEmptyLines(t *testing.T) {
	var tests = []struct {
		in  string
		out []string
	}{
		{"", nil},
		{"a\n", []string{"a"}},
		{"a\r\n", []string{"a"}},
		{"a\r\nb", []string{"a", "b"}},
		{"a\r\nb\n\n", []string{"a", "b"}},
		{"\na\r\n\n\n", []string{"a"}},
	}
	for _, test := range tests {
		testutil.Run(t, "", func(t *testutil.T) {
			result := NonEmptyLines([]byte(test.in))

			t.CheckDeepEqual(test.out, result)
		})
	}
}

func TestCloneThroughJSON(t *testing.T) {
	tests := []struct {
		description string
		old         interface{}
		new         interface{}
		expected    interface{}
	}{
		{
			description: "google cloud build",
			old: map[string]string{
				"projectId": "unit-test",
			},
			new: &latest.GoogleCloudBuild{},
			expected: &latest.GoogleCloudBuild{
				ProjectID: "unit-test",
			},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			err := CloneThroughJSON(test.old, test.new)

			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected, test.new)
		})
	}
}

func TestIsHiddenDir(t *testing.T) {
	tests := []struct {
		description string
		filename    string
		expected    bool
	}{
		{
			description: "hidden dir",
			filename:    ".hidden",
			expected:    true,
		},
		{
			description: "not hidden dir",
			filename:    "not_hidden",
			expected:    false,
		},
		{
			description: "current dir",
			filename:    ".",
			expected:    false,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			isHidden := IsHiddenDir(test.filename)

			t.CheckDeepEqual(test.expected, isHidden)
		})
	}
}

func TestIsHiddenFile(t *testing.T) {
	tests := []struct {
		description string
		filename    string
		expected    bool
	}{
		{
			description: "hidden file name",
			filename:    ".hidden",
			expected:    true,
		},
		{
			description: "not hidden file",
			filename:    "not_hidden",
			expected:    false,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			isHidden := IsHiddenDir(test.filename)

			t.CheckDeepEqual(test.expected, isHidden)
		})
	}
}

func TestRemoveFromSlice(t *testing.T) {
	testutil.CheckDeepEqual(t, []string{""}, RemoveFromSlice([]string{""}, "ANY"))
	testutil.CheckDeepEqual(t, []string{"A", "B", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "ANY"))
	testutil.CheckDeepEqual(t, []string{"A", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "B"))
	testutil.CheckDeepEqual(t, []string{"B", "C"}, RemoveFromSlice([]string{"A", "B", "C"}, "A"))
	testutil.CheckDeepEqual(t, []string{"A", "C"}, RemoveFromSlice([]string{"A", "B", "B", "C"}, "B"))
	testutil.CheckDeepEqual(t, []string{}, RemoveFromSlice([]string{"B", "B"}, "B"))
}
