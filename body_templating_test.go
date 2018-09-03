package atticus_test

import (
	"encoding/json"
	"testing"

	"github.com/NearlyUnique/atticus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_string_templates(t *testing.T) {
	type blob map[string]interface{}
	c := map[string]interface{}{
		"key": "start {{.Header.id}} end",
	}
	data := atticus.TemplateData{
		Header: blob{
			"id": "from-header-id",
		},
	}
	var actual struct{ Key string }

	jsonBody, err := atticus.ApplyTemplate(c, &data)

	assert.NoError(t, err)

	err = json.Unmarshal(jsonBody, &actual)
	require.NoError(t, err)

	assert.Equal(t, "start from-header-id end", actual.Key)
}

func Test_simple_json_structure_can_be_rendered(t *testing.T) {
	c := createRequestBody(t, `{
		"a-string":"a-value",
		"a-number":12,
		"a-boolean":true,
		"an-object":{"sub-key":"ok"},
		"an-array":[10,11,12]
		}`)
	data := atticus.TemplateData{}
	jsonBody, err := atticus.ApplyTemplate(c, &data)

	assert.NoError(t, err)

	var actual struct {
		AString  string  `json:"a-string"`
		ANumber  float64 `json:"a-number"`
		ABoolean bool    `json:"a-boolean"`
		AnObject struct {
			SubKey string `json:"sub-key"`
		} `json:"an-object"`
		AnArray []int `json:"an-array"`
	}
	err = json.Unmarshal(jsonBody, &actual)
	require.NoError(t, err)

	assert.Equal(t, "a-value", actual.AString)
	assert.Equal(t, 12.0, actual.ANumber)
	assert.Equal(t, true, actual.ABoolean)

	assert.Equal(t, "ok", actual.AnObject.SubKey)

	if assert.Equal(t, 3, len(actual.AnArray)) {
		assert.Equal(t, 10, actual.AnArray[0])
		assert.Equal(t, 11, actual.AnArray[1])
		assert.Equal(t, 12, actual.AnArray[2])
	}
}

func createRequestBody(t *testing.T, bodyText string) interface{} {
	var resp map[string]interface{}
	err := json.Unmarshal([]byte(bodyText), &resp)
	require.NoError(t, err)
	return resp
}
