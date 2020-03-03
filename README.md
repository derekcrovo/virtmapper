# Virtmapper
## About

Virtmapper is a utility for mapping libvirt guests to hosts.  It provides answers to the questions "What host is x VM on?" and "What are the VMs on x host?" in an easily automatable way.

Virtmapper was made in an enviroment where we didn't have an operations team or proper VM orchestration software.  If you are fortunate enough to have an environment where VMs are cattle rather than pets you probably don't have a need for it.  That being said it was valuable for us so maybe others will find it useful as well.

This is a program I tend to rewrite each time I'm learning a new programming language.  This is the latest incarnation in Go, and is so far the best one.

## Usage
Virtmapper works with Ansible.  Ansible periodically runs the "virsh list" command on the libvirt hosts and writes the output to a tempfile.  This is typically done as an Ansible ad-hoc command run from cron.  Virtmapper parses this file regularly and builds up a map of hosts to guests.

The virtmapper binary is used both as a server and as a client.  

Usage:
virtmapper -http <listen_address> | -server <server_address> [-query <host_to_look_up>] | -version
