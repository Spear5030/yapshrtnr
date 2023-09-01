package server

import (
	"context"
	"github.com/Spear5030/yapshrtnr/internal/config"
	"github.com/Spear5030/yapshrtnr/internal/pb"
	testStorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"github.com/Spear5030/yapshrtnr/pkg/logger"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"testing"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	cfg, _ := config.New()
	lg, _ := logger.New(true)
	_, IPNet, _ := net.ParseCIDR("127.0.0.0/8")
	srv := New(testStorage.NewMemoryStorage(), lg, cfg.GRPCPort, cfg.BaseURL, cfg.Key, *IPNet)

	go func() {
		if err := srv.Server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestShortenerServer_GetInternalStats(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufconn", grpc.WithContextDialer(dialer()), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufconn: %v", err)
	}
	defer conn.Close()
	client := pb.NewShortenerClient(conn)
	resp, err := client.GetInternalStats(ctx, &emptypb.Empty{})
	require.Nil(t, resp)
	if grpcErr, ok := status.FromError(err); ok {
		require.Equal(t, codes.PermissionDenied, grpcErr.Code())
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", "127.0.0.1")
	resp, err = client.GetInternalStats(ctx, &emptypb.Empty{})
	require.NotNil(t, resp)
	if grpcErr, ok := status.FromError(err); ok {
		require.Equal(t, codes.OK, grpcErr.Code())
	}

	ctx = context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", "192.168.0.1")
	resp, err = client.GetInternalStats(ctx, &emptypb.Empty{})
	require.Nil(t, resp)
	if grpcErr, ok := status.FromError(err); ok {
		require.Equal(t, codes.PermissionDenied, grpcErr.Code())
	}
}

func TestShortenerServer_AuthInterceptor(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufconn", grpc.WithContextDialer(dialer()), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufconn: %v", err)
	}
	defer conn.Close()
	client := pb.NewShortenerClient(conn)
	resp, err := client.PostURL(ctx, &pb.Long{Long: "https://google.com"})
	require.Nil(t, resp)
	if grpcErr, ok := status.FromError(err); ok {
		require.Equal(t, codes.Unauthenticated, grpcErr.Code())
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "id", "12345")
	ctx = metadata.AppendToOutgoingContext(ctx, "token", "f5d1cf1a06e1c9e562ea9203c56bf9556012b4cc56d26d19f2d9537e2af64c6d")
	resp, err = client.PostURL(ctx, &pb.Long{Long: "https://google.com"})
	require.NotNil(t, resp)
	if grpcErr, ok := status.FromError(err); ok {
		require.Equal(t, codes.OK, grpcErr.Code())
	}

}
