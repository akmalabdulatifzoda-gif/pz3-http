#!/usr/bin/env bash
# Прогон всех тестовых запросов из requests.md. Сервер должен быть запущен: go run ./cmd/server
# Использование: ./scripts/demo.sh 1   (части 1–4 — по одному скриншоту на часть)
B=${BASE_URL:-http://localhost:8080}
run() { printf '\033[1;32m$\033[0m %s\n' "$*"; eval "$@"; echo; echo; }
J='-H "Content-Type: application/json"'
case "${1:-all}" in
1|all) run curl -sS -i $B/health
       run curl -sS -i -X POST $B/tasks $J -d "'{\"title\":\"Купить молоко\"}'"
       run curl -sS -X POST $B/tasks $J -d "'{\"title\":\"Сделать ПЗ №3\"}'"
       run curl -sS -X POST $B/tasks $J -d "'{\"title\":\"Подготовить отчёт\"}'" ;;&
2|all) run curl -sS -i $B/tasks
       run curl -sS -i -G $B/tasks --data-urlencode "'q=молоко'"
       run curl -sS -i $B/tasks/1 ;;&
3|all) run curl -sS -i -X POST $B/tasks $J -d "'{}'"
       run curl -sS -i -X POST $B/tasks $J -d "'{\"title\":\"ab\"}'"
       run curl -sS -i $B/tasks/abc
       run curl -sS -i $B/tasks/9999 ;;&
4|all) run curl -sS -i -X PATCH $B/tasks/2 $J -d "'{\"done\":true}'"
       run curl -sS -i -X DELETE $B/tasks/1
       run curl -sS $B/tasks
       run curl -sS -i -X OPTIONS $B/tasks -H "'Origin: http://localhost:3000'" ;;
esac
