package main

import (
    "context"
    "fmt"

    recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
    recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

// CreateAssessment creates an assessment to analyze the risk of a UI action.
//
// projectID: Your Google Cloud Project ID.
// recaptchaKey: The reCAPTCHA key associated with the site/app
// token: The generated token obtained from the client.
// recaptchaAction: Action name corresponding to the token.
func CreateAssessment(projectID string, recaptchaKey string, token string, recaptchaAction string) error {
    // Create the reCAPTCHA client.
    ctx := context.Background()
    client, err := recaptcha.NewClient(ctx)
    if err != nil {
        return fmt.Errorf("error creating reCAPTCHA client: %w", err)
    }
    defer client.Close()

    // Set the properties of the event to be tracked.
    event := &recaptchapb.Event{
        Token:   token,
        SiteKey: recaptchaKey,
    }

    assessment := &recaptchapb.Assessment{
        Event: event,
    }

    // Build the assessment request.
    request := &recaptchapb.CreateAssessmentRequest{
        Assessment: assessment,
        Parent:     fmt.Sprintf("projects/%s", projectID),
    }

    response, err := client.CreateAssessment(ctx, request)
    if err != nil {
        return fmt.Errorf("error calling CreateAssessment: %w", err)
    }

    // Check if the token is valid.
    if !response.TokenProperties.Valid {
        return fmt.Errorf("token was invalid for reasons: %v", response.TokenProperties.InvalidReason)
    }

    // Check if the expected action was executed.
    if response.TokenProperties.Action != recaptchaAction {
        return fmt.Errorf("action mismatch: expected %s", recaptchaAction)
    }

    // Log the results
    fmt.Printf("The reCAPTCHA score for this token is: %v\n", response.RiskAnalysis.Score)
    for _, reason := range response.RiskAnalysis.Reasons {
        fmt.Printf("%s\n", reason.String())
    }
    if response.RiskAnalysis.Score < 0.5 {
        return fmt.Errorf("Bad score error")
    }

    return nil
}
