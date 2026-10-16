package sohocloud

// JFA conformance bindings for the coordination protocol.
//
// Test names carry the Janus registry ID they bind (jfa-conformance-suite.py
// in NTARI-RAND/Janus), so `jfa-conformance-suite.py --bindings` and a CI log
// parser can map a PASS/FAIL back to the invariant: TestInvariant_<ID>_<Name>,
// with any "-" in the ID written "_".

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestInvariant_P_lean_code_StdlibOnlyCopyleft binds registry ID P-lean-code:
// "Protocol software stays small, depends on nothing but its language's
// standard library, and is auditable whole" — enforced as a dependency audit
// in CI and a copyleft check.
//
// The dependency-leaf workflow already fails the build when go.mod gains a
// requirement. This binding states the same rule as a test the registry can
// see, and closes what a go.mod check alone cannot:
//
//   - go.mod declares no require, replace, exclude, or retract, and there is
//     no go.sum and no vendor directory;
//   - every Go file — tests included — imports only the standard library or
//     this module's own packages, and none uses cgo (import "C"), which would
//     pull in C the auditor must also read;
//   - the independent Python reproducer imports only the Python standard
//     library — every name in a comma list, and no dynamic imports — so a
//     foreign implementer needs nothing but a stock interpreter;
//   - the software is copyleft: LICENSE is the GNU AGPL v3, every license
//     file in any folder is the AGPL, and no comment declares any other
//     license;
//   - the audit actually runs in CI: ci.yml runs `go test ./...` (which runs
//     this test) and the Python reproducer, and the dependency-leaf workflow
//     is still present and still fails the build. Residual, named: a
//     workflow's effect cannot be proven from its text; these checks only
//     raise the bar.
func TestInvariant_P_lean_code_StdlibOnlyCopyleft(t *testing.T) {
	t.Run("go.mod declares no dependencies", testLeanGoMod)
	t.Run("Go imports are stdlib or this module only", testLeanGoImports)
	t.Run("the Python reproducer is stdlib only", testLeanPython)
	t.Run("the software is AGPL-3.0 copyleft", testLeanCopyleft)
	t.Run("the audit runs in CI", testLeanCI)
}

func modulePath(t *testing.T) string {
	t.Helper()
	f, err := os.Open("go.mod")
	if err != nil {
		t.Fatalf("open go.mod: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if fields := strings.Fields(sc.Text()); len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	t.Fatal("go.mod declares no module path")
	return ""
}

func testLeanGoMod(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for i, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "require", "replace", "exclude", "retract":
			t.Errorf("go.mod:%d: %q — the protocol must depend on nothing but the Go standard library", i+1, strings.TrimSpace(line))
		}
	}
	for _, p := range []string{"go.sum", "vendor"} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("%s exists — a dependency leaf has nothing to sum or vendor", p)
		}
	}
}

