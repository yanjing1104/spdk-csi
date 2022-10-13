function e2e-spdk-start() {
    e2e-spdk-running && {
        echo "spdk already running. Restart with:"
        echo "e2e-spdk-stop; e2e-spdk-start"
        return 0
    }
    echo "======== start spdk target ========"
    # allocate 1024*2M hugepage
    sudo sh -c 'echo 1024 > /proc/sys/vm/nr_hugepages'
    # start spdk target
    sudo docker run -id --name "${SPDK_CONTAINER}" --privileged --net host -v /dev/hugepages:/dev/hugepages -v /dev/shm:/dev/shm ${SPDKIMAGE} /root/spdk/build/bin/spdk_tgt
    sleep 5s
    # wait for spdk target ready
    sudo docker exec -i "${SPDK_CONTAINER}" timeout 5s /root/spdk/scripts/rpc.py framework_wait_init
    # create 1G malloc bdev
    sudo docker exec -i "${SPDK_CONTAINER}" /root/spdk/scripts/rpc.py bdev_malloc_create -b Malloc0 1024 4096
    # create lvstore
    sudo docker exec -i "${SPDK_CONTAINER}" /root/spdk/scripts/rpc.py bdev_lvol_create_lvstore Malloc0 lvs0
    # start jsonrpc http proxy
    sudo docker exec -id "${SPDK_CONTAINER}" /root/spdk/scripts/rpc_http_proxy.py ${JSONRPC_IP} ${JSONRPC_PORT} ${JSONRPC_USER} ${JSONRPC_PASS}
    echo "======== start sma server at ${SMA_ADDRESS}:${SMA_PORT} ========"
    # prepare the target
    sudo docker exec -i "${SPDK_CONTAINER}" /root/spdk/scripts/rpc.py bdev_null_create null0 100 4096
    # create TCP transports
    sudo docker exec -i "${SPDK_CONTAINER}" /root/spdk/scripts/rpc.py nvmf_get_transports --trtype tcp
    # start sma server
    sudo docker exec -id "${SPDK_CONTAINER}" "${SMA_SERVER}" --config /root/sma.yaml
    sleep 1
    while ! nc -z "${SMA_ADDRESS}" "${SMA_PORT}"; do
        sleep 1
        echo -n "."
    done
    echo ok
}

