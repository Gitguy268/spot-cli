package hook

import "testing"

func TestIsTest(t *testing.T) {
	m := NewMatcher(nil, []string{"./scripts/test.sh"}, []string{"go test -run XXX"})
	yes := []string{
		"go test ./...",
		"go test -race -count=1 ./internal/...",
		"/usr/local/go/bin/go test .",
		"CGO_ENABLED=0 go test ./...",
		"sudo -E go test ./...",
		"time cargo test",
		"env -i FOO=1 pytest -x",
		"npm test",
		"npm run test:unit",
		"pnpm run test",
		"bun test",
		"npx vitest run",
		"python3 -m pytest tests/",
		"uv run pytest",
		"cargo nextest run",
		"make test-unit",
		"make -j8 test",
		"./gradlew test",
		"bundle exec rspec spec/models",
		"mix test",
		"dotnet test",
		"zig build test",
		"cd api && go test ./... && cd ..",
		"go build ./... && go test ./...",
		"go test ./... | tee out.txt",
		"./scripts/test.sh",
		"(cd web; npm test)",
	}
	no := []string{
		"",
		"git status",
		"echo test",
		"go build ./...",
		"go vet ./...",
		"npm install",
		"npm run build",
		"cat test.txt",
		"vim internal/test.go",
		"testing",
		"make",
		"make build",
		"go test -run XXX ./...",
		"spot run -- go test ./...",
		"FOO=bar",
		"cargo build --tests",
	}
	for _, c := range yes {
		if !m.IsTest(c) {
			t.Errorf("IsTest(%q) = false, want true", c)
		}
	}
	for _, c := range no {
		if m.IsTest(c) {
			t.Errorf("IsTest(%q) = true, want false", c)
		}
	}
}

func TestReplaceDefaults(t *testing.T) {
	m := NewMatcher([]string{"task check"}, nil, nil)
	if m.IsTest("go test ./...") || !m.IsTest("task check --fast") {
		t.Fatal("replace did not replace")
	}
}

func FuzzIsTest(f *testing.F) {
	for _, s := range []string{"go test ./...", "a && b || c; d | e &", "=x", "\x00", "sudo"} {
		f.Add(s)
	}
	m := NewMatcher(nil, nil, nil)
	f.Fuzz(func(t *testing.T, s string) { m.IsTest(s) })
}

func TestCanonical(t *testing.T) {
	for in, want := range map[string]string{
		"BREAK=1 go test ./...":      "go test ./...",
		"sudo -E  go   test ./...":   "go test ./...",
		"cd api && go test ./...":    "cd api && go test ./...",
		"time CGO_ENABLED=0 go test": "go test",
	} {
		if got := Canonical(in); got != want {
			t.Errorf("Canonical(%q) = %q, want %q", in, got, want)
		}
	}
}
