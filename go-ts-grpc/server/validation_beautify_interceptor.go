package main

import (
	"context"
	"fmt"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BeautifyValidationInterceptor wraps the protovalidate middleware in the chain.
// On InvalidArgument errors that carry a *validate.Violations detail, it repackages
// them into the standard google.rpc.BadRequest shape so clients see field-level
// violations via the well-known errdetails contract instead of the validator's
// raw Violations proto.
func BeautifyValidationInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		return resp, err
	}

	for _, d := range st.Details() {
		violations, ok := d.(*validate.Violations)
		if !ok {
			continue
		}
		br := &errdetails.BadRequest{}
		fmt.Printf("[validation] %s — %d violation(s):\n", info.FullMethod, len(violations.GetViolations()))
		for _, v := range violations.GetViolations() {
			field := v.GetField().String()
			ruleID := v.GetRuleId()
			msg := v.GetMessage()
			fmt.Printf("  • field=%s rule=%s msg=%q\n", field, ruleID, msg)
			br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: msg,
			})
		}
		newSt, wErr := status.New(codes.InvalidArgument, "validation failed").WithDetails(br)
		if wErr != nil {
			return resp, err
		}
		return resp, newSt.Err()
	}
	return resp, err
}
