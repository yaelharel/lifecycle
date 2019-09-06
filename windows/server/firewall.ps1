# Also requires an IaaS port firewall for the same port
New-NetFirewallRule -LocalPort 2376 -DisplayName DockerTLS -Protocol TCP

