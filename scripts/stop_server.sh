PID=$(sudo lsof -t -i:5000)

if [ -n "$PID" ]; then
  sudo kill -9 $PID
fi
