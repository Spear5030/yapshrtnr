package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Spear5030/yapshrtnr/internal/domain"
	"github.com/Spear5030/yapshrtnr/internal/module"
	"github.com/Spear5030/yapshrtnr/internal/pb"
	pckgstorage "github.com/Spear5030/yapshrtnr/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
)

// ShortenerServer - сервер с точки зрения grpc
type ShortenerServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortenerServer
	Storage       storage
	logger        *zap.Logger
	baseURL       string
	secretKey     string
	trustedSubnet net.IPNet
}

// GRPCServer с портом для запуска
type GRPCServer struct {
	Server *grpc.Server
	Port   string
}

type storage interface {
	SetURL(ctx context.Context, user, short, long string) error
	GetURL(ctx context.Context, short string) (string, bool)
	GetURLsByUser(ctx context.Context, user string) (urls map[string]string)
	SetBatchURLs(ctx context.Context, urls []domain.URL) error
	DeleteURLs(ctx context.Context, user string, shorts []string)
	GetUsersCount(ctx context.Context) (int, error)
	GetUrlsCount(ctx context.Context) (int, error)
}

// New конструктор GRPCServer
func New(storage storage, logger *zap.Logger, port string, baseURL string, skey string, ipNet net.IPNet) *GRPCServer {
	shortenerServer := &ShortenerServer{
		Storage:       storage,
		logger:        logger,
		baseURL:       baseURL,
		secretKey:     skey,
		trustedSubnet: ipNet,
	}
	s := GRPCServer{
		Server: grpc.NewServer(grpc.UnaryInterceptor(shortenerServer.AuthInterceptor)),
		Port:   port,
	}
	reflection.Register(s.Server) // for postman
	pb.RegisterShortenerServer(s.Server, shortenerServer)
	return &s
}

// Start слушает определенный порт и запускает в горутине grpc сервер
func (s *GRPCServer) Start() {
	l, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		log.Fatal("error with listen gRPC:", err)
	}
	go func() {
		if err = s.Server.Serve(l); err != nil {
			log.Fatal("error with serve gRPC:", err)
		}
	}()
}

// Ping проверяет соединение с PostgreSQL
func (s *ShortenerServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	var response *emptypb.Empty
	pinger, ok := s.Storage.(pckgstorage.Pinger)

	if ok {
		if pinger.Ping() == nil {

			return response, nil
		}
		return nil, status.Error(codes.Unavailable, "")
	}
	return nil, status.Error(codes.Unimplemented, "Ping not implemented")
	//log.Fatal("Storage haven't pinger")
}

// GetURL возвращает полную ссылку по короткому представлению
func (s *ShortenerServer) GetURL(ctx context.Context, in *pb.Short) (*pb.GetResponse, error) {
	var response pb.GetResponse
	if len(in.GetShort()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Missing short url")
	}
	response.Long, response.Deleted = s.Storage.GetURL(ctx, in.GetShort())
	return &response, nil
}

// PostURL получает URL. Преобразует и отправляет в storage. Возвращает ответ c сокращенным URL
func (s *ShortenerServer) PostURL(ctx context.Context, in *pb.Long) (*pb.Short, error) {
	if len(in.Long) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No url for shorting")
	}
	short, err := module.ShortingURL(in.Long)
	if err != nil {
		s.logger.Info("Error shorting", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user := getUserByMD(ctx)
	err = s.Storage.SetURL(ctx, user, short, in.Long)
	var de *pckgstorage.DuplicationError
	var response pb.Short
	if err != nil {
		if errors.As(err, &de) {
			response.Short = fmt.Sprintf("%s/%s", s.baseURL, de.Duplication)
			return &response, status.Error(codes.AlreadyExists, err.Error())
		}
		s.logger.Info("Error save short", zap.String("err", err.Error()))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	response.Short = short
	return &response, nil
}

// GetURLsByUser возвращает слайс ссылок, которые созданы текущим пользователем
func (s *ShortenerServer) GetURLsByUser(ctx context.Context, in *emptypb.Empty) (*pb.ResponseGetURLsByUser, error) {
	user := getUserByMD(ctx)
	urls := s.Storage.GetURLsByUser(ctx, user)
	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "0 urls")
	}
	response := &pb.ResponseGetURLsByUser{}
	for short, long := range urls {
		response.Urls = append(response.Urls, &pb.URL{
			Short: s.baseURL + "/" + short,
			Long:  long})
	}
	return response, nil
}

