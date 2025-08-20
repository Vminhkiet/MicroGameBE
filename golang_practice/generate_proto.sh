#!/bin/bash
set -e

echo "🔧 Tạo thư mục đầu ra..."
mkdir -p pkg/packets/ping
mkdir -p pkg/packets/action
mkdir -p pkg/packets/shared
mkdir -p pkg/packets/game
mkdir -p pkg/packets/match

protoc -I proto proto/ping/ping.proto \
  --go_out=pkg/packets \
  --go_opt=paths=source_relative

protoc -I proto proto/action/action.proto \
  --go_out=pkg/packets \
  --go_opt=paths=source_relative

protoc -I proto proto/shared/envelope.proto \
  --go_out=pkg/packets \
  --go_opt=paths=source_relative

protoc -I proto proto/game/game.proto \
  --go_out=pkg/packets \
  --go_opt=paths=source_relative

protoc -I proto proto/match/match.proto \
  --go_out=pkg/packets \
  --go_opt=paths=source_relative

echo "✅ Done generating .pb.go!"
