package squadcast

import "context"

// V3 API doesn't allow create incident
// Use V2 Incident API https://support.squadcast.com/docs/apiv2
func (c *WebhookClient) CreateIncident(ctx context.Context, request *PostIncidentRequest) error {
	return c.PostIncident(ctx, request)
}
