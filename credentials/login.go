package credentials

import (
	"context"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/netlify/open-api/go/models"
	"github.com/netlify/open-api/go/porcelain"
	"github.com/skratchdot/open-golang/open"

	apiContext "github.com/netlify/open-api/go/porcelain/context"
)

const (
	netlifyApiScheme = "https"
	netlifyApiHost   = "api.netlify.com"
	netlifyTicketURL = "https://app.netlify.com/authorize?response_type=ticket&ticket="
)

var apiSchemes = []string{netlifyApiScheme}

func Login(clientID string) (string, error) {
	transport := client.New(netlifyApiHost, "/api/v1", apiSchemes)
	client := porcelain.New(transport, strfmt.Default)

	creds := runtime.ClientAuthInfoWriterFunc(noCredentials)
	ctx := apiContext.WithAuthInfo(context.Background(), creds)

	ticket, err := client.CreateTicket(ctx, clientID)
	if err != nil {
		return "", err
	}

	if err := openAuthUI(ticket); err != nil {
		return "", err
	}

	if !ticket.Authorized {
		a, err := client.WaitUntilTicketAuthorized(ctx, ticket)
		if err != nil {
			return "", err
		}

		ticket = a
	}

	token, err := client.ExchangeTicket(ctx, ticket.ID)
	if err != nil {
		return "", err
	}

	if err := SaveAccessToken(token.AccessToken); err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func openAuthUI(ticket *models.Ticket) error {
	return open.Run(netlifyTicketURL + ticket.ID)
}

func noCredentials(r runtime.ClientRequest, _ strfmt.Registry) error {
	r.SetHeaderParam("User-Agent", "git-credential-netlify")
	return nil
}
