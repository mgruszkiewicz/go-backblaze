package backblaze

import (
	"testing"
	"time"
)


func TestCreateApplicationKey(T *testing.T) {

	accountID := "test"

	keyName := "my-new-key"
	applicationKeyId := "fffff000123"
	applicationKey := "abcdfxyz"

	timestampNow := time.Now().Unix()
	keyValidDurationInSeconds := 3600
	expirationTimeStamp := timestampNow + int64(keyValidDurationInSeconds)

	client, server := prepareResponses([]response{
		{code: 200, body: authorizeAccountResponse{
			AccountID:          accountID,
			APIEndpoint:        "http://api.url",
			AuthorizationToken: "testToken",
			DownloadURL:        "http://download.url",
		}},
		{code: 200, body: ApplicationKeyResponse{
			KeyName:             keyName,
			ApplicationKeyId:    applicationKeyId,
			ApplicationKey:      applicationKey,
			ExpirationTimestamp: expirationTimeStamp,
		}},
	})
	defer server.Close()

	b2 := &B2{
		Credentials: Credentials{
			AccountID:      accountID,
			ApplicationKey: "test",
		},
		Debug:      testing.Verbose(),
		httpClient: *client,
		host:       server.URL,
	}

	appkey, err := b2.CreateApplicationKey(&CreateKeyRequest{
		Capabilities:           []string{"listAllBucketNames", "listKeys"},
		KeyName:                keyName,
		ValidDurationInSeconds: keyValidDurationInSeconds,
	})

	if err != nil {
		T.Fatal(err)
	}

	if appkey.KeyName != keyName {
		T.Errorf("Expected to get keyName '%s', received '%s'", keyName, appkey.KeyName)
	}

	if appkey.ApplicationKeyId != applicationKeyId {
		T.Errorf("Expected to get applicationKeyId '%s', received '%s'", applicationKeyId, appkey.ApplicationKeyId)
	}

	if appkey.ApplicationKey != applicationKey {
		T.Errorf("Expected to get applicationKey '%s', received '%s'", applicationKey, appkey.ApplicationKey)
	}

	if appkey.ExpirationTimestamp != expirationTimeStamp {
		T.Errorf("Expected to get ExpirationTimestamp '%d', received '%d'", expirationTimeStamp, appkey.ExpirationTimestamp)
	}

}

func TestCreateApplicationKeyNoCapabilities(T *testing.T) {

	accountID := "test"

	keyName := "my-new-key"
	applicationKeyId := "fffff000123"
	applicationKey := "abcdfxyz"

	timestampNow := time.Now().Unix()
	keyValidDurationInSeconds := 3600
	expirationTimeStamp := timestampNow + int64(keyValidDurationInSeconds)

	client, server := prepareResponses([]response{
		{code: 200, body: authorizeAccountResponse{
			AccountID:          accountID,
			APIEndpoint:        "http://api.url",
			AuthorizationToken: "testToken",
			DownloadURL:        "http://download.url",
		}},
		{code: 200, body: ApplicationKeyResponse{
			KeyName:             keyName,
			ApplicationKeyId:    applicationKeyId,
			ApplicationKey:      applicationKey,
			ExpirationTimestamp: expirationTimeStamp,
		}},
	})
	defer server.Close()

	b2 := &B2{
		Credentials: Credentials{
			AccountID:      accountID,
			ApplicationKey: "test",
		},
		Debug:      testing.Verbose(),
		httpClient: *client,
		host:       server.URL,
	}

	_, err := b2.CreateApplicationKey(&CreateKeyRequest{
		Capabilities:           []string{},
		KeyName:                keyName,
		ValidDurationInSeconds: keyValidDurationInSeconds,
	})

	if err == nil {
		T.Fatal(err)
	}
}

func TestListApplicationKeys(T *testing.T) {

	accountID := "test"

	client, server := prepareResponses([]response{
		{code: 200, body: authorizeAccountResponse{
			AccountID:          accountID,
			APIEndpoint:        "http://api.url",
			AuthorizationToken: "testToken",
			DownloadURL:        "http://download.url",
		}},
		{code: 200, body: listKeysResponse{
			Keys: []*ApplicationKeyResponse{
				{
					KeyName:          "key-one",
					ApplicationKeyId: "id1",
					AccountID:        accountID,
					Capabilities:     []string{"listKeys"},
				},
				{
					KeyName:          "key-two",
					ApplicationKeyId: "id2",
					AccountID:        accountID,
					Capabilities:     []string{"listBuckets"},
				},
			},
		}},
	})
	defer server.Close()

	b2 := &B2{
		Credentials: Credentials{
			AccountID:      accountID,
			ApplicationKey: "test",
		},
		Debug:      testing.Verbose(),
		httpClient: *client,
		host:       server.URL,
	}

	keys, err := b2.ListApplicationKeys()
	if err != nil {
		T.Fatal(err)
	}

	if len(keys) != 2 {
		T.Fatalf("Expected 2 keys, received %d", len(keys))
	}
	if keys[0].KeyName != "key-one" || keys[0].ApplicationKeyId != "id1" {
		T.Errorf("Unexpected first key: %+v", keys[0])
	}
	if keys[1].KeyName != "key-two" || keys[1].ApplicationKeyId != "id2" {
		T.Errorf("Unexpected second key: %+v", keys[1])
	}
}

