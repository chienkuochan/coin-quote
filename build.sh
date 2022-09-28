#!/bin/sh
go build -o ./cmd/getData sourceData/getData.go
go build -o ./cmd/api api/api.go
go build -o ./cmd/ws ws/ws.go
go build -o ./cmd/workers workers/workers.go
go build -o ./cmd/schedule schedules/schedule.go
