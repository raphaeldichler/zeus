// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package runtime

import (
	"testing"
	"time"
)

const (
	cmdRunBackground = `trap "exit 0" TERM; tail -f /dev/null & wait`
)

func pullStartAndRunAlping(cmd string) (*Container, error) {
	return NewContainer(
		"testing",
		"docker-runtime",
		WithImage("alpine:3.14"),
		WithPulling(),
		WithCmd("sh", "-c", cmd),
	)
}

func assertPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected panic, but function did not panic")
		}
	}()
	fn()
}

func assertNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic: %v", r)
		}
	}()
	fn()
}

func assertPathNotExist(
	t *testing.T,
	container *Container,
	path string,
) {
	exists, err := container.ExitsPath(path)
	if err != nil {
		t.Errorf("checking path exists failed, got %q", err)
	}
	if exists {
		t.Errorf("checking path  exists returned true, but file not should exist")
	}
}

func assertContainerNotRuns(
	t *testing.T,
	container *Container,
) {
	runs, err := container.IsRunning()
	if err != nil {
		t.Errorf("checking container runs failed, got %q", err)
	}
	if runs {
		t.Errorf("checking container running returned true, but it should not run")
	}
}

func assertContainerRuns(
	t *testing.T,
	container *Container,
) {
	runs, err := container.IsRunning()
	if err != nil {
		t.Errorf("checking container runs failed, got %q", err)
	}
	if !runs {
		t.Errorf("checking container running returned false, but it should run")
	}
}

func assertPathExist(
	t *testing.T,
	container *Container,
	path string,
) {
	exists, err := container.ExitsPath(path)
	if err != nil {
		t.Errorf("checking path exists failed, got %q", err)
	}
	if !exists {
		t.Errorf("checking path exists returned false, but file should exist")
	}
}

func assertFileRead(
	t *testing.T,
	container *Container,
	path string,
	content string,
) {
	read, err := container.ReadFile(path)
	if err != nil {
		t.Errorf("reading file failed, got %q", err)
	}
	if read != content {
		t.Errorf("data read differs than stored. expected '%q', got '%q'", content, read)
	}
}

func TestCheckCopyReadSubsequently(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	path, data := "file.txt", "foobra"
	assertPathNotExist(t, cont, path)

	f := BasicFileContent{
		Path:    path,
		Content: []byte(data),
	}
	err = cont.CopyInto(&f)
	if err != nil {
		t.Fatalf("coping data failed, got %q", err)
	}

	assertPathExist(t, cont, path)
	assertFileRead(t, cont, path, data)
}

func TestReadFileInContainer(t *testing.T) {
	cont, err := pullStartAndRunAlping(
		`touch /tmp/file.txt && echo -n "foobar" > /tmp/file.txt && ` + cmdRunBackground,
	)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	assertFileRead(t, cont, "/tmp/file.txt", "foobar")
}

func TestEnsurePathExists(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	path := "/tmp/this-should-really-not-exists"
	assertPathNotExist(t, cont, path)

	if err := cont.EnsurePathExists(path); err != nil {
		t.Errorf("failed to ensure path, got %q", err)
	}

	assertPathExist(t, cont, path)
}

func TestCopyInto(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	data := "foobra"
	path := "/tmp/this-should-really-not-exists.txt"
	assertPathNotExist(t, cont, path)

	f := BasicFileContent{
		Path:    path,
		Content: []byte(data),
	}
	if err := cont.CopyInto(&f); err != nil {
		t.Errorf("failed to copy data into container, got %q", err)
	}

	assertPathExist(t, cont, path)
	assertFileRead(t, cont, path, data)
}

func TestCopyIntoOverwrite(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	data1 := "foobra"
	path := "/tmp/this-should-really-not-exists.txt"
	assertPathNotExist(t, cont, path)

	f1 := BasicFileContent{
		Path:    path,
		Content: []byte(data1),
	}
	if err := cont.CopyInto(&f1); err != nil {
		t.Errorf("failed to copy data into container, got %q", err)
	}

	assertPathExist(t, cont, path)
	assertFileRead(t, cont, path, data1)

	data2 := "hello world"
	f2 := BasicFileContent{
		Path:    path,
		Content: []byte(data2),
	}
	if err := cont.CopyInto(&f2); err != nil {
		t.Errorf("failed to copy data into container, got %q", err)
	}

	assertPathExist(t, cont, path)
	assertFileRead(t, cont, path, data2)
}

func TestIsRunning(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	assertContainerRuns(t, cont)
	err = cont.Shutdown()
	if err != nil {
		t.Errorf("failed shutdown container, got %q", err)
	}

	assertContainerNotRuns(t, cont)
}

func TestAssertIsRunning(t *testing.T) {
	cont, err := pullStartAndRunAlping(cmdRunBackground)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}

	assertContainerRuns(t, cont)
	assertNoPanic(t, func() {
		err := cont.AssertIsRunning()
		if err != nil {
			t.Errorf("failed to assert container is running. could not check if running, err %q", err)
		}
	})

	err = cont.Shutdown()
	if err != nil {
		t.Errorf("failed shutdown container, err %q", err)
	}

	assertContainerNotRuns(t, cont)
	assertPanic(t, func() {
		err := cont.AssertIsRunning()
		if err != nil {
			t.Errorf("failed to assert container is running. could not check if running, err %q", err)
		}
	})
}

func TestAssertPathExists(t *testing.T) {
	cont, err := pullStartAndRunAlping(
		`touch /tmp/file.txt && echo -n "foobar" > /tmp/file.txt && ` + cmdRunBackground,
	)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}
	defer cont.Shutdown()

	assertNoPanic(t, func() {
		err := cont.AssertPathExists("/tmp")
		if err != nil {
			t.Errorf("failed to assert path exists. could not check if running, err %q", err)
		}
	})

	assertNoPanic(t, func() {
		err := cont.AssertPathExists("/tmp/file.txt")
		if err != nil {
			t.Errorf("failed to assert is running. could not check if running, err %q", err)
		}
	})

	assertPanic(t, func() {
		err := cont.AssertPathExists("/tmp/this-should-really-not-exists")
		if err != nil {
			t.Errorf("failed to assert path exists. could not check if running, err %q", err)
		}
	})
}

func TestSighup(t *testing.T) {
	cont, err := pullStartAndRunAlping(
		`trap 'echo \"Received SIGHUP. Exiting...\"; exit 0' SIGHUP; while true; do sleep 1; done`,
	)
	if err != nil {
		t.Fatalf("failed starting container, got %q", err)
	}

	assertContainerRuns(t, cont)

	if err := cont.Sighup(); err != nil {
		t.Errorf("failed to send sighup, err %q", err)
	}
	time.Sleep(time.Second * 2)

	existsContainer, err := existsContaienr(cont.id)
	if err != nil {
		t.Errorf("failed check if container exists, err %q", err)
	}

	if existsContainer {
		t.Errorf("container should stopped and removed, but still found it")
	}
}
