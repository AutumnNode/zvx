package service

import "time"


type PortIpRule struct {
    ID       string `json:"id"`
    IP       string `json:"ip"`
    Port     int    `json:"port"`
    Protocol string `json:"protocol"`
    Type     string `json:"type"`
}

var mockPortIpRules = []PortIpRule{
    {ID: "pi-1", IP: "192.168.1.100", Port: 8080, Protocol: "TCP", Type: "allow"},
    {ID: "pi-2", IP: "10.0.0.5", Port: 443, Protocol: "TCP", Type: "allow"},
}

func GetMockPortIpRules() []PortIpRule {
    return mockPortIpRules
}

func CreateMockPortIpRule(rule PortIpRule) PortIpRule {
    rule.ID = "pi-" + time.Now().Format("20060102150405")
    mockPortIpRules = append(mockPortIpRules, rule)
    return rule
}

func UpdateMockPortIpRule(id string, updatedRule PortIpRule) PortIpRule {
    for i, rule := range mockPortIpRules {
        if rule.ID == id {
            updatedRule.ID = id
            mockPortIpRules[i] = updatedRule
            return updatedRule
        }
    }
    return PortIpRule{}
}

func DeleteMockPortIpRule(id string) {
    for i, rule := range mockPortIpRules {
        if rule.ID == id {
            mockPortIpRules = append(mockPortIpRules[:i], mockPortIpRules[i+1:]...)
            return
        }
    }
}


type NetworkStats struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    Traffic struct {
        Inbound     string  `json:"inbound"`
        Outbound    string  `json:"outbound"`
        InboundRaw  float64 `json:"inboundRaw"`
        OutboundRaw float64 `json:"outboundRaw"`
    } `json:"traffic"`
    Connections struct {
        Active int `json:"active"`
        Trend  struct {
            Direction  string  `json:"direction"`
            Percentage float64 `json:"percentage"`
        } `json:"trend"`
    } `json:"connections"`
}

func GetMockNetworkStats(ns string) NetworkStats {
    return NetworkStats{
        Status:  "online",
        Message: "一切正常",
        Traffic: struct {
            Inbound     string  `json:"inbound"`
            Outbound    string  `json:"outbound"`
            InboundRaw  float64 `json:"inboundRaw"`
            OutboundRaw float64 `json:"outboundRaw"`
        }{"1.2 MB/s", "500 KB/s", 1.2, 0.5},
        Connections: struct {
            Active int `json:"active"`
            Trend  struct {
                Direction  string  `json:"direction"`
                Percentage float64 `json:"percentage"`
            } `json:"trend"`
        }{32, struct {
            Direction  string  `json:"direction"`
            Percentage float64 `json:"percentage"`
        }{"up", 5.5}},
    }
}

type TopologyData struct {
    Nodes []struct {
        ID   string  `json:"id"`
        Name string  `json:"name"`
        Type string  `json:"type"`
        X    float64 `json:"x"`
        Y    float64 `json:"y"`
    } `json:"nodes"`
    Connections []struct {
        Source string `json:"source"`
        Target string `json:"target"`
        Status string `json:"status"`
    } `json:"connections"`
}

func GetMockTopologyData(ns, view string) TopologyData {
    return TopologyData{
        Nodes: []struct {
            ID   string  `json:"id"`
            Name string  `json:"name"`
            Type string  `json:"type"`
            X    float64 `json:"x"`
            Y    float64 `json:"y"`
        }{
            {"pod-1", "web", "pod", 10.0, 20.0},
            {"svc-1", "web-svc", "service", 50.0, 50.0},
        },
        Connections: []struct {
            Source string `json:"source"`
            Target string `json:"target"`
            Status string `json:"status"`
        }{
            {"pod-1", "svc-1", "active"},
        },
    }
}
