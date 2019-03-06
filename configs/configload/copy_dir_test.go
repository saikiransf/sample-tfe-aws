package configload

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyDir_symlinks(t *testing.T) {
	// this test sets up a $TMPDIR/modules/ directory with two submodules,
	// one being a symlink to the other
	tmpdir, err := ioutil.TempDir("", "copy-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	moduleDir := filepath.Join(tmpdir, "modules")
	err = os.Mkdir(moduleDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	subModuleDir := filepath.Join(moduleDir, "test-module")
	err = os.Mkdir(subModuleDir, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(subModuleDir, "main.tf"), []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink("test-module", filepath.Join(moduleDir, "symlink-module"))
	if err != nil {
		t.Fatal(err)
	}

	targetDir := filepath.Join(tmpdir, "target")
	os.Mkdir(targetDir, os.ModePerm)

	err = copyDir(targetDir, moduleDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = os.Lstat(filepath.Join(targetDir, "test-module", "main.tf")); os.IsNotExist(err) {
		t.Fatal("target test-module/main.tf was not created")
	}

	if _, err = os.Lstat(filepath.Join(targetDir, "symlink-module", "main.tf")); os.IsNotExist(err) {
		t.Fatal("target symlink-module/main.tf was not created")
	}
}
