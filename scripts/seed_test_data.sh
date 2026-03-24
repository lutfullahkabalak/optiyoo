#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
REGISTER_URL="$BASE_URL/api/register"
LOGIN_URL="$BASE_URL/api/login"
SURVEY_URL="$BASE_URL/api/surveys"

USER_COUNT=10
SURVEY_COUNT=20
PASSWORD="123456"

json_field() {
  local json="$1"
  local field="$2"
  python3 -c 'import json,sys; data=json.loads(sys.argv[1]); print(data.get(sys.argv[2], ""))' "$json" "$field"
}

request_json() {
  local method="$1"
  local url="$2"
  local payload="$3"
  curl -sS -X "$method" "$url" \
    -H "Content-Type: application/json" \
    -d "$payload"
}

echo "Seeding users and surveys on: $BASE_URL"

declare -a USER_IDS=()

for i in $(seq 1 "$USER_COUNT"); do
  email="testuser${i}@optiyoo.local"
  name="Test User ${i}"

  register_payload=$(cat <<EOF
{"email":"$email","password":"$PASSWORD","name":"$name"}
EOF
)

  register_resp="$(request_json "POST" "$REGISTER_URL" "$register_payload" || true)"
  user_id="$(json_field "$register_resp" "id" || true)"

  # If user already exists, try login and reuse existing account.
  if [[ -z "$user_id" ]]; then
    login_payload=$(cat <<EOF
{"email":"$email","password":"$PASSWORD"}
EOF
)
    login_resp="$(request_json "POST" "$LOGIN_URL" "$login_payload" || true)"
    user_id="$(json_field "$login_resp" "id" || true)"
  fi

  if [[ -z "$user_id" ]]; then
    echo "Kullanıcı oluşturulamadı/giriş yapılamadı: $email"
    echo "Register response: $register_resp"
    exit 1
  fi

  USER_IDS+=("$user_id")
  echo "User hazır: $email -> $user_id"
done

for i in $(seq 1 "$SURVEY_COUNT"); do
  idx=$(( (i - 1) % USER_COUNT ))
  creator_id="${USER_IDS[$idx]}"

  survey_payload=$(cat <<EOF
{
  "creator_id": "$creator_id",
  "questions": [
    {
      "type": "single_choice",
      "text": "Test Anket #$i: Hangi seçenek daha iyi?",
      "order": 1,
      "options": [
        { "text": "Seçenek A" },
        { "text": "Seçenek B" },
        { "text": "Seçenek C" },
        { "text": "Seçenek D" }
      ]
    }
  ]
}
EOF
)

  survey_resp="$(request_json "POST" "$SURVEY_URL" "$survey_payload" || true)"
  survey_id="$(json_field "$survey_resp" "id" || true)"

  if [[ -z "$survey_id" ]]; then
    echo "Anket oluşturulamadı: #$i (creator_id=$creator_id)"
    echo "Response: $survey_resp"
    exit 1
  fi

  echo "Survey oluşturuldu: #$i -> $survey_id (creator_id=$creator_id)"
done

echo "Tamamlandı: $USER_COUNT kullanıcı + $SURVEY_COUNT anket."
