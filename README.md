# Virtmapper
## About

Virtmapper is a utility for mapping libvirt guests to hosts.  It provides answers to the questions "What host is x VM on?" and "What are the VMs on x host?" in an easily automatable way.

This is a toy program I tend to rewrite each time I'm learning a new programming language.  The current incarnation is in Go, and is so far the most robust one.

Virtmapper was made in an enviroment where there was no operations team or proper virtualization orchestration software.  If you are fortunate enough to have an environment where VMs are cattle rather than pets then you probably don't have a need for it.  That being said it was a valuable tool for helping to manage a moderately-sized production virtualization cluster so maybe others will find it useful as well.


## How it works
Virtmapper works with Ansible.  Ansible periodically runs the "virsh list" command on the libvirt hosts and writes the output to a tempfile.  Virtmapper parses this file regularly and builds up a map of hosts to guests.

## Usage
The virtmapper binary is used both as a server and as a client.  

Server Usage
```bash
virtmapper serve [options]
OPTIONS:
   --address value, -a value            address and port to listen on (default: ":7474")
   --logfile value, -l value            log file for server activity (default: "/var/log/virtmapper")
   --refreshInterval value, -r value    map refresh interval in minutes (default: 60)
   --ansibleOutputFile value, -v value  path to Ansible output file to read (default: "/tmp/virtmapper.txt")
```

Client Usage
```bash
virtmapper serve query <hostname> [options]
OPTIONS:
   --server value, -s value  address of server to query
```

### Examples
```bash
# Launch the server in the background
$ virtmapper serve &

# Make queries
$ virtmapper query compute-64
compute-64 is a virtual guest on host: kvm43
$ virtmapper query kvm09
kvm09 is a virtual host for guests: olh, tam
```

## Ansible
Ansible is needed to provide the input that Virtmapper consumes.  It can be a simple as running an ad-hoc command from cron:

```bash
*/15 * * * * /usr/bin/ansible vhosts -a '/usr/bin/virsh list --all' &> /tmp/virtmapper.txt
```

## API
The REST API is used by the CLI client but may be consumed by other tools.  It exposes one endpoint, `api/v1/vmap`, for the querying of hosts.  A query is an arbitrary hostname, it may correspond to a virtual host or a virtual guest in virtmapper's main map.  The response is a JSON encoded Vmap structure.  Errors (such as the given hostname not existing in the map) are returned as a JSON object with a single key "error" and a value containing the error string.
A successful query for a hostname will return a Vmap with either a single host or a single guest object.  A query on the vmap endpoint with no hostname will return virtmapper's entire vmap containing many hosts and guests.

### Examples

#### Query returning a virtual host

Request:  `http://localhost:7474/api/v1/vmap/kvm09`

Response:
```json
{
	"hosts": {
		"kvm09": {
			"state": "up",
			"guests": [
				"olh",
				"tam"
			]
		}
	},
	"guests": null
}
```

#### Query returning a virtual guest

Request:  `http://localhost:7474/api/v1/vmap/tam`

Response:
```json
{
	"hosts": null,
	"guests": {
		"tam": {
			"state": "running",
			"host": "kvm09"
		}
	}
}
```