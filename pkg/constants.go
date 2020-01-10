package pkg

const (
	chronyConfigFile = "/etc/chrony.conf"
	chronyServiceName = "chronyd"
	chronyLeader = `{{- range .NTPServers -}}
server {{ . }} iburst
{{end -}}
allow all
stratumweight 0
driftfile /var/lib/chrony/drift
rtcsync
makestep 10 3
bindcmdaddress 127.0.0.1
bindcmdaddress ::1
keyfile /etc/chrony.keys
commandkey 1
generatecommandkey
noclientlog
logchange 0.5
logdir /var/log/chrony
local stratum 10
`

	chronyFollower = `server {{ .LeaderIP }} iburst
allow all
stratumweight 0
driftfile /var/lib/chrony/drift
rtcsync
makestep 10 3
bindcmdaddress 127.0.0.1
bindcmdaddress ::1
keyfile /etc/chrony.keys
commandkey 1
generatecommandkey
noclientlog
logchange 0.5
logdir /var/log/chrony
local stratum 10
`
)


