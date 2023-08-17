package apparmor

import (
	"text/template"
)

var qemuProfileTpl = template.Must(template.New("qemuProfile").Parse(`#include <tunables/global>
profile "{{ .name }}" flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/consoles>
  #include <abstractions/nameservice>

  capability dac_override,
  capability dac_read_search,
  capability ipc_lock,
  capability setgid,
  capability setuid,
  capability sys_chroot,
  capability sys_resource,

  # Needed by qemu
  /dev/hugepages/**                         rw,
  /dev/kvm                                  rw,
  /dev/net/tun                              rw,
  /dev/ptmx                                 rw,
  /dev/sev                                  rw,
  /dev/vfio/**                              rw,
  /dev/vhost-net                            rw,
  /dev/vhost-vsock                          rw,
  /etc/ceph/**                              r,
  /run/udev/data/*                          r,
  /sys/bus/                                 r,
  /sys/bus/nd/devices/                      r,
  /sys/bus/usb/devices/                     r,
  /sys/bus/usb/devices/**                   r,
  /sys/class/                               r,
  /sys/devices/**                           r,
  /sys/module/vhost/**                      r,
  /tmp/lxd_sev_*                            r,
  /{,usr/}bin/qemu*                         mrix,
  {{ .ovmfPath }}/OVMF_CODE.fd              kr,
  {{ .ovmfPath }}/OVMF_CODE_*.fd            kr,
  {{ .ovmfPath }}/OVMF_CODE.*.fd            kr,
  /usr/share/qemu/**                        kr,
  /usr/share/seabios/**                     kr,
  owner @{PROC}/@{pid}/cpuset               r,
  owner @{PROC}/@{pid}/task/@{tid}/comm     rw,
  /etc/nsswitch.conf         r,
  /etc/passwd                r,
  /etc/group                 r,
  @{PROC}/version                           r,

  # Used by qemu for live migration NBD server and client
  unix (bind, listen, accept, send, receive, connect) type=stream,

  # Used by qemu when inside a container
{{- if .userns }}
  unix (send, receive) type=stream,
{{- end }}

  # Instance specific paths
  {{ .logPath }}/** rwk,
  {{ .path }}/** rwk,
  {{ .devicesPath }}/** rwk,

  # Needed for lxd fork commands
  {{ .exePath }} mr,
  @{PROC}/@{pid}/cmdline r,
  /{etc,lib,usr/lib}/os-release r,

  # Things that we definitely don't need
  deny @{PROC}/@{pid}/cgroup r,
  deny /sys/module/apparmor/parameters/enabled r,
  deny /sys/kernel/mm/transparent_hugepage/hpage_pmd_size r,

{{if .libraryPath -}}
  # Entries from LD_LIBRARY_PATH
{{range $index, $element := .libraryPath}}
  {{$element}}/** mr,
{{- end }}
{{- end }}

{{- if .raw }}

  ### Configuration: raw.apparmor
{{ .raw }}
{{- end }}
}
`))
