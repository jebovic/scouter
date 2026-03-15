package comment

import "context"

// missionServiceBridge satisfies MissionResolver using a plain function,
// keeping the comment package free of a direct dependency on mission.Service.
type missionServiceBridge struct {
	fn func(ctx context.Context, slug string) (string, bool, error)
}

func (b *missionServiceBridge) GetMissionIDBySlug(ctx context.Context, slug string) (string, bool, error) {
	return b.fn(ctx, slug)
}

// NewBridgeResolver creates a MissionResolver from a plain lookup function.
// Usage in main.go:
//
//	comment.NewBridgeResolver(func(ctx context.Context, slug string) (string, bool, error) {
//	    m, err := missionSvc.GetBySlug(ctx, slug)
//	    if err != nil { return "", false, err }
//	    if m == nil { return "", false, nil }
//	    return m.ID.String(), true, nil
//	})
func NewBridgeResolver(fn func(ctx context.Context, slug string) (string, bool, error)) MissionResolver {
	return &missionServiceBridge{fn: fn}
}
