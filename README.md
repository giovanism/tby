# tby

Teleport tunnels manager. Define your teleport tunnels using simple YAML file
and connect to your remote databases and services fast with this script.

```
$ tby up 0
$ tby up 2
$ tby ls
Id  Name                       Port       Status
0   root@staging-database      5432:5432  up
1   root@production-database   5433:5432
2   svc/staging-api-server     8080:80    up
3   svc/staging-frontend       3000:80
4   svc/production-api-server  8081:80
5   svc/production-frontend    3001:80
```

Example `tby.yml` config file:

```yml
tunnels:
- type: ssh
  user: root
  node_name: staging-database
  remote_port: 5432
  local_port: 5432
- type: ssh
  user: root
  node_name: production-database
  remote_port: 5432
  local_port: 5433
- type: k8s
  context: staging-cluster
  resource_namespace: default
  resource_kind: svc
  resource_name: staging-api-server
  remote_port: 80
  local_port: 8080
- type: k8s
  context: staging-cluster
  resource_namespace: default
  resource_kind: svc
  resource_name: staging-frontend
  remote_port: 80
  local_port: 3000
- type: k8s
  context: production-cluster
  resource_namespace: default
  resource_kind: svc
  resource_name: production-api-server
  remote_port: 80
  local_port: 8081
- type: k8s
  context: production-cluster
  resource_namespace: default
  resource_kind: svc
  resource_name: production-frontend
  remote_port: 80
  local_port: 3001
```

Look for where to put `tby.yml` config file this way.

```
$ tby -v ls
8:59AM FTL Can't load tby config file: open /Users/user.name/Library/Application Support/tby/tby.yml: no such file or directory
```

## Install

```sh
go install github.com/giovanism/tby@latest
```
