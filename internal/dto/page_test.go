package dto

import (
	"encoding/json"
	"testing"
)

func TestNewPageResp(t *testing.T) {
	list := []string{"admin", "operator"}
	resp := NewPageResp(list, 2)

	if resp.Total != 2 {
		t.Fatalf("Total = %d, want 2", resp.Total)
	}
	if len(resp.List) != 2 || resp.List[0] != "admin" || resp.List[1] != "operator" {
		t.Fatalf("List = %#v, want %#v", resp.List, list)
	}
}

func TestNewPageRespUsesEmptyListForNil(t *testing.T) {
	resp := NewPageResp[string](nil, 0)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal page response: %v", err)
	}
	if got, want := string(data), `{"list":[],"total":0}`; got != want {
		t.Fatalf("JSON = %s, want %s", got, want)
	}
}
