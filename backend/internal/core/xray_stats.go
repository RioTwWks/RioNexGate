package core

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"rionexgate/internal/core/xraystats"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// queryXrayStatGRPC reads a single counter from Xray StatsService (gRPC).
func queryXrayStatGRPC(apiAddr, name string) (int64, error) {
	apiAddr = strings.TrimSpace(apiAddr)
	if apiAddr == "" {
		return 0, fmt.Errorf("xray api address not configured")
	}
	// Accept host:port; strip accidental http:// prefix from older configs.
	apiAddr = strings.TrimPrefix(apiAddr, "https://")
	apiAddr = strings.TrimPrefix(apiAddr, "http://")
	if _, _, err := net.SplitHostPort(apiAddr); err != nil {
		return 0, fmt.Errorf("invalid xray api address %q: %w", apiAddr, err)
	}

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
