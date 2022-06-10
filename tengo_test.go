package caddytengo

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddytest"
)

func TestCaddyfileConfig(t *testing.T) {
	var testdir = filepath.Join("testdata", "caddyfiles")
	for _, fi := range sourceFiles(t, testdir, ".caddyfile") {
		t.Run(fi.Name(), func(t *testing.T) {
			srcFile := filepath.Join(testdir, fi.Name())
			config := fileContent(t, srcFile)
			wantFile, errFile := srcFile+".want", srcFile+".err"

			if *testUpdateConfigTests {
				success, failure := adaptConfig(t, config, "caddyfile")
				if failure == nil {
					if err := os.WriteFile(wantFile, success, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if failure != nil {
					// marshal as JSON so quotes are escaped in the message, as that is
					// how it will be compared in AssertLoadError. Then remove the
					// wrapping quotes.
					b, _ := json.Marshal(failure.Error())
					if err := os.WriteFile(errFile, b[1:len(b)-1], 0o600); err != nil {
						t.Fatal(err)
					}
				}
				return
			}

			var asserted bool
			if want := fileContent(t, filepath.Join(testdir, fi.Name()+".want")); want != nil {
				caddytest.AssertAdapt(t, string(config), "caddyfile", string(want))
				asserted = true
			}
			if err := fileContent(t, filepath.Join(testdir, fi.Name()+".err")); err != nil {
				caddytest.AssertLoadError(t, string(config), "caddyfile", string(err))
				asserted = true
			}
			if !asserted {
				t.Fatal("no expected result")
			}
		})
	}
}

var testUpdateConfigTests = flag.Bool("test.update-config-tests", false, "If set, update the expected test results with the actual results.")

// returns the file content or nil if it does not exist
func fileContent(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	return bytes.TrimSpace(b)
}

// sourceFiles returns the list of source files in dir corresponding to the
// specified extension.
func sourceFiles(t *testing.T, dir, ext string) []os.FileInfo {
	t.Helper()

	if ext != "" && ext[0] != '.' {
		ext = "." + ext
	}

	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	res := make([]os.FileInfo, 0, len(fis))
	for _, fi := range fis {
		if !fi.Mode().IsRegular() {
			continue
		}
		if ext != "" && filepath.Ext(fi.Name()) != ext {
			continue
		}
		res = append(res, fi)
	}
	return res
}

func adaptConfig(t *testing.T, config []byte, adapterName string) ([]byte, error) {
	cfgAdapter := caddyconfig.GetAdapter(adapterName)
	if cfgAdapter == nil {
		t.Fatalf("unrecognized config adapter '%s'", adapterName)
	}

	options := make(map[string]interface{})
	result, _, err := cfgAdapter.Adapt(config, options)
	if err != nil {
		return nil, err
	}

	// prettify results to keep tests human-manageable
	var prettyBuf bytes.Buffer
	err = json.Indent(&prettyBuf, result, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	result = prettyBuf.Bytes()
	return result, nil
}
