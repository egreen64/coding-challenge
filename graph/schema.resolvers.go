package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/egreen64/codingchallenge/auth"
	"github.com/egreen64/codingchallenge/graph/generated"
	"github.com/egreen64/codingchallenge/graph/model"
	"github.com/egreen64/codingchallenge/utils"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *mutationResolver) Authenticate(ctx context.Context, username string, password string) (string, error) {
	if username != r.Config.Auth.Username || password != r.Config.Auth.Password {
		return "", gqlerror.Errorf("invalid credentials")
	}

	tokenString, err := auth.CreateJWT(username, password)
	tokenString = "Bearer " + tokenString

	return tokenString, err
}

func (r *mutationResolver) Enqueue(ctx context.Context, ip []string) (*bool, error) {
	authToken := auth.GetContextToken(ctx)
	if authToken == "" {
		return nil, gqlerror.Errorf("missing auth token")
	}

	tokenString := strings.TrimPrefix(authToken, "Bearer ")
	valid, err := auth.ValidateJWT(tokenString, r.Config.Auth.Username, r.Config.Auth.Password)

	if !valid || err != nil {
		return nil, gqlerror.Errorf("not authorized")
	}

	//Validate ip addresses
	invalidIPAddresses := false
	for _, ipAddr := range ip {
		if !utils.IsValidIPV4Address(ipAddr) {
			invalidIPAddresses = true
			graphql.AddError(ctx, gqlerror.Errorf("invalid IPV4 address: %s", ipAddr))
		}
	}
	if invalidIPAddresses {
		return nil, gqlerror.Errorf("validation error(s)")
	}

	jobsNotAdded := false
	for _, ipAddr := range ip {
		if !r.JobQueue.AddJob(ipAddr) {
			jobsNotAdded = true
			graphql.AddError(ctx, gqlerror.Errorf("unable to queue job for ip address: %s. queue is curently full. please try again", ipAddr))
		}
	}
	if jobsNotAdded {
		return nil, gqlerror.Errorf("scheduling error(s)")
	}

	result := true

	return &result, nil
}

func (r *queryResolver) GetIPDetails(ctx context.Context, ip *string) (*model.DNSBlockListRecord, error) {
	authToken := auth.GetContextToken(ctx)
	if authToken == "" {
		return nil, gqlerror.Errorf("missing auth token")
	}

	tokenString := strings.TrimPrefix(authToken, "Bearer ")
	valid, err := auth.ValidateJWT(tokenString, r.Config.Auth.Username, r.Config.Auth.Password)

	if !valid || err != nil {
		return nil, gqlerror.Errorf("not authorized")
	}

	if !utils.IsValidIPV4Address(*ip) {
		return nil, gqlerror.Errorf("Invalid IPV4 address: %s", *ip)
	}

	dblRec, err := r.Database.SelectRecord(*ip)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return dblRec, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
