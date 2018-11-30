package credentials

import (
	"context"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/netlify/open-api/go/models"
	"github.com/netlify/open-api/go/porcelain"
	"github.com/skratchdot/open-golang/open"
)

const (
	netlifyApiScheme = "https"
	netlifyApiHost   = "api.netlify.com"
	netlifyTicketURL = "https://app.netlify.com/authorize?response_type=ticket&ticket="
)

var apiSchemes = []string{netlifyApiScheme}

func Login(clientID string) (string, error) {
	transport := client.New(netlifyApiHost, "", apiSchemes)
	client := porcelain.New(transport, strfmt.Default)

	ctx := context.Background()

	ticket, err := client.CreateTicket(ctx, clientID)
	if err != nil {
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
