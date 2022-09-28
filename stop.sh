cat pids/api.pid  | xargs kill -INT
cat pids/ws.pid  | xargs kill -9
cat pids/getData.pid  | xargs kill -INT
cat pids/schedule.pid  | xargs kill -INT
cat pids/workers.pid  | xargs kill -INT
