#!/bin/bash

WIKI_URL=$(curl -s -I https://en.wikipedia.org/wiki/Special:Random | grep -i "location:" | cut -d " " -f2 | tr -d '\r')

TODO_TEXT="Read ${WIKI_URL}"

curl -X POST -H "Content-Type: application/json" \
     -d "{\"text\": \"$TODO_TEXT\"}" \
     http://todo-backend-svc:2345/todos
