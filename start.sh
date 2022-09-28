nohup ./cmd/api >> logs/api.log &
nohup ./cmd/ws >> logs/ws.log &
nohup ./cmd/schedule >> logs/schedule.log &
nohup ./cmd/workers >> logs/workers.log &
nohup ./cmd/getData >> logs/getData.log &
