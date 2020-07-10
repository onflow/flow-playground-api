package sessions

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"

	"github.com/dapperlabs/flow-playground-api/middleware/httpcontext"
)

type Manager struct {
	sc            *securecookie.SecureCookie
	cookieName    string
	cookieOptions CookieOptions
}

type CookieOptions struct {
	MaxAge   int
	Secure   bool
	SameSite http.SameSite
	HTTPOnly bool
}

func NewManager(
	cookieName string,
	cookieHashKey []byte,
	cookieOptions CookieOptions,
) *Manager {
	sc := securecookie.New(cookieHashKey, nil)

	return &Manager{
		sc:            sc,
		cookieName:    cookieName,
		cookieOptions: cookieOptions,
	}
}

func (m *Manager) CurrentSessionID(ctx context.Context) (uuid.UUID, error) {
	value, err := m.getSessionCookie(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	if value == "" {
		return uuid.Nil, nil
	}

	var sessionID uuid.UUID

	err = sessionID.UnmarshalText([]byte(value))
	if err != nil {
		return uuid.Nil, err
	}

	return sessionID, nil
}

func (m *Manager) SaveSession(ctx context.Context, sessionID uuid.UUID) error {
	value := sessionID.String()
	return m.setSessionCookie(ctx, value)
}

func (m *Manager) getSessionCookie(ctx context.Context) (string, error) {
	r := httpcontext.Request(ctx)
	c, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}

		return "", err
	}

	var value string
	err = m.sc.Decode(m.cookieName, c.Value, &value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (m *Manager) setSessionCookie(ctx context.Context, value string) error {
	encodedValue, err := m.sc.Encode(m.cookieName, value)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     m.cookieName,
		MaxAge:   m.cookieOptions.MaxAge,
		Secure:   m.cookieOptions.Secure,
		SameSite: m.cookieOptions.SameSite,
		HttpOnly: m.cookieOptions.HTTPOnly,
		Value:    encodedValue,
	}

	http.SetCookie(httpcontext.Writer(ctx), cookie)

	return nil
}
