package label_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/jx-gitops/pkg/cmd/label"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateLabelsInYamlFiles(t *testing.T) {
	sourceData := filepath.Join("test_data")
	files, err := ioutil.ReadDir(sourceData)
	assert.NoError(t, err)

	argTests := [][]string{
		{"beer=stella", "wine=merlot"},
		{"wine=merlot", "beer=stella"},
	}

	for _, args := range argTests {
		tmpDir, err := ioutil.TempDir("", "")
		require.NoError(t, err, "could not create temp dir")

		type testCase struct {
			SourceFile   string
			ResultFile   string
			ExpectedFile string
		}

		var testCases []testCase
		for _, f := range files {
			if f.IsDir() {
				name := f.Name()
				srcFile := filepath.Join(sourceData, name, "source.yaml")
				expectedFile := filepath.Join(sourceData, name, "expected.yaml")
				require.FileExists(t, srcFile)
				require.FileExists(t, expectedFile)

				outFile := filepath.Join(tmpDir, name+".yaml")
				err = util.CopyFile(srcFile, outFile)
				require.NoError(t, err, "failed to copy %s to %s", srcFile, outFile)

				testCases = append(testCases, testCase{
					SourceFile:   srcFile,
					ResultFile:   outFile,
					ExpectedFile: expectedFile,
				})
			}
		}
		err = label.UpdateLabelArgsInYamlFiles(tmpDir, args)
		require.NoError(t, err, "failed to update namespace in dir %s for args %#v", tmpDir, args)

		for _, tc := range testCases {
			resultData, err := ioutil.ReadFile(tc.ResultFile)
			require.NoError(t, err, "failed to load results %s for args %#v", tc.ResultFile, args)

			expectData, err := ioutil.ReadFile(tc.ExpectedFile)
			require.NoError(t, err, "failed to load results %s for args %#v", tc.ExpectedFile, args)

			result := strings.TrimSpace(string(resultData))
			expectedText := strings.TrimSpace(string(expectData))
			if d := cmp.Diff(result, expectedText); d != "" {
				t.Errorf("Generated Pipeline for file %s did not match expected: %s for args %#v", tc.SourceFile, d, args)
			}
			t.Logf("generated for file %s with args %#v file\n%s\n", tc.SourceFile, args, result)
		}
	}
}
