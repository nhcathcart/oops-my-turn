package server

import "context"

type HelloOutput struct {
	Body helloResponseBody
}

type helloResponseBody struct {
	Message string `json:"message" doc:"Public oops-my-turn message"`
}

func (s *Server) handleHello(ctx context.Context, input *struct{}) (*HelloOutput, error) {
	return &HelloOutput{
		Body: helloResponseBody{
			Message: "Hello from oops-my-turn.",
		},
	}, nil
}
