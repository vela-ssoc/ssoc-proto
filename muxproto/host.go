package muxproto

import (
	"net/url"
	"strconv"
)

const (
	AgentDomain   = "agent.ssoc.internal"
	ManagerDomain = "manager.ssoc.internal"
	BrokerDomain  = "broker.ssoc.internal"
)

func ResolveHostname(id int64, domain string) string {
	sid := strconv.FormatInt(id, 10)

	return sid + "." + domain
}

func ToAgentURL(agentID int64, path string, ws ...bool) *url.URL {
	host := ResolveHostname(agentID, AgentDomain)
	return buildURL(host, path, ws)
}

func ToManagerURL(path string, ws ...bool) *url.URL {
	return buildURL(ManagerDomain, path, ws)
}

func ServerToBrokerURL(brokerID int64, path string, ws ...bool) *url.URL {
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
