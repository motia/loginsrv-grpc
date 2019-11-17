package auth

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	md "google.golang.org/grpc/metadata"
)

// LoginSrvServer proxies the REST api though grpc
type LoginSrvServer struct {
	UnimplementedAuthServer
	apiClient *http.Client
	baseURL   *string
}

// Option allows functional configuration for the loginServer
type Option func(*LoginSrvServer)

// NewLoginSrvServer creates the AuthServer
func NewLoginSrvServer(url string, options ...Option) *LoginSrvServer {
	srv := &LoginSrvServer{
		baseURL: &url,
		apiClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	for i := range options {
		options[i](srv)
	}
	return srv
}

// Authenticate validates the access token available in the context meatadata
func Authenticate(ctx context.Context) (context.Context, error) {
	accessToken, err := grpc_auth.AuthFromMD(ctx, "bearer")

	if err != nil {
		return nil, err
	}

	// TODO: validate token from request
	if len(accessToken) == 0 {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}

	return ctx, nil
}

// AuthFuncOverride used internally to skip authentication for login route
func (s *LoginSrvServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	if strings.Contains(fullMethodName, "AttemptLogin") {
		return ctx, nil
	}
	return Authenticate(ctx)
}

// AttemptLogin is a basic authentication
func (s *LoginSrvServer) AttemptLogin(ctx context.Context, request *LoginRequest) (*LoginReply, error) {
	data := "username=" + request.Username + "&password=" + request.Password
	return s.loginWithAPI(&data, nil)
}

func (s *LoginSrvServer) loginWithAPI(loginData *string, cookie *string) (*LoginReply, error) {
	var reader io.Reader = nil
	if loginData != nil {
		reader = strings.NewReader(*loginData)
	}
	req, err := http.NewRequest(
		"POST",
		*s.baseURL+"/login", reader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.Header.Add("Cookie", "jwt_token="+*cookie)
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "Unknown")
	}

	resp, err := s.apiClient.Do(req)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "Unknown")
	}

	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Println("bye", err2)
	}

	if resp.StatusCode == 400 {
		return nil, grpc.Errorf(codes.InvalidArgument, string(body))
	}

	if resp.StatusCode == 403 {
		return nil, grpc.Errorf(codes.PermissionDenied, string(body))
	}

	if resp.StatusCode == 200 {
		return &LoginReply{AccessToken: string(body)}, nil
	}

	return nil, grpc.Errorf(codes.Unknown, string(body))
}

// RefreshToken refreshes the token sent through the context metadata
func (s *LoginSrvServer) RefreshToken(ctx context.Context, request *RefreshRequest) (*LoginReply, error) {
	metadata, _ := md.FromIncomingContext(ctx)
	oldToken := strings.SplitN(
		metadata.Get("authorization")[0],
		" ",
		2,
	)[1]

	return s.loginWithAPI(nil, &oldToken)
}

// GetProfile returns the user profile
func (s *LoginSrvServer) GetProfile(context.Context, *ProfileRequest) (*Profile, error) {
	return &Profile{Name: "Motia"}, nil
}
