package clients

import (
	"context"
	"fmt"

	pb "graphql-gateway/pkg/gateway"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LLMGatewayClient struct {
	conn   *grpc.ClientConn
	client pb.GatewayServiceClient
}

func NewLLMGatewayClient(addr string) (*LLMGatewayClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to llm-gateway: %w", err)
	}
	return &LLMGatewayClient{
		conn:   conn,
		client: pb.NewGatewayServiceClient(conn),
	}, nil
}

func (c *LLMGatewayClient) Ask(ctx context.Context, question string) (string, error) {
	resp, err := c.client.Ask(ctx, &pb.AskRequest{Question: question})
	if err != nil {
		return "", fmt.Errorf("Ask RPC failed: %w", err)
	}
	return resp.Answer, nil
}

func (c *LLMGatewayClient) Close() error {
	return c.conn.Close()
}
