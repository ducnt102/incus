test_init_preseed() {
  # - incusd init --preseed
  incus_backend=$(storage_backend "$INCUS_DIR")
  INCUS_INIT_DIR=$(mktemp -d -p "${TEST_DIR}" XXX)
  chmod +x "${INCUS_INIT_DIR}"
  spawn_incus "${INCUS_INIT_DIR}" false

  (
    set -e
    # shellcheck disable=SC2034
    INCUS_DIR=${INCUS_INIT_DIR}

    storage_pool="incustest-$(basename "${INCUS_DIR}")-data"
    # In case we're running against the ZFS backend, let's test
    # creating a zfs storage pool, otherwise just use dir.
    if [ "$incus_backend" = "zfs" ]; then
        configure_loop_device loop_file_4 loop_device_4
        # shellcheck disable=SC2154
        zpool create -f -m none -O compression=on "incustest-$(basename "${INCUS_DIR}")-preseed-pool" "${loop_device_4}"
        driver="zfs"
        source="incustest-$(basename "${INCUS_DIR}")-preseed-pool"
    elif [ "$incus_backend" = "ceph" ]; then
        driver="ceph"
        source=""
    else
        driver="dir"
        source=""
    fi

    cat <<EOF | incusd init --preseed
config:
  core.https_address: 127.0.0.1:9999
  images.auto_update_interval: 15
storage_pools:
- name: ${storage_pool}
  driver: $driver
  config:
    source: $source
networks:
- name: inct$$
  type: bridge
  config:
    ipv4.address: none
    ipv6.address: none
profiles:
- name: default
  devices:
    root:
      path: /
      pool: ${storage_pool}
      type: disk
- name: test-profile
  description: "Test profile"
  config:
    limits.memory: 2GiB
  devices:
    test0:
      name: test0
      nictype: bridged
      parent: inct$$
      type: nic
EOF

    inc info | grep -q 'core.https_address: 127.0.0.1:9999'
    inc info | grep -q 'images.auto_update_interval: "15"'
    inc network list | grep -q "inct$$"
    inc storage list | grep -q "${storage_pool}"
    inc storage show "${storage_pool}" | grep -q "$source"
    inc profile list | grep -q "test-profile"
    inc profile show default | grep -q "pool: ${storage_pool}"
    inc profile show test-profile | grep -q "limits.memory: 2GiB"
    inc profile show test-profile | grep -q "nictype: bridged"
    inc profile show test-profile | grep -q "parent: inct$$"
    printf 'config: {}\ndevices: {}' | inc profile edit default
    inc profile delete test-profile
    inc network delete inct$$
    inc storage delete "${storage_pool}"

    if [ "$incus_backend" = "zfs" ]; then
        # shellcheck disable=SC2154
        deconfigure_loop_device "${loop_file_4}" "${loop_device_4}"
    fi
  )
  kill_incus "${INCUS_INIT_DIR}"
}
