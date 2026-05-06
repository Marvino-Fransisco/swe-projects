#!/bin/sh

/usr/local/bin/workers &
W_PID=$!
/usr/local/bin/web-backend &
WB_PID=$!
/usr/local/bin/mobile-backend &
MB_PID=$!

trap "kill $W_PID $WB_PID $MB_PID 2>/dev/null; exit 0" TERM INT QUIT

while sleep 1; do
    if ! kill -0 $W_PID 2>/dev/null || ! kill -0 $WB_PID 2>/dev/null || ! kill -0 $MB_PID 2>/dev/null; then
        break
    fi
done

kill $W_PID $WB_PID $MB_PID 2>/dev/null
wait 2>/dev/null
exit 1
