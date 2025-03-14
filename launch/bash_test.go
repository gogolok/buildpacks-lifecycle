package launch_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"github.com/buildpacks/lifecycle/launch"
	hl "github.com/buildpacks/lifecycle/launch/testhelpers"
	h "github.com/buildpacks/lifecycle/testhelpers"
)

func TestBash(t *testing.T) {
	spec.Run(t, "Bash", testBash, spec.Report(report.Terminal{}))
}

func testBash(t *testing.T, when spec.G, it spec.S) {
	var (
		shell  launch.Shell
		tmpDir string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "shell-test")
		h.AssertNil(t, err)
		shell = &launch.BashShell{Exec: hl.SyscallExecWithStdout(t, tmpDir)}
	})

	it.After(func() {
		h.AssertNil(t, os.RemoveAll(tmpDir))
	})

	when("#Launch", func() {
		var process launch.ShellProcess

		when("script", func() {
			when("there are profiles", func() {
				it.Before(func() {
					process = launch.ShellProcess{
						Script:  true,
						Command: `printf "profile env: '%s'" "$PROFILE_VAR"`,
						Caller:  "some-profile-argv0",
						Env: []string{
							"SOME_VAR=some-val",
						},
					}
					process.Profiles = []string{
						filepath.Join("testdata", "profiles", "print_argv0"),
						filepath.Join("testdata", "profiles", "print_env"),
						filepath.Join("testdata", "profiles", "set_env"),
					}
				})

				it("sets argv0 for profile scripts", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "profile argv0: 'some-profile-argv0'")
				})

				it("sets env for profile scripts", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "SOME_VAR: 'some-val'")
				})

				it("env vars set in profile scripts are available to the command", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "profile env: 'some-profile-var'")
				})
			})

			it("sets env", func() {
				process = launch.ShellProcess{
					Script:  true,
					Command: `printf "SOME_VAR: '%s'" "$SOME_VAR"`,
					Caller:  "some-profile-argv0",
					Env: []string{
						"SOME_VAR=some-val",
					},
				}
				err := shell.Launch(process)
				h.AssertNil(t, err)
				stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
				if len(stdout) == 0 {
					stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
					t.Fatalf("stdout was empty: stderr: %s\n", stderr)
				}
				h.AssertStringContains(t, stdout, "SOME_VAR: 'some-val'")
			})

			it("provides args to bash", func() {
				process = launch.ShellProcess{
					Script:  true,
					Command: `printf "SOME_ARG: '%s'" "$1"`,
					Args:    []string{"bash", "some arg1"},
					Caller:  "some-profile-argv0",
					Env: []string{
						"SOME_VAR=some-val",
					},
				}
				err := shell.Launch(process)
				h.AssertNil(t, err)
				stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
				if len(stdout) == 0 {
					stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
					t.Fatalf("stdout was empty: stderr: %s\n", stderr)
				}
				h.AssertStringContains(t, stdout, "SOME_ARG: 'some arg1'")
			})

			it("handles many args", func() {
				process = launch.ShellProcess{
					Script:  false,
					Command: `echo`,
					Args: []string{
						"one",
						"two",
						"three",
						"four",
						"five",
						"six",
						"seven",
						"eight",
						"nine",
						"ten",
					},
					Caller: "some-profile-argv0",
					Env: []string{
						"SOME_VAR=some-val",
					},
				}
				err := shell.Launch(process)
				h.AssertNil(t, err)
				stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
				if len(stdout) == 0 {
					stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
					t.Fatalf("stdout was empty: stderr: %s\n", stderr)
				}
				h.AssertStringContains(t, stdout, "one two three four five six seven eight nine ten")
			})
		})

		when("is not script", func() {
			when("there are profiles", func() {
				var (
					err        error
					defaultDir string
				)

				it.Before(func() {
					defaultDir, err = os.Getwd()
					h.AssertNil(t, err)
					process = launch.ShellProcess{
						Script:  false,
						Command: "printf",
						Args:    []string{"profile env: '%s'", "$PROFILE_VAR"},
						Caller:  "some-profile-argv0",
						Env: []string{
							"SOME_VAR=some-val",
						},
					}
					process.Profiles = []string{
						filepath.Join("testdata", "profiles", "print_argv0"),
						filepath.Join("testdata", "profiles", "print_env"),
						filepath.Join("testdata", "profiles", "set_env"),
					}
				})

				it("runs the profiles from the default directory", func() {
					process.Profiles = []string{
						filepath.Join("testdata", "profiles", "pwd"),
					}
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					h.AssertStringContains(t, stdout, fmt.Sprintf("profile directory: %s", defaultDir))
				})

				it("runs the command in the specified working directory", func() {
					process.Command = "echo"
					process.Args = []string{"Process working directory: $(pwd)"}
					process.WorkingDirectory = tmpDir
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					h.AssertStringContains(t, stdout, fmt.Sprintf("Process working directory: %s", tmpDir))
				})

				it("sets argv0 for profile scripts", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "profile argv0: 'some-profile-argv0'")
				})

				it("sets env for profile scripts", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "SOME_VAR: 'some-val'")
				})

				it("env vars set in profile scripts are available to the command", func() {
					err := shell.Launch(process)
					h.AssertNil(t, err)
					stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
					if len(stdout) == 0 {
						stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
						t.Fatalf("stdout was empty: stderr: %s\n", stderr)
					}
					h.AssertStringContains(t, stdout, "profile env: 'some-profile-var'")
				})
			})

			it("sets env", func() {
				process = launch.ShellProcess{
					Script:  false,
					Command: `printf`,
					Args:    []string{"SOME_VAR: '%s'", "$SOME_VAR"},
					Caller:  "some-profile-argv0",
					Env: []string{
						"SOME_VAR=some-val",
					},
				}
				err := shell.Launch(process)
				h.AssertNil(t, err)
				stdout := rdfile(t, filepath.Join(tmpDir, "stdout"))
				if len(stdout) == 0 {
					stderr := rdfile(t, filepath.Join(tmpDir, "stderr"))
					t.Fatalf("stdout was empty: stderr: %s\n", stderr)
				}
				h.AssertStringContains(t, stdout, "SOME_VAR: 'some-val'")
			})
		})
	})
}