// GetInternalStats возвращает статистику по пользователям и ссылкам, если запрос идет из доверенных подсетей
func (s *ShortenerServer) GetInternalStats(ctx context.Context, in *emptypb.Empty) (*pb.StatsResponse, error) {
	//Можно вынести в interceptor, но доверенные сети нужны только в одной функции
	var ip string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		fmt.Println(md)
		values := md.Get("x-real-ip")
		if len(values) > 0 {
			ip = values[0]
		}
	}
	netIP := net.ParseIP(ip)
	if netIP == nil {
		s.logger.Info("no x-real-ip header")
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}
	if !s.trustedSubnet.Contains(netIP) {
		s.logger.Info("GetInternalStats from non trusted subnet", zap.String("IP", netIP.String()))
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}
	response := &pb.StatsResponse{}
	users, err := s.Storage.GetUsersCount(ctx)
	response.Users = int32(users)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	urls, err := s.Storage.GetUrlsCount(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response.Urls = int32(urls)
	return response, nil
}

// PostBatchURLs получает список URL.  Преобразует и отправляет в storage. Возвращает ответ c сокращенными URL и CorrelationID
func (s *ShortenerServer) PostBatchURLs(ctx context.Context, in *pb.RequestBatchURLs) (*pb.ResponseBatchURLs, error) {
	if len(in.Inputs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No urls for shorting")
	}

	user := getUserByMD(ctx)
	urls := make([]domain.URL, 0, len(in.Inputs))
	response := &pb.ResponseBatchURLs{}
	for _, input := range in.Inputs {
		tmpShort, errInput := module.ShortingURL(input.Long)
		if errInput != nil {
			return nil, status.Error(codes.Internal, errInput.Error())
		}
		urls = append(urls, domain.URL{
			Short: tmpShort,
			Long:  input.Long,
			User:  user,
		})
		response.Outputs = append(response.Outputs, &pb.ResponseBatchURLsOutput{
			Short:         tmpShort,
			CorrelationId: input.CorrelationId,
		})
	}

	err := s.Storage.SetBatchURLs(ctx, urls)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return response, nil
}

// DeleteBatchByUser Пакетное удаление ссылок пользователя.
func (s *ShortenerServer) DeleteBatchByUser(ctx context.Context, in *pb.RequestDeleteBatch) (*emptypb.Empty, error) {

	if len(in.Shorts) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No urls for delete")
	}
	shorts := make([]string, len(in.Shorts))
	for _, short := range in.Shorts {
		shorts = append(shorts, short.Short)
	}
	user := getUserByMD(ctx)
	s.Storage.DeleteURLs(ctx, user, shorts)
	return &emptypb.Empty{}, nil
}

// AuthInterceptor проверяет наличие токена и его валидность
func (s *ShortenerServer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	switch info.FullMethod {
	case "/yapshrtnr.Shortener/PingDB":
		return handler(ctx, req)
	case "/yapshrtnr.Shortener/GetInternalStats":
		return handler(ctx, req)
	case "/yapshrtnr.Shortener/GetURL":
		return handler(ctx, req)
	}
	var id, token []byte
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("token")
		if len(values) > 0 {
			token, _ = hex.DecodeString(values[0])
		}
		values = md.Get("id")
		if len(values) > 0 {
			id = []byte(values[0])
		}
	}
	if len(token) == 0 || len(id) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write(id)

	if !hmac.Equal(h.Sum(nil), token) {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return handler(ctx, req)
}

// getUserByMD получает id пользователя из метаданных. ошибки уже отловлены на уровне interceptor'a
func getUserByMD(ctx context.Context) (user string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("id")
		if len(values) > 0 {
			user = values[0]
			return user
		}
	}
	return ""
}
