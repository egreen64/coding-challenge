package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/egreen64/codingchallenge/auth"
	"github.com/egreen64/codingchallenge/graph/generated"
	"github.com/egreen64/codingchallenge/graph/model"
	"github.com/egreen64/codingchallenge/utils"
	"github.com/google/uuid"
)

func (r *mutationResolver) Authenticate(ctx context.Context, username string, password string) (string, error) {
	if username != r.Config.Auth.Username || password != r.Config.Auth.Password {
		return "", errors.New("invalid credentials")
	}

	tokenString, err := auth.CreateJWT(username, password)
	return tokenString, err
}

func (r *mutationResolver) Enqueue(ctx context.Context, ip []string) (*bool, error) {
	authToken := auth.GetContextToken(ctx)
	if authToken == "" {
		return nil, errors.New("not authorized")
	}

	tokenString := strings.TrimPrefix(authToken, "Bearer ")
	valid, err := auth.ValidateJWT(tokenString, r.Config.Auth.Username, r.Config.Auth.Password)

	if !valid || err != nil {
		return nil, errors.New("not authorized")
	}

	for _, ipAddr := range ip {
		if !utils.IsValidIPV4Address(ipAddr) {
			errorString := fmt.Sprintf("Invalid IPV4 address: %s", ip)
			return nil, errors.New(errorString)
		}
		resp := r.DNSBL.Lookup(ipAddr)

		respCode := "NXDOMAIN"
		if resp.Responses[0].Resp != "" {
			respCode = resp.Responses[0].Resp
		}

		DNSBlockListRecord := model.DNSBlockListRecord{
			UUID:         uuid.New().String(),
			IPAddress:    ipAddr,
			ResponseCode: respCode,
		}

		r.Database.UpsertRecord(&DNSBlockListRecord)
	}

	result := true

	return &result, nil
}

func (r *queryResolver) GetIPDetails(ctx context.Context, ip *string) (*model.DNSBlockListRecord, error) {
	authToken := auth.GetContextToken(ctx)
	if authToken == "" {
		return nil, errors.New("not authorized")
	}

	tokenString := strings.TrimPrefix(authToken, "Bearer ")
	valid, err := auth.ValidateJWT(tokenString, r.Config.Auth.Username, r.Config.Auth.Password)

	if !valid || err != nil {
		return nil, errors.New("not authorized")
	}

	if !utils.IsValidIPV4Address(*ip) {
		errorString := fmt.Sprintf("Invalid IPV4 address: %s", *ip)
		return nil, errors.New(errorString)
	}

	dblRec, _ := r.Database.SelectRecord(*ip)

	return dblRec, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
