package auth

import "testing"

func FuzzExtractOAuth2ClientCreds(f *testing.F) {
	// Seed corpus from existing test cases
	f.Add([]byte(`{"installed":{"client_id":"id123","client_secret":"sec456"}}`))
	f.Add([]byte(`{"web":{"client_id":"webid","client_secret":"websec"}}`))
	f.Add([]byte(`{"other":{}}`))
	f.Add([]byte(`{"installed":{"client_id":"","client_secret":"sec"}}`))
	f.Add([]byte(`not json`))
	f.Add([]byte(`{"installed":{"client_id":"id","client_secret":""}}`))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should never panic regardless of input
		extractOAuth2ClientCreds(data)
	})
}
