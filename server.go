package loginsrv_grpc

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
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

// AuthFuncOverride used internally to skip authentication for login route
func (s *LoginSrvServer) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {
	return ctx, nil
}

// Authenticate asserts a token is attached to the RPC context
// clients can attach it with NewClientTokenInterceptor
func (s *LoginSrvServer) Authenticate(ctx context.Context) (context.Context, error) {
	accessToken, err := grpc_auth.AuthFromMD(ctx, "bearer")

	if err != nil {
		return nil, err
	}

	// validate token on microservice
	if len(accessToken) == 0 {
		oldToken := getTokenFromContext(ctx)
		if oldToken == nil {
			return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
		}
		_, err = s.GetProfile(ctx, &ProfileRequest{})

		return ctx, err
	}

	return ctx, nil
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

// AttemptLogin is a basic authentication
func (s *LoginSrvServer) AttemptLogin(ctx context.Context, request *LoginRequest) (*LoginReply, error) {
	data := "username=" + request.Username + "&password=" + request.Password
	return s.postLogin(&data, nil)
}

func (s *LoginSrvServer) postLogin(data *string, token *string) (*LoginReply, error) {
	body, err := s.loginWithAPI("POST", "jwt", data, token)
	if err == nil {
		return &LoginReply{AccessToken: *body}, err
	}
	return nil, err
}

func (s *LoginSrvServer) loginWithAPI(method string, contentType string, loginData *string, cookie *string) (*string, error) {
	var reader io.Reader = nil
	if loginData != nil {
		reader = strings.NewReader(*loginData)
	}
	req, err := http.NewRequest(method, *s.baseURL+"/login", reader)
	req.Header.Add("Accept", "application/"+contentType)
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
		return nil, grpc.Errorf(codes.Internal, "Internal")
	}

	if resp.StatusCode == 400 {
		return nil, grpc.Errorf(codes.InvalidArgument, string(body))
	}

	if resp.StatusCode == 403 {
		return nil, grpc.Errorf(codes.PermissionDenied, string(body))
	}

	if resp.StatusCode == 200 {
		temp := string(body)
		return &temp, nil
	}

	return nil, grpc.Errorf(codes.Unknown, string(body))
}

// RefreshToken refreshes the token sent through the context metadata
func (s *LoginSrvServer) RefreshToken(ctx context.Context, request *RefreshRequest) (*LoginReply, error) {
	oldToken := getTokenFromContext(ctx)
	if oldToken == nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
	return s.postLogin(nil, oldToken)
}

func getTokenFromContext(ctx context.Context) *string {
	metadata, _ := md.FromIncomingContext(ctx)
	authHeader := metadata.Get(AuthTokenMetadataKey)
	if len(authHeader) == 0 {
		return nil
	}

	segs := strings.SplitN(
		authHeader[0],
		" ",
		2,
	)
	if len(segs) < 2 {
		return nil
	}
	return &segs[1]
}

// GetProfile returns the user profile
func (s *LoginSrvServer) GetProfile(ctx context.Context, profileRequest *ProfileRequest) (*Profile, error) {
	oldToken := getTokenFromContext(ctx)
	if oldToken == nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	}
	jsonStr, err := s.loginWithAPI("GET", "json", nil, oldToken)
	if err != nil {
		return nil, err
	}

	user := &userInfo{}
	err = json.Unmarshal([]byte(*jsonStr), user)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Internal")
	}
	return &Profile{
		Sub:       user.Sub,
		Picture:   user.Picture,
		Name:      user.Name,
		Email:     user.Email,
		Origin:    user.Origin,
		Expiry:    user.Expiry,
		Refreshes: int32(user.Refreshes),
		Domain:    user.Domain,
		Groups:    user.Groups,
	}, nil
}

// Loginsrv returned profile type
type userInfo struct {
	Sub       string   `json:"sub"`
	Picture   string   `json:"picture,omitempty"`
	Name      string   `json:"name,omitempty"`
	Email     string   `json:"email,omitempty"`
	Origin    string   `json:"origin,omitempty"`
	Expiry    int64    `json:"exp,omitempty"`
	Refreshes int      `json:"refs,omitempty"`
	Domain    string   `json:"domain,omitempty"`
	Groups    []string `json:"groups,omitempty"`
}
