if [  $(docker inspect -f '{{.State.Running}}' postgres) == "true" ]
then
    echo "Container đang chạy, đang dừng..."
    echo "Container đã được dừng."
fi