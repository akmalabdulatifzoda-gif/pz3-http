# Использование: .\run.ps1 run | build | test
param([ValidateSet("run","build","test")][string]$Target = "run")
if (-not $env:PORT) { $env:PORT = "8080" }
switch ($Target) {
  "run"   { go run ./cmd/server }
  "build" { go build -o .\bin\server.exe .\cmd\server }
  "test"  { go test ./... -v }
}