func testLeanGoImports(t *testing.T) {
	mod := modulePath(t)
	fset := token.NewFileSet()
	scanned := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if n := info.Name(); n == ".git" || (n != "." && strings.HasPrefix(n, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		scanned++
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			switch {
			case p == "C":
				t.Errorf("%s uses cgo (import \"C\"): the protocol must be auditable as Go alone", path)
			case p == mod || strings.HasPrefix(p, mod+"/"):
				// this module's own packages
			case !strings.Contains(strings.SplitN(p, "/", 2)[0], "."):
				// standard library: no domain in the first path element
			default:
				t.Errorf("%s imports %s: the protocol may import only the Go standard library and itself", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if scanned < 20 {
		t.Fatalf("scanned only %d Go files — the audit is not looking at the module", scanned)
	}
}

// pythonStdlib is the Python standard-library modules the reproducer may use.
// A module outside this list fails the test until a reviewer confirms it is
// stdlib and adds it — the review moment for any new import.
var pythonStdlib = map[string]bool{
	"sys": true, "os": true, "json": true, "hashlib": true, "struct": true,
	"base64": true, "binascii": true, "re": true, "io": true, "pathlib": true,
	"argparse": true, "itertools": true, "functools": true, "collections": true,
	"typing": true, "unicodedata": true, "datetime": true, "math": true,
	"string": true, "textwrap": true, "hmac": true, "secrets": true,
}

var (
	pyImportLine = regexp.MustCompile(`^\s*import\s+(.+)$`)
	pyFromLine   = regexp.MustCompile(`^\s*from\s+([A-Za-z_][A-Za-z0-9_.]*)\s+import(\s|$)`)
	pyDynamic    = regexp.MustCompile(`__import__\s*\(|(^|[^A-Za-z0-9_])importlib`)
)

// pyModules returns the top-level module names a line imports, covering
// "import a, b.c as d" and "from a.b import x".
func pyModules(line string) []string {
	if m := pyFromLine.FindStringSubmatch(line); m != nil {
		return []string{strings.SplitN(m[1], ".", 2)[0]}
	}
	m := pyImportLine.FindStringSubmatch(line)
	if m == nil {
		return nil
	}
	var out []string
	for _, part := range strings.Split(m[1], ",") {
		if f := strings.Fields(strings.TrimSpace(part)); len(f) > 0 {
			out = append(out, strings.SplitN(f[0], ".", 2)[0])
		}
	}
	return out
}

func testLeanPython(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.py"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("found no Python reproducer in testdata; the audit is not looking at it")
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "#") && pyDynamic.MatchString(line) {
				t.Errorf("%s:%d uses a dynamic import: the reproducer's imports must be auditable as plain import statements", path, i+1)
			}
			for _, name := range pyModules(line) {
				if !pythonStdlib[name] {
					t.Errorf("%s:%d imports %q, which is not in pythonStdlib: the reproducer must run on a stock interpreter", path, i+1, name)
				}
			}
		}
	}
}

// spdx matches a license tag anywhere on a comment line (Go, Python, shell,
// YAML, Markdown or C-style), so "// Copyright … <tag>: MIT" counts, while the
// pattern below and prose outside comments do not count as declarations.
var spdx = regexp.MustCompile(`(?m)^[ \t]*(?://|#|<!--|/\*|\*).*?SPDX-License-Identifier:[ \t]*([A-Za-z0-9.+-]+)`)

// licenseFile matches a license file name at any depth: LICENSE, LICENCE or
// COPYING, with or without an extension.
var licenseFile = regexp.MustCompile(`^(?i)(LICENSE|LICENCE|COPYING)(\..*)?$`)

func testLeanCopyleft(t *testing.T) {
	lic, err := os.ReadFile("LICENSE")
	if err != nil {
		t.Fatalf("read LICENSE: %v", err)
	}
	head := string(lic)
	if len(head) > 400 {
		head = head[:400]
	}
	if !strings.Contains(head, "GNU AFFERO GENERAL PUBLIC LICENSE") || !strings.Contains(head, "Version 3") {
		t.Error("LICENSE is not the GNU Affero General Public License, version 3")
	}
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if n := info.Name(); n == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if path == "LICENSE" || info.Size() > 1<<20 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if licenseFile.MatchString(info.Name()) && !strings.Contains(string(data), "GNU AFFERO GENERAL PUBLIC LICENSE") {
			t.Errorf("%s is a license file that is not the GNU AGPL: protocol software is AGPL-3.0 only, in every folder", path)
		}
		for _, m := range spdx.FindAllStringSubmatch(string(data), -1) {
			if !strings.HasPrefix(m[1], "AGPL-3.0") {
				t.Errorf("%s declares SPDX license %s: protocol software is AGPL-3.0 only", path, m[1])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func testLeanCI(t *testing.T) {
	ci, err := os.ReadFile(filepath.Join(".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	for _, want := range []string{"go test ./...", "reproduce_spec.py"} {
		if !strings.Contains(string(ci), want) {
			t.Errorf("ci.yml no longer runs %q: the dependency audit must run in CI", want)
		}
	}
	leaf, err := os.ReadFile(filepath.Join(".github", "workflows", "dependency-leaf.yml"))
	if err != nil {
		t.Fatalf("dependency-leaf workflow is gone: %v", err)
	}
	if !strings.Contains(string(leaf), "go list -m all") {
		t.Error("dependency-leaf.yml no longer lists module dependencies")
	}
	// Residual, named: a workflow's effect cannot be proven from its text.
	// Requiring the failing exit raises the bar; it does not close it.
	if !strings.Contains(string(leaf), "exit 1") {
		t.Error("dependency-leaf.yml no longer fails the build (no exit 1) when a dependency appears")
	}
}