func TestListApplicationKeysPagination(T *testing.T) {

	accountID := "test"

	client, server := prepareResponses([]response{
		{code: 200, body: authorizeAccountResponse{
			AccountID:          accountID,
			APIEndpoint:        "http://api.url",
			AuthorizationToken: "testToken",
			DownloadURL:        "http://download.url",
		}},
		{code: 200, body: listKeysResponse{
			Keys: []*ApplicationKeyResponse{
				{
					KeyName:          "key-one",
					ApplicationKeyId: "id1",
					AccountID:        accountID,
				},
				{
					KeyName:          "key-two",
					ApplicationKeyId: "id2",
					AccountID:        accountID,
				},
			},
			NextApplicationKeyId: "id2",
		}},
		{code: 200, body: listKeysResponse{
			Keys: []*ApplicationKeyResponse{
				{
					KeyName:          "key-three",
					ApplicationKeyId: "id3",
					AccountID:        accountID,
					BucketIds:        []string{"bucket1"},
				},
			},
		}},
	})
	defer server.Close()

	b2 := &B2{
		Credentials: Credentials{
			AccountID:      accountID,
			ApplicationKey: "test",
		},
		Debug:      testing.Verbose(),
		httpClient: *client,
		host:       server.URL,
	}

	keys, err := b2.ListApplicationKeys()
	if err != nil {
		T.Fatal(err)
	}

	if len(keys) != 3 {
		T.Fatalf("Expected 3 keys across pages, received %d", len(keys))
	}

	expected := []string{"key-one", "key-two", "key-three"}
	for i, name := range expected {
		if keys[i].KeyName != name {
			T.Errorf("Expected key %d to be %q, received %q", i, name, keys[i].KeyName)
		}
	}
}

func TestListApplicationKey(T *testing.T) {

	accountID := "test"

	newClient := func(keys []*ApplicationKeyResponse) (*B2, func()) {
		client, server := prepareResponses([]response{
			{code: 200, body: authorizeAccountResponse{
				AccountID:          accountID,
				APIEndpoint:        "http://api.url",
				AuthorizationToken: "testToken",
				DownloadURL:        "http://download.url",
			}},
			{code: 200, body: listKeysResponse{
				Keys: keys,
			}},
		})

		b2 := &B2{
			Credentials: Credentials{
				AccountID:      accountID,
				ApplicationKey: "test",
			},
			Debug:      testing.Verbose(),
			httpClient: *client,
			host:       server.URL,
		}

		return b2, server.Close
	}

	keys := []*ApplicationKeyResponse{
		{
			KeyName:          "key-one",
			ApplicationKeyId: "id1",
			AccountID:        accountID,
		},
		{
			KeyName:          "key-two",
			ApplicationKeyId: "id2",
			AccountID:        accountID,
		},
	}

	T.Run("filter by name returns just the one", func(T *testing.T) {
		b2, close := newClient(keys)
		defer close()

		filtered, err := b2.ListApplicationKey("key-two")
		if err != nil {
			T.Fatal(err)
		}

		if len(filtered) != 1 {
			T.Fatalf("Expected 1 key, received %d", len(filtered))
		}
		if filtered[0].KeyName != "key-two" || filtered[0].ApplicationKeyId != "id2" {
			T.Errorf("Unexpected key returned: %+v", filtered[0])
		}
	})

	T.Run("no filter returns all", func(T *testing.T) {
		b2, close := newClient(keys)
		defer close()

		all, err := b2.ListApplicationKey()
		if err != nil {
			T.Fatal(err)
		}

		if len(all) != 2 {
			T.Errorf("Expected 2 keys, received %d", len(all))
		}
	})

	T.Run("unknown name returns empty", func(T *testing.T) {
		b2, close := newClient(keys)
		defer close()

		filtered, err := b2.ListApplicationKey("no-such-key")
		if err != nil {
			T.Fatal(err)
		}

		if len(filtered) != 0 {
			T.Errorf("Expected 0 keys, received %d", len(filtered))
		}
	})
}
