package core

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"rionexgate/internal/core/xraystats"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// queryXrayStatGRPC reads a single counter from Xray StatsService (gRPC).
// Missing counters (user idle) return 0. When api_address uses host.docker.internal
// and the primary dial fails, 172.17.0.1:port is tried as a fallback.
func queryXrayStatGRPC(apiAddr, name string) (int64, error) {
	apiAddr = normalizeAPIAddr(apiAddr)
	if apiAddr == "" {
		return 0, fmt.Errorf("xray api address not configured")
	}
	if _, _, err := net.SplitHostPort(apiAddr); err != nil {
		return 0, fmt.Errorf("invalid xray api address %q: %w", apiAddr, err)
	}

	addrs := []string{apiAddr}
	if host, port, err := net.SplitHostPort(apiAddr); err == nil && host == "host.docker.internal" {
		addrs = append(addrs, "172.17.0.1:"+port)
	}

	var lastErr error
	for _, addr := range addrs {
		val, err := getStatsOnce(addr, name)
		if err == nil {
			return val, nil
		}
		lastErr = err
		// NotFound / unknown stat → idle user, not a connectivity failure.
		if isStatsNotFound(err) {
			return 0, nil
		}
	}
	return 0, lastErr
}

func normalizeAPIAddr(apiAddr string) string {
	apiAddr = strings.TrimSpace(apiAddr)
	apiAddr = strings.TrimPrefix(apiAddr, "https://")
	apiAddr = strings.TrimPrefix(apiAddr, "http://")
	return apiAddr
}

func isStatsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.NotFound {
			return true
		}
		msg := strings.ToLower(st.Message())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist")
}

func getStatsOnce(apiAddr, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(apiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := xraystats.NewStatsServiceClient(conn)
	resp, err := client.GetStats(ctx, &xraystats.GetStatsRequest{
		Name:   name,
		Reset_: false,
	})
	if err != nil {
		return 0, err
	}
	if resp.GetStat() == nil {
		return 0, nil
	}
	return resp.GetStat().GetValue(), nil
}
