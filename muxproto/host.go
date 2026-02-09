package muxproto

import (
	"net/url"
)

const (
	AgentDomain   = "agent.ssoc.internal"
	ManagerDomain = "manager.ssoc.internal"
	BrokerDomain  = "broker.ssoc.internal"
)

func ResolveHostname(id, domain string) string {
	return id + "." + domain
}

func ToAgentURL(agentID, path string, ws ...bool) *url.URL {
	host := ResolveHostname(agentID, AgentDomain)
	return buildURL(host, path, ws)
}

func ToManagerURL(path string, ws ...bool) *url.URL {
	return buildURL(ManagerDomain, path, ws)
}

func ManagerToBrokerURL(brokerID, path string, ws ...bool) *url.URL {
	host := ResolveHostname(brokerID, BrokerDomain)
	return buildURL(host, path, ws)
}

func AgentToBrokerURL(path string, ws ...bool) *url.URL {
	return buildURL(BrokerDomain, path, ws)
}

func buildURL(host, path string, ws []bool) *url.URL {
	scheme := "http"
	if len(ws) != 0 && ws[0] {
		scheme = "ws"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
}
