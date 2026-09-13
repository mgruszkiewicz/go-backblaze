package backblaze


// CreateApplicationKey creates application key
// It is possible to limit application key capabilities by defining an
// array with allowed permission
// (which can be find here https://www.backblaze.com/apidocs/b2-create-key)
// This function to work requires master application key to
// be used for authenticaion
func (b *B2) CreateApplicationKey(keyDetails *CreateKeyRequest) (*ApplicationKeyResponse, error) {
	request := &CreateKeyRequest{
		AccountID:              b.AccountID,
		Capabilities:           keyDetails.Capabilities,
		KeyName:                keyDetails.KeyName,
		ValidDurationInSeconds: keyDetails.ValidDurationInSeconds,
		BucketId:               keyDetails.BucketId,
		NamePrefix:             keyDetails.NamePrefix,
	}

	response := &ApplicationKeyResponse{}

	if err := b.apiRequest("b2_create_key", request, response); err != nil {
		return nil, err
	}

	return response, nil
}

// ListApplicationKeys lists all application keys associated with the account.
//
// It follows pagination via nextApplicationKeyId until all keys have been
// retrieved. This function requires the listKeys capability.
// Note that the returned keys do not include the secret applicationKey
// string; the API only returns key metadata when listing.
func (b *B2) ListApplicationKeys() ([]*ApplicationKeyResponse, error) {
	var keys []*ApplicationKeyResponse

	startApplicationKeyId := ""
	for {
		request := &listKeysRequest{
			AccountID:             b.AccountID,
			MaxKeyCount:           10000,
			StartApplicationKeyId: startApplicationKeyId,
		}
		response := &listKeysResponse{}

		if err := b.apiRequest("b2_list_keys", request, response); err != nil {
			return nil, err
		}

		keys = append(keys, response.Keys...)

		if response.NextApplicationKeyId == "" {
			break
		}
		startApplicationKeyId = response.NextApplicationKeyId
	}

	return keys, nil
}

// ListApplicationKey returns all application keys, optionally filtered by
// key name. When called without arguments it behaves like ListApplicationKeys
// and returns every key. When called with a key name, only keys with a
// matching name are returned (key names may not be unique, so a slice is
// still returned -- typically it holds just the one key).
//
// Note: the B2 API has no server-side name filter for b2_list_keys, so the
// filtering is done client-side after listing.
func (b *B2) ListApplicationKey(keyName ...string) ([]*ApplicationKeyResponse, error) {
	keys, err := b.ListApplicationKeys()
	if err != nil {
		return nil, err
	}

	if len(keyName) == 0 || keyName[0] == "" {
		return keys, nil
	}

	filtered := make([]*ApplicationKeyResponse, 0, len(keys))
	for _, key := range keys {
		if key.KeyName == keyName[0] {
			filtered = append(filtered, key)
		}
	}

	return filtered, nil
}

// DeleteApplicationKey deletes application keys by providing applicationKeyID
// This function to work requires master application key to
// be used for authenticaion
func (b *B2) DeleteApplicationKey(applicationKeyId string) (*ApplicationKeyResponse, error) {
	request := &DeleteKeyRequest{
		ApplicationKeyId: applicationKeyId,
	}

	response := &ApplicationKeyResponse{}

	if err := b.apiRequest("b2_delete_key", request, response); err != nil {
		return nil, err
	}

	return response, nil
}
