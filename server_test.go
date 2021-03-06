package loginsrv_grpc

import (
	"testing"
	"time"

	md "google.golang.org/grpc/metadata"
)

func TestAuthenticateWithoutTokenFails(t *testing.T) {
	srv := NewLoginSrvServer("http://localhost:8080")
	token := ""
	ctx := &contextWithAuthorizationStub{authToken: token}

	_, err := srv.Authenticate(ctx)
	if err == nil {
		t.Error("Authenticate should fail", err)
	}
}

func TestAuthenticateWithTokenSucceeds(t *testing.T) {
	srv := NewLoginSrvServer("http://localhost:8080")
	token := obtainTokenOrFail(t, srv)
	ctx := &contextWithAuthorizationStub{authToken: token}

	x, err := srv.Authenticate(ctx)
	if err != nil {
		t.Error("Token could not be obtained", err)
	}

	if x == nil {
		t.Error("Context should be returned")
	}
}

func TestGetProfile(t *testing.T) {
	srv := NewLoginSrvServer("http://localhost:8080")
	token := obtainTokenOrFail(t, srv)
	ctx := &contextWithAuthorizationStub{authToken: token}

	profile, err := srv.GetProfile(ctx, &ProfileRequest{})
	if err != nil {
		t.Error("Error while loading profile", err)
	}

	if profile == nil {
		t.Error("Profile should not be nil", err)
	}

	if profile.Sub != "bob" {
		t.Error("Expected 'bob' profile but got " + profile.Sub)
	}
}

func TestAttemptLoginThenRefresh(t *testing.T) {
	srv := NewLoginSrvServer("http://localhost:8080")

	loginReply, err := srv.AttemptLogin(nil, &LoginRequest{
		Username: "bob",
		Password: "secret",
	})

	if err != nil {
		t.Error("Login failed", err)

		return
	}

	assertHasAccessToken(t, loginReply)

	// md, ok = ctx.Value(mdIncomingKey{}).(MD)
	refreshCtx := &contextWithAuthorizationStub{
		authToken: loginReply.AccessToken,
	}

	refreshReply, _ := srv.RefreshToken(refreshCtx, &RefreshRequest{})
	if err != nil {
		t.Error("Refresh failed ", err)

		return
	}

	assertHasAccessToken(t, refreshReply)
}

func obtainTokenOrFail(t *testing.T, srv *LoginSrvServer) string {
	loginReply, err := srv.AttemptLogin(nil, &LoginRequest{
		Username: "bob",
		Password: "secret",
	})

	if err != nil {
		t.Error("Token could not be obtained", err)
	}
	return loginReply.AccessToken
}

func assertHasAccessToken(t *testing.T, r *LoginReply) {
	if r.GetAccessToken() == "" {
		t.Error("No access token in reply")
	}
}

type contextWithAuthorizationStub struct {
	authToken string
}

func (*contextWithAuthorizationStub) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, true
}

func (*contextWithAuthorizationStub) Done() <-chan struct{} {
	return make(<-chan struct{})
}

func (*contextWithAuthorizationStub) Err() error {
	return nil
}

func (m *contextWithAuthorizationStub) Value(key interface{}) interface{} {
	if m.authToken == "" {
		return md.New(map[string]string{})
	}
	return md.New(map[string]string{
		AuthTokenMetadataKey: "bearer " + m.authToken,
	})
}
